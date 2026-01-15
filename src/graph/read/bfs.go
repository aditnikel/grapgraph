package read

import (
	"strconv"

	"github.com/aditnikel/grapgraph/src/config"
	cypher "github.com/aditnikel/grapgraph/src/graph/cypher"
	infra "github.com/aditnikel/grapgraph/src/infra/redis"
)

type BFS struct {
	Redis *infra.Client
	Cfg   config.Config
}

func (b *BFS) Traverse(start string, ts int64) error {
	if !b.Cfg.EnableBFS {
		return nil
	}

	resp := b.Redis.RDB.Do(
		b.Redis.Ctx,
		b.Redis.RDB.B().
			Arbitrary("GRAPH.QUERY").
			Args(
				infra.GraphKey(b.Redis.Ledger),
				cypher.WindowedBFS,
				"--compact",
				"PARAMS",
				"start", start,
				"startTs", strconv.FormatInt(ts-b.Cfg.BFSTimeWindowSec, 10),
				"endTs", strconv.FormatInt(ts+b.Cfg.BFSTimeWindowSec, 10),
				"limit", strconv.FormatInt(int64(b.Cfg.BFSResultLimit), 10),
			).
			Build(),
	)

	return resp.Error()
}
