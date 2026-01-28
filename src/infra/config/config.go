package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	HTTPAddr string

	RedisAddrs    []string
	RedisPassword string
	GraphName     string

	DBTimeout time.Duration
	LogLevel  string

	// Default graph query behavior (used when request omits/zero)
	DefaultMaxNodes      int
	DefaultMaxEdges      int
	DefaultMinEventCount int
	DefaultRankBy        string
}

func Load() (Config, error) {
	c := Config{
		HTTPAddr:      envStr("HTTP_ADDR", ":8080"),
		RedisPassword: os.Getenv("REDIS_PASSWORD"),
		GraphName:     envStr("GRAPH_NAME", "fraudnet"),
		LogLevel:      envStr("LOG_LEVEL", "info"),
	}

	addrs := envStr("REDIS_ADDRS", "localhost:6379")
	c.RedisAddrs = splitCSV(addrs)

	toMS := envInt("DB_TIMEOUT_MS", 1500)
	c.DBTimeout = time.Duration(toMS) * time.Millisecond

	c.DefaultMaxNodes = envInt("DEFAULT_MAX_NODES", 200)
	c.DefaultMaxEdges = envInt("DEFAULT_MAX_EDGES", 400)
	c.DefaultMinEventCount = envInt("DEFAULT_MIN_EVENT_COUNT", 1)
	c.DefaultRankBy = envStr("DEFAULT_RANK_BY", "event_count_30d")

	if len(c.RedisAddrs) == 0 {
		return Config{}, fmt.Errorf("REDIS_ADDRS must not be empty")
	}
	if c.GraphName == "" {
		return Config{}, fmt.Errorf("GRAPH_NAME must not be empty")
	}

	if c.DefaultMaxNodes <= 0 {
		c.DefaultMaxNodes = 200
	}
	if c.DefaultMaxEdges <= 0 {
		c.DefaultMaxEdges = 400
	}
	if c.DefaultMinEventCount <= 0 {
		c.DefaultMinEventCount = 1
	}
	switch c.DefaultRankBy {
	case "event_count_30d", "event_count", "total_amount":
	default:
		c.DefaultRankBy = "event_count_30d"
	}

	return c, nil
}

func splitCSV(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func envStr(k, d string) string {
	if v := strings.TrimSpace(os.Getenv(k)); v != "" {
		return v
	}
	return d
}

func envInt(k string, d int) int {
	v := strings.TrimSpace(os.Getenv(k))
	if v == "" {
		return d
	}
	i, err := strconv.Atoi(v)
	if err != nil {
		return d
	}
	return i
}
