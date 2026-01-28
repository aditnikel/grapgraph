package graph

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/redis/rueidis"

	"github.com/aditnikel/grapgraph/src/graph/cypher"
	"github.com/aditnikel/grapgraph/src/model"
	"github.com/aditnikel/grapgraph/src/observability"
)

type Repo struct {
	rdb       rueidis.Client
	graphName string
	timeout   time.Duration
	log       *observability.Logger
}

func New(rdb rueidis.Client, graphName string, timeout time.Duration, log *observability.Logger) *Repo {
	return &Repo{rdb: rdb, graphName: graphName, timeout: timeout, log: log}
}

func (g *Repo) interpolate(query string, params map[string]any) string {
	for k, v := range params {
		placeholder := "$" + k
		var val string
		switch x := v.(type) {
		case string:
			val = fmt.Sprintf("'%s'", strings.ReplaceAll(x, "'", "\\'"))
		case int, int64, float64:
			val = fmt.Sprintf("%v", x)
		default:
			val = fmt.Sprintf("'%v'", x)
		}
		query = strings.ReplaceAll(query, placeholder, val)
	}
	return query
}

func (g *Repo) Ping(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, g.timeout)
	defer cancel()

	cmd := g.rdb.B().Arbitrary("GRAPH.QUERY").Args(g.graphName, "RETURN 1", "--compact").Build()
	return g.rdb.Do(ctx, cmd).Error()
}

func (g *Repo) EnsureSchema(ctx context.Context) {
	queries := []string{
		cypher.CreateUserIndex,
		cypher.CreateMerchantIndex,
		cypher.CreateExchangeIndex,
		cypher.CreateWalletIndex,
		cypher.CreatePaymentMethodIndex,
		cypher.CreateBankIndex,
		cypher.CreateDeviceIndex,
	}
	for _, q := range queries {
		_ = g.exec(ctx, q, false)
	}
}

func (g *Repo) DeleteGraph(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, g.timeout)
	defer cancel()

	// FalkorDB supports GRAPH.DELETE <graph>
	cmd := g.rdb.B().Arbitrary("GRAPH.DELETE").Args(g.graphName).Build()
	return g.rdb.Do(ctx, cmd).Error()
}

func (g *Repo) UpsertAggregated(ctx context.Context, ev model.CustomerEvent, et model.EventType, targetLabel, targetKeyProp, targetKey string, tsMillis int64, amount *float64) error {
	ctx, cancel := context.WithTimeout(ctx, g.timeout)
	defer cancel()

	relType := string(et)

	params := map[string]any{
		"user_id":    ev.UserID,
		"target_key": targetKey,
		"ts":         tsMillis,
	}

	query := fmt.Sprintf(
		cypher.UpsertAggregatedEdgeTemplate,
		targetLabel,
		targetKeyProp,
		relType)

	query = g.interpolate(query, params)

	return g.exec(ctx, query, true)
}

func (g *Repo) UpsertManualEdge(ctx context.Context, fromLabel, fromKeyProp, fromKey, toLabel, toKeyProp, toKey, relType string) error {
	ctx, cancel := context.WithTimeout(ctx, g.timeout)
	defer cancel()

	params := map[string]any{
		"from_key": fromKey,
		"to_key":   toKey,
		"ts":       time.Now().UnixMilli(),
	}

	query := fmt.Sprintf(
		cypher.UpsertManualEdgeTemplate,
		fromLabel,
		fromKeyProp,
		toLabel,
		toKeyProp,
		relType,
	)

	query = g.interpolate(query, params)

	return g.exec(ctx, query, true)
}

func (g *Repo) SubgraphHop(ctx context.Context, query string, params map[string]any) (any, error) {
	ctx, cancel := context.WithTimeout(ctx, g.timeout)
	defer cancel()

	if len(params) > 0 {
		query = g.interpolate(query, params)
	}

	start := time.Now()
	if g.log != nil {
		g.log.Info("graph_query", observability.Fields{
			"graph": g.graphName,
			"query": query,
		})
	}

	args := []string{g.graphName, query, "--compact"}
	cmd := g.rdb.B().Arbitrary("GRAPH.QUERY").Args(args...).Build()
	res := g.rdb.Do(ctx, cmd)
	if err := res.Error(); err != nil {
		if g.log != nil {
			g.log.Error("graph_query_error", observability.Fields{
				"graph": g.graphName,
				"query": query,
				"err":   err.Error(),
			})
		}
		return nil, err
	}
	respAny, err := res.ToAny()
	if err != nil {
		if g.log != nil {
			g.log.Error("graph_result_decode_error", observability.Fields{
				"graph": g.graphName,
				"query": query,
				"err":   err.Error(),
			})
		}
		return nil, err
	}
	if g.log != nil {
		g.log.Info("graph_result", observability.Fields{
			"graph":       g.graphName,
			"duration_ms": time.Since(start).Milliseconds(),
			"result":      respAny,
		})
	}
	return respAny, nil
}

func (g *Repo) QueryRows(ctx context.Context, query string, params map[string]any) ([]map[string]any, error) {
	respAny, err := g.SubgraphHop(ctx, query, params)
	if err != nil {
		return nil, err
	}
	return ParseCompact(respAny)
}

func (g *Repo) exec(ctx context.Context, query string, compact bool) error {
	args := []string{g.graphName, query}
	if compact {
		args = append(args, "--compact")
	}
	if g.log != nil {
		g.log.Debug("graph_exec", observability.Fields{
			"graph":   g.graphName,
			"query":   query,
			"compact": compact,
		})
	}
	cmd := g.rdb.B().Arbitrary("GRAPH.QUERY").Args(args...).Build()
	return g.rdb.Do(ctx, cmd).Error()
}

func StableNodeID(t model.NodeType, key string) string {
	return fmt.Sprintf("%s:%s", string(t), key)
}

func StableEdgeID(from, to, et string) string {
	h := sha1.Sum([]byte(from + "|" + to + "|" + et))
	return "e_" + hex.EncodeToString(h[:8])
}

func QuoteEdgeTypes(types []model.EventType) string {
	parts := make([]string, 0, len(types))
	for _, t := range types {
		parts = append(parts, fmt.Sprintf("'%s'", strings.ReplaceAll(string(t), "'", "")))
	}
	return strings.Join(parts, ",")
}
