package write

import (
	"strconv"

	"github.com/aditnikel/grapgraph/src/domain"
	"github.com/aditnikel/grapgraph/src/graph/cypher"
	infra "github.com/aditnikel/grapgraph/src/infra/redis"
)

type TransferWriter struct {
	Redis *infra.Client
}

func (w *TransferWriter) Write(t domain.Transfer) error {
	resp := w.Redis.RDB.Do(
		w.Redis.Ctx,
		w.Redis.RDB.B().
			Arbitrary("GRAPH.QUERY").
			Args(
				infra.GraphKey(w.Redis.Ledger),
				cypher.AddTransfer,
				"--compact",
				"PARAMS",
				"tx", t.Tx,
				"from", t.From,
				"to", t.To,
				"amount", strconv.FormatFloat(t.Amount, 'f', -1, 64),
				"ts", strconv.FormatInt(t.Ts, 10),
			).
			Build(),
	)
	if resp.Error() != nil {
		return resp.Error()
	}

	w.Redis.RDB.Do(
		w.Redis.Ctx,
		w.Redis.RDB.B().
			Xadd().
			Key(infra.StreamKey(w.Redis.Ledger)).
			Id("*").
			FieldValue().
			FieldValue("from", t.From).
			FieldValue("ts", strconv.FormatInt(t.Ts, 10)).
			Build(),
	)

	return nil
}
