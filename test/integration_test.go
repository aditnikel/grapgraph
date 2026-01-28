package test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/redis/rueidis"

	"github.com/aditnikel/grapgraph/src/graph"
	"github.com/aditnikel/grapgraph/src/observability"
)

func TestHealthPing(t *testing.T) {
	addrs := os.Getenv("REDIS_ADDRS")
	graphName := os.Getenv("GRAPH_NAME")
	if addrs == "" || graphName == "" {
		t.Skip("set REDIS_ADDRS and GRAPH_NAME to run integration test")
	}

	rdb, err := rueidis.NewClient(rueidis.ClientOption{InitAddress: []string{addrs}})
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	defer rdb.Close()

	repo := graph.New(rdb, graphName, 1500*time.Millisecond, observability.New("error"))
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := repo.Ping(ctx); err != nil {
		t.Fatalf("ping failed: %v", err)
	}
}
