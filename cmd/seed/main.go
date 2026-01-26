package main

import (
	"context"
	"flag"
	"log"
	"time"

	"github.com/joho/godotenv"
	"github.com/redis/rueidis"

	"github.com/aditnikel/grapgraph/src/config"
	"github.com/aditnikel/grapgraph/src/graph"
	"github.com/aditnikel/grapgraph/src/seed"
)

func main() {
	_ = godotenv.Load()

	reset := flag.Bool("reset", false, "delete the graph before seeding (dev only)")
	flag.Parse()

	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	rdb, err := rueidis.NewClient(rueidis.ClientOption{
		InitAddress: cfg.RedisAddrs,
		Password:    cfg.RedisPassword,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer rdb.Close()

	repo := graph.New(rdb, cfg.GraphName, cfg.DBTimeout)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if *reset {
		_ = repo.DeleteGraph(ctx) // best-effort
	}

	repo.EnsureSchema(ctx)

	if err := seed.SeedDemo(ctx, repo); err != nil {
		log.Fatal(err)
	}

	log.Println("seed completed")
}
