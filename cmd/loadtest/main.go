package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/aditnikel/grapgraph/gen/graph"
	graphclient "github.com/aditnikel/grapgraph/gen/http/graph/client"
	ingestclient "github.com/aditnikel/grapgraph/gen/http/ingest/client"
	"github.com/aditnikel/grapgraph/gen/ingest"
)

var (
	targetURL         = flag.String("url", "http://localhost:8080/v1/ingest/event", "Target URL")
	scenario          = flag.String("scenario", "ingest", "Load test scenario: ingest|graph-subgraph|graph-metadata")
	numWorkers        = flag.Int("workers", 10, "Number of concurrent workers")
	duration          = flag.Duration("duration", 10*time.Second, "Duration of the load test")
	timeout           = flag.Duration("timeout", 5*time.Second, "HTTP client timeout")
	minEvents         = flag.Int("min-events", 1, "Minimum events per request")
	maxEvents         = flag.Int("max-events", 5, "Maximum events per request")
	pace              = flag.Duration("pace", 0, "Optional sleep between requests per worker")
	eventTypes        = flag.String("event-types", "LOGOUT,WITHDRAWAL,KYC_UPDATE,PROFILE_UPDATE,PAYMENT,ACCOUNT_UPDATE,CUSTOMER_EVENT,TRANSACTION,KYC,REGISTER,PASSWORD_CHANGE,LOGIN,MANUAL", "Comma-separated event types")
	graphRootType     = flag.String("graph-root-type", "USER", "Root node type for graph subgraph requests")
	graphRootPrefix   = flag.String("graph-root-prefix", "user_", "Root node key prefix for graph subgraph requests")
	graphRootMax      = flag.Int("graph-root-max", 1000, "Max integer suffix (exclusive) for graph root keys")
	graphHops         = flag.Int("graph-hops", 5, "Graph subgraph hops (>=1)")
	graphMaxNodes     = flag.Int("graph-max-nodes", 100, "Graph subgraph max nodes")
	graphMaxEdges     = flag.Int("graph-max-edges", 200, "Graph subgraph max edges")
	graphMinEvent     = flag.Int("graph-min-event-count", 0, "Graph subgraph minimum event count per edge")
	graphTimeWindowMs = flag.Int64("graph-time-window-ms", 0, "Graph subgraph time window in ms (0 = all)")
	graphEdgeTypes    = flag.String("graph-edge-types", "", "Comma-separated edge types for graph subgraph (empty = all)")
	errorLog          = flag.String("error-log", "", "Write request errors to this file (append). Empty disables.")
	totalReqs         atomic.Uint64
	successReqs       atomic.Uint64
	failedReqs        atomic.Uint64
)

var errLogger *log.Logger
var parsedEventTypes []string
var parsedGraphEdgeTypes []string
var scenarioMode string

func main() {
	flag.Parse()

	scenarioMode = strings.ToLower(strings.TrimSpace(*scenario))
	switch scenarioMode {
	case "ingest":
		if *minEvents < 1 || *maxEvents < 1 || *minEvents > *maxEvents {
			log.Fatalf("Invalid event bounds: min-events=%d max-events=%d", *minEvents, *maxEvents)
		}

		parsedEventTypes = parseEventTypes(*eventTypes)
		if len(parsedEventTypes) == 0 {
			log.Fatalf("Invalid event-types: %q", *eventTypes)
		}
	case "graph-subgraph":
		if *graphRootType == "" {
			log.Fatalf("Invalid graph-root-type: %q", *graphRootType)
		}
		if *graphRootMax < 1 {
			log.Fatalf("Invalid graph-root-max: %d", *graphRootMax)
		}
		if *graphHops < 1 {
			log.Fatalf("Invalid graph-hops: %d", *graphHops)
		}
		if *graphMaxNodes < 1 || *graphMaxEdges < 1 {
			log.Fatalf("Invalid graph limits: graph-max-nodes=%d graph-max-edges=%d", *graphMaxNodes, *graphMaxEdges)
		}
		if *graphMinEvent < 0 {
			log.Fatalf("Invalid graph-min-event-count: %d", *graphMinEvent)
		}
		if *graphTimeWindowMs < 0 {
			log.Fatalf("Invalid graph-time-window-ms: %d", *graphTimeWindowMs)
		}
		parsedGraphEdgeTypes = parseEventTypes(*graphEdgeTypes)
	case "graph-metadata":
		// No additional validation.
	default:
		log.Fatalf("Unknown scenario: %q (expected ingest|graph-subgraph|graph-metadata)", *scenario)
	}

	if *errorLog != "" {
		f, err := os.OpenFile(*errorLog, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
		if err != nil {
			log.Fatalf("Failed to open error log file: %v", err)
		}
		defer f.Close()
		errLogger = log.New(f, "", log.LstdFlags)
	}

	log.Printf("Starting %s load test with %d workers for %v (url=%s)...", scenarioMode, *numWorkers, *duration, *targetURL)

	var wg sync.WaitGroup
	start := time.Now()
	done := make(chan struct{})

	// Timer to stop the test
	go func() {
		time.Sleep(*duration)
		close(done)
	}()

	for i := 0; i < *numWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			client := &http.Client{Timeout: *timeout}
			rng := rand.New(rand.NewSource(time.Now().UnixNano() + int64(workerID)))
			for {
				select {
				case <-done:
					return
				default:
					sendRequest(client, rng)
					if *pace > 0 {
						time.Sleep(*pace)
					}
				}
			}
		}(i)
	}

	wg.Wait()
	elapsed := time.Since(start)

	printReport(elapsed)
}

func sendRequest(client *http.Client, rng *rand.Rand) {
	switch scenarioMode {
	case "ingest":
		sendIngestRequest(client, rng)
	case "graph-subgraph":
		sendGraphSubgraphRequest(client, rng)
	case "graph-metadata":
		sendGraphMetadataRequest(client)
	default:
		log.Printf("Unknown scenario %q", scenarioMode)
	}
}

func sendIngestRequest(client *http.Client, rng *rand.Rand) {
	payload := generatePayload(rng)
	reqBody := ingestclient.NewPostEventRequestBody(payload)
	data, err := json.Marshal(reqBody)
	if err != nil {
		log.Printf("Failed to marshal payload: %v", err)
		logErrorf("marshal error: %v", err)
		return
	}

	resp, err := client.Post(*targetURL, "application/json", bytes.NewReader(data))
	totalReqs.Add(1)
	if err != nil {
		failedReqs.Add(1)
		log.Printf("Request failed: %v", err)
		logErrorf("request failed: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		successReqs.Add(1)
		return
	}

	failedReqs.Add(1)
	body, _ := io.ReadAll(resp.Body)
	if len(body) > 0 {
		logErrorf("server error: %s body=%s", resp.Status, bytes.TrimSpace(body))
	} else {
		logErrorf("server error: %s", resp.Status)
	}
}

func sendGraphSubgraphRequest(client *http.Client, rng *rand.Rand) {
	payload := generateSubgraphPayload(rng)
	reqBody := graphclient.NewPostSubgraphRequestBody(payload)
	data, err := json.Marshal(reqBody)
	if err != nil {
		log.Printf("Failed to marshal graph payload: %v", err)
		logErrorf("marshal error: %v", err)
		return
	}

	req, err := http.NewRequest(http.MethodPost, *targetURL, bytes.NewReader(data))
	if err != nil {
		log.Printf("Failed to build graph request: %v", err)
		logErrorf("request build failed: %v", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	totalReqs.Add(1)
	if err != nil {
		failedReqs.Add(1)
		log.Printf("Request failed: %v", err)
		logErrorf("request failed: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		successReqs.Add(1)
		return
	}

	failedReqs.Add(1)
	body, _ := io.ReadAll(resp.Body)
	if len(body) > 0 {
		logErrorf("server error: %s body=%s", resp.Status, bytes.TrimSpace(body))
	} else {
		logErrorf("server error: %s", resp.Status)
	}
}

func sendGraphMetadataRequest(client *http.Client) {
	req, err := http.NewRequest(http.MethodGet, *targetURL, nil)
	if err != nil {
		log.Printf("Failed to build graph metadata request: %v", err)
		logErrorf("request build failed: %v", err)
		return
	}

	resp, err := client.Do(req)
	totalReqs.Add(1)
	if err != nil {
		failedReqs.Add(1)
		log.Printf("Request failed: %v", err)
		logErrorf("request failed: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		successReqs.Add(1)
		return
	}

	failedReqs.Add(1)
	body, _ := io.ReadAll(resp.Body)
	if len(body) > 0 {
		logErrorf("server error: %s body=%s", resp.Status, bytes.TrimSpace(body))
	} else {
		logErrorf("server error: %s", resp.Status)
	}
}

func generatePayload(rng *rand.Rand) *ingest.BulkCustomerEvents {
	// Generate N events per batch
	limit := rng.Intn(*maxEvents-*minEvents+1) + *minEvents
	events := make([]*ingest.CustomerEvent, limit)

	amountEventTypes := map[string]struct{}{
		"PAYMENT":     {},
		"WITHDRAWAL":  {},
		"TRANSACTION": {},
	}

	for i := 0; i < limit; i++ {
		uid := fmt.Sprintf("user_%d", rng.Intn(1000))
		etype := parsedEventTypes[rng.Intn(len(parsedEventTypes))]
		mid := fmt.Sprintf("merch_%d", rng.Intn(100))
		ts := time.Now().Format(time.RFC3339)

		var amt *float64
		if _, ok := amountEventTypes[etype]; ok {
			v := rng.Float64() * 1000
			amt = &v
		}

		events[i] = &ingest.CustomerEvent{
			UserID:                 uid,
			EventType:              etype,
			EventTimestamp:         ts,
			TotalTransactionAmount: amt,
			MerchantIDMpan:         &mid,
		}
	}

	return &ingest.BulkCustomerEvents{
		Events: events,
	}
}

func generateSubgraphPayload(rng *rand.Rand) *graph.SubgraphRequest {
	key := fmt.Sprintf("%s%d", *graphRootPrefix, rng.Intn(*graphRootMax))
	req := &graph.SubgraphRequest{
		Root: &struct {
			Type string
			Key  string
		}{
			Type: *graphRootType,
			Key:  key,
		},
		Hops:          *graphHops,
		MinEventCount: *graphMinEvent,
		TimeWindowMs:  *graphTimeWindowMs,
		Limit: &struct {
			MaxNodes int
			MaxEdges int
		}{
			MaxNodes: *graphMaxNodes,
			MaxEdges: *graphMaxEdges,
		},
	}

	if len(parsedGraphEdgeTypes) > 0 {
		req.EdgeTypes = append([]string(nil), parsedGraphEdgeTypes...)
	}

	return req
}

func printReport(elapsed time.Duration) {
	total := totalReqs.Load()
	success := successReqs.Load()
	failed := failedReqs.Load()
	rps := float64(total) / elapsed.Seconds()

	fmt.Println("\n--- Load Test Results ---")
	fmt.Printf("Duration: %v\n", elapsed)
	fmt.Printf("Total Requests: %d\n", total)
	fmt.Printf("Success: %d\n", success)
	fmt.Printf("Failed: %d\n", failed)
	fmt.Printf("RPS: %.2f\n", rps)
}

func logErrorf(format string, args ...any) {
	if errLogger != nil {
		errLogger.Printf(format, args...)
	}
}

func parseEventTypes(v string) []string {
	parts := strings.Split(v, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
