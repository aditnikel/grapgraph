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

	"github.com/aditnikel/grapgraph/gen/ingest"
	ingestclient "github.com/aditnikel/grapgraph/gen/http/ingest/client"
)

var (
	targetURL   = flag.String("url", "http://localhost:8080/v1/ingest/event", "Target URL")
	numWorkers  = flag.Int("workers", 10, "Number of concurrent workers")
	duration    = flag.Duration("duration", 10*time.Second, "Duration of the load test")
	timeout     = flag.Duration("timeout", 5*time.Second, "HTTP client timeout")
	minEvents   = flag.Int("min-events", 1, "Minimum events per request")
	maxEvents   = flag.Int("max-events", 5, "Maximum events per request")
	pace        = flag.Duration("pace", 0, "Optional sleep between requests per worker")
	eventTypes  = flag.String("event-types", "LOGOUT,WITHDRAWAL,KYC_UPDATE,PROFILE_UPDATE,PAYMENT,ACCOUNT_UPDATE,CUSTOMER_EVENT,TRANSACTION,KYC,REGISTER,PASSWORD_CHANGE,LOGIN,MANUAL", "Comma-separated event types")
	errorLog    = flag.String("error-log", "", "Write request errors to this file (append). Empty disables.")
	totalReqs   atomic.Uint64
	successReqs atomic.Uint64
	failedReqs  atomic.Uint64
)

var errLogger *log.Logger
var parsedEventTypes []string

func main() {
	flag.Parse()

	if *minEvents < 1 || *maxEvents < 1 || *minEvents > *maxEvents {
		log.Fatalf("Invalid event bounds: min-events=%d max-events=%d", *minEvents, *maxEvents)
	}

	parsedEventTypes = parseEventTypes(*eventTypes)
	if len(parsedEventTypes) == 0 {
		log.Fatalf("Invalid event-types: %q", *eventTypes)
	}

	if *errorLog != "" {
		f, err := os.OpenFile(*errorLog, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
		if err != nil {
			log.Fatalf("Failed to open error log file: %v", err)
		}
		defer f.Close()
		errLogger = log.New(f, "", log.LstdFlags)
	}

	log.Printf("Starting load test with %d workers for %v...", *numWorkers, *duration)

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
	} else {
		failedReqs.Add(1)
		body, _ := io.ReadAll(resp.Body)
		if len(body) > 0 {
			logErrorf("server error: %s body=%s", resp.Status, bytes.TrimSpace(body))
		} else {
			logErrorf("server error: %s", resp.Status)
		}
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
