package maintenance

import (
	"strconv"

	"github.com/aditnikel/grapgraph/src/graph/cypher"
	infra "github.com/aditnikel/grapgraph/src/infra/redis"
)

type Compactor struct {
	Redis *infra.Client
}

func (c *Compactor) Prune(cutoff int64) {
	c.Redis.RDB.Do(
		c.Redis.Ctx,
		c.Redis.RDB.B().
			Arbitrary("GRAPH.QUERY").
			Args(
				infra.GraphKey(c.Redis.Ledger),
				cypher.PruneOldEdges,
				"--compact",
				"PARAMS",
				"cutoff", strconv.FormatInt(cutoff, 10),
			).
			Build(),
	)
}
