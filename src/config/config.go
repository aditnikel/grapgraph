package config

import (
	"os"
	"strconv"
)

type Config struct {
	RedisAddr string
	LedgerID  string
	Goal      string

	EnableStreams bool
	EnableBFS     bool

	BFSMaxDepth      int
	BFSTimeWindowSec int64
	BFSResultLimit   int
	BFSMaxFanout     int
	BFSMaxVisits     int

	ConsumerGroup string
	ConsumerName  string
}

func Load() Config {
	return Config{
		RedisAddr: os.Getenv("REDIS_ADDR"),
		LedgerID:  os.Getenv("LEDGER_ID"),
		Goal:      envStr("GOAL", "fraud"),

		EnableStreams: envBool("ENABLE_STREAMS", true),
		EnableBFS:     envBool("ENABLE_BFS", true),

		BFSMaxDepth:      envInt("BFS_MAX_DEPTH", 3),
		BFSTimeWindowSec: envInt64("BFS_TIME_WINDOW_SEC", 3600),
		BFSResultLimit:   envInt("BFS_RESULT_LIMIT", 10),
		BFSMaxFanout:     envInt("BFS_MAX_FANOUT", 50),
		BFSMaxVisits:     envInt("BFS_MAX_VISITS", 1000),

		ConsumerGroup: envStr("CONSUMER_GROUP", "bfs"),
		ConsumerName:  envStr("CONSUMER_NAME", "worker-1"),
	}
}

func envStr(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
func envBool(k string, d bool) bool {
	if v := os.Getenv(k); v != "" {
		b, _ := strconv.ParseBool(v)
		return b
	}
	return d
}
func envInt(k string, d int) int {
	if v := os.Getenv(k); v != "" {
		i, _ := strconv.Atoi(v)
		return i
	}
	return d
}
func envInt64(k string, d int64) int64 {
	if v := os.Getenv(k); v != "" {
		i, _ := strconv.ParseInt(v, 10, 64)
		return i
	}
	return d
}
