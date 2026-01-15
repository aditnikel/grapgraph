package main

import (
	"log"
	"time"

	"github.com/joho/godotenv"

	"github.com/aditnikel/grapgraph/src/config"
	"github.com/aditnikel/grapgraph/src/domain"
	"github.com/aditnikel/grapgraph/src/graph/read"
	"github.com/aditnikel/grapgraph/src/graph/write"
	infra "github.com/aditnikel/grapgraph/src/infra/redis"
	"github.com/aditnikel/grapgraph/src/streams"
)

func main() {
	godotenv.Load()
	cfg := config.Load()

	redis, err := infra.New(cfg.RedisAddr, cfg.LedgerID)
	if err != nil {
		log.Fatal(err)
	}

	writer := &write.TransferWriter{Redis: redis}
	bfs := &read.BFS{Redis: redis, Cfg: cfg}

	streams.Start(redis, bfs, cfg)

	writer.Write(domain.Transfer{
		Tx:     "tx1",
		From:   "a1",
		To:     "a2",
		Amount: 5000,
		Ts:     time.Now().Unix(),
	})

	select {}
}
