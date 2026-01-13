package graph

import (
	"context"
	"os"
	"strconv"

	"github.com/redis/go-redis/v9"
)

type Client struct {
	rdb   *redis.Client
	ctx   context.Context
	graph string
}

func NewFromEnv() (*Client, error) {
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		addr = "localhost:6379"
	}

	password := os.Getenv("REDIS_PASSWORD")

	db := 0
	if v := os.Getenv("REDIS_DB"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil {
			db = parsed
		}
	}

	graph := os.Getenv("REDIS_GRAPH")
	if graph == "" {
		graph = "money"
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	return &Client{
		rdb:   rdb,
		ctx:   context.Background(),
		graph: graph,
	}, nil
}

func (c *Client) Query(query string) (any, error) {
	return c.rdb.Do(
		c.ctx,
		"GRAPH.QUERY",
		c.graph,
		query,
		"--compact",
	).Result()
}
