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
)

type Repo struct {
	rdb       rueidis.Client
	graphName string
	timeout   time.Duration
}

func New(rdb rueidis.Client, graphName string, timeout time.Duration) *Repo {
	return &Repo{rdb: rdb, graphName: graphName, timeout: timeout}
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
		_ = g.exec(ctx, q, nil, false)
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

	moneyBlock := ""
	params := map[string]any{
		"user_id":    ev.UserID,
		"target_key": targetKey,
		"ts":         tsMillis,
	}
	if model.IsMoneyBearing(et) && amount != nil {
		moneyBlock = cypher.MoneyAggBlock
		params["amount"] = *amount
	}

	query := fmt.Sprintf(
		cypher.UpsertAggregatedEdgeTemplate,
		targetLabel,
		targetKeyProp,
		relType,
		moneyBlock,
	)

	return g.exec(ctx, query, params, true)
}

func (g *Repo) SubgraphHop(ctx context.Context, query string, params map[string]any) (any, error) {
	ctx, cancel := context.WithTimeout(ctx, g.timeout)
	defer cancel()

	args := []string{g.graphName, query, "--compact"}
	if len(params) > 0 {
		args = append(args, "PARAMS")
		for k, v := range params {
			args = append(args, k, fmt.Sprintf("%v", v))
		}
	}
	cmd := g.rdb.B().Arbitrary("GRAPH.QUERY").Args(args...).Build()
	res := g.rdb.Do(ctx, cmd)
	if err := res.Error(); err != nil {
		return nil, err
	}
	return res.ToAny()
}

func (g *Repo) QueryRows(ctx context.Context, query string, params map[string]any) ([]map[string]any, error) {
	respAny, err := g.SubgraphHop(ctx, query, params)
	if err != nil {
		return nil, err
	}
	return ParseCompact(respAny)
}

func (g *Repo) exec(ctx context.Context, query string, params map[string]any, compact bool) error {
	args := []string{g.graphName, query}
	if compact {
		args = append(args, "--compact")
	}
	if len(params) > 0 {
		args = append(args, "PARAMS")
		for k, v := range params {
			args = append(args, k, fmt.Sprintf("%v", v))
		}
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
