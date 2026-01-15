package streams

import (
	"strconv"

	"github.com/aditnikel/grapgraph/src/config"
	"github.com/aditnikel/grapgraph/src/graph/read"
	infra "github.com/aditnikel/grapgraph/src/infra/redis"
)

func Start(redis *infra.Client, bfs *read.BFS, cfg config.Config) {
	// Create consumer group (ignore error if exists)
	redis.RDB.Do(
		redis.Ctx,
		redis.RDB.B().
			XgroupCreate().
			Key(infra.StreamKey(redis.Ledger)).
			Group(cfg.ConsumerGroup).
			Id("$").
			Mkstream().
			Build(),
	)

	go func() {
		for {
			res, err := redis.RDB.Do(
				redis.Ctx,
				redis.RDB.B().
					Xreadgroup().
					Group(cfg.ConsumerGroup, cfg.ConsumerName).
					Block(5000).
					Streams().
					Key(infra.StreamKey(redis.Ledger)).
					Id(">").
					Build(),
			).AsXRead()

			if err != nil || len(res) == 0 {
				continue
			}

			for _, stream := range res {
				for _, msg := range stream {
					ts, _ := strconv.ParseInt(msg.FieldValues["ts"], 10, 64)
					bfs.Traverse(msg.FieldValues["from"], ts)

					redis.RDB.Do(
						redis.Ctx,
						redis.RDB.B().
							Xack().
							Key(infra.StreamKey(redis.Ledger)).
							Group(cfg.ConsumerGroup).
							Id(msg.ID).
							Build(),
					)
				}
			}
		}
	}()
}
