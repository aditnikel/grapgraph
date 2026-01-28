package domain

import (
	"context"
	"fmt"

	"github.com/aditnikel/grapgraph/src/config"
	"github.com/aditnikel/grapgraph/src/graph"
	"github.com/aditnikel/grapgraph/src/graph/cypher"
	"github.com/aditnikel/grapgraph/src/model"
)

type GraphService struct {
	Repo *graph.Repo
	Cfg  config.Config
}

func (s *GraphService) Ping(ctx context.Context) error {
	return s.Repo.Ping(ctx)
}

func (s *GraphService) Subgraph(ctx context.Context, req model.SubgraphRequest) (model.SubgraphResponse, error) {
	if req.Root.Type != "USER" {
		return model.SubgraphResponse{}, fmt.Errorf("root.type must be USER")
	}
	if req.Root.Key == "" {
		return model.SubgraphResponse{}, fmt.Errorf("root.key required")
	}
	if req.Hops < 1 || req.Hops > 3 {
		return model.SubgraphResponse{}, fmt.Errorf("hops must be 1..3")
	}

	if req.Limit.MaxNodes <= 0 {
		req.Limit.MaxNodes = s.Cfg.DefaultMaxNodes
	}
	if req.Limit.MaxEdges <= 0 {
		req.Limit.MaxEdges = s.Cfg.DefaultMaxEdges
	}

	edgeTypes := make([]model.EventType, 0, len(req.EdgeTypes))
	for _, t := range req.EdgeTypes {
		et, err := model.ParseEventType(t)
		if err != nil {
			return model.SubgraphResponse{}, err
		}
		edgeTypes = append(edgeTypes, et)
	}
	if len(edgeTypes) == 0 {
		edgeTypes = model.AllEventTypes()
	}
	quotedEdgeTypes := graph.QuoteEdgeTypes(edgeTypes)

	nodes := map[string]model.GraphNode{}
	edges := map[string]model.GraphEdge{}
	truncated := false

	remainingNodes := req.Limit.MaxNodes
	remainingEdges := req.Limit.MaxEdges

	rootID := graph.StableNodeID(model.NodeUser, req.Root.Key)
	nodes[rootID] = model.GraphNode{
		ID:    rootID,
		Type:  "USER",
		Key:   req.Root.Key,
		Label: "User " + req.Root.Key,
	}
	remainingNodes--

	putNode := func(nt, key string) bool {
		if nt == "" || key == "" {
			return false
		}
		id := graph.StableNodeID(model.NodeType(nt), key)
		if _, ok := nodes[id]; ok {
			return false
		}
		if remainingNodes <= 0 {
			truncated = true
			return false
		}
		nodes[id] = model.GraphNode{
			ID:    id,
			Type:  nt,
			Key:   key,
			Label: fmt.Sprintf("%s %s", nt, key),
		}
		remainingNodes--
		return true
	}

	putEdge := func(fromType, fromKey, toType, toKey, edgeType string) bool {
		if fromType == "" || fromKey == "" || toType == "" || toKey == "" || edgeType == "" {
			return false
		}
		fromID := graph.StableNodeID(model.NodeType(fromType), fromKey)
		toID := graph.StableNodeID(model.NodeType(toType), toKey)

		eid := graph.StableEdgeID(fromID, toID, edgeType)
		if _, ok := edges[eid]; ok {
			return false
		}
		if remainingEdges <= 0 {
			truncated = true
			return false
		}
		edges[eid] = model.GraphEdge{
			ID:       eid,
			Type:     edgeType,
			From:     fromID,
			To:       toID,
			Directed: true,
		}
		remainingEdges--
		return true
	}

	// Tight budget-aware hop limits
	hop1Limit := func() int {
		n := remainingEdges / 2
		if n < 1 {
			return 0
		}
		if n > 200 {
			n = 200
		}
		return n
	}
	hop2PerEntityLimit := func(numEntities int) int {
		if numEntities <= 0 {
			return 0
		}
		total := remainingEdges * 30 / 100
		if total < 1 {
			return 0
		}
		per := total / numEntities
		if per < 1 {
			per = 1
		}
		if per > 50 {
			per = 50
		}
		return per
	}
	hop3PerUserLimit := func(numUsers int) int {
		if numUsers <= 0 {
			return 0
		}
		total := remainingEdges * 20 / 100
		if total < 1 {
			return 0
		}
		per := total / numUsers
		if per < 1 {
			per = 1
		}
		if per > 30 {
			per = 30
		}
		return per
	}

	type entityRef struct {
		id int64
	}
	entityFrontier := []entityRef{}
	userFrontier := []string{}

	// Hop 1: USER -> ENTITY
	if req.Hops >= 1 && !truncated {
		limit := hop1Limit()
		if limit == 0 {
			truncated = true
		} else {
			q := fmt.Sprintf(cypher.UserToEntityTemplate, quotedEdgeTypes)
			rows, err := s.Repo.QueryRows(ctx, q, map[string]any{
				"user_id": req.Root.Key,
				"limit":   limit,
			})
			if err != nil {
				return model.SubgraphResponse{}, fmt.Errorf("graph query hop1 failed: %v", err)
			}

			for _, r := range rows {
				if truncated {
					break
				}
				fromType := fmt.Sprint(r["from_type"])
				fromKey := fmt.Sprint(r["from_key"])
				toType := fmt.Sprint(r["to_type"])
				toKey := fmt.Sprint(r["to_key"])
				et := fmt.Sprint(r["edge_type"])

				if toType == "UNKNOWN" || toKey == "" {
					continue
				}

				_ = putNode(fromType, fromKey)
				_ = putNode(toType, toKey)
				_ = putEdge(fromType, fromKey, toType, toKey, et)

				if eid, ok := s.resolveEntityInternalID(ctx, toType, toKey); ok {
					entityFrontier = append(entityFrontier, entityRef{id: eid})
				}

				if remainingEdges <= 0 || remainingNodes <= 0 {
					truncated = true
					break
				}
			}
		}
	}

	// Hop 2: ENTITY -> USER
	if req.Hops >= 2 && !truncated {
		perEntity := hop2PerEntityLimit(len(entityFrontier))
		if perEntity == 0 {
			truncated = true
		} else {
			seenUsers := map[string]struct{}{}
			for _, e := range entityFrontier {
				if truncated {
					break
				}
				if remainingEdges <= 0 || remainingNodes <= 0 {
					truncated = true
					break
				}

				q := fmt.Sprintf(cypher.EntityToUserTemplate, quotedEdgeTypes)
				rows, err := s.Repo.QueryRows(ctx, q, map[string]any{
					"entity_id": e.id,
					"limit":     perEntity,
				})
				if err != nil {
					return model.SubgraphResponse{}, fmt.Errorf("graph query hop2 failed: %v", err)
				}

				for _, r := range rows {
					if truncated {
						break
					}
					fromType := fmt.Sprint(r["from_type"])
					fromKey := fmt.Sprint(r["from_key"])
					toType := fmt.Sprint(r["to_type"])
					toKey := fmt.Sprint(r["to_key"])
					et := fmt.Sprint(r["edge_type"])

					if toKey == "" || fromType == "UNKNOWN" || fromKey == "" {
						continue
					}

					_ = putNode(fromType, fromKey)
					_ = putNode(toType, toKey)
					_ = putEdge(fromType, fromKey, toType, toKey, et)

					if _, ok := seenUsers[toKey]; !ok {
						seenUsers[toKey] = struct{}{}
						userFrontier = append(userFrontier, toKey)
					}

					if remainingEdges <= 0 || remainingNodes <= 0 {
						truncated = true
						break
					}
				}
			}
		}
	}

	// Hop 3: USER -> ENTITY
	if req.Hops >= 3 && !truncated {
		perUser := hop3PerUserLimit(len(userFrontier))
		if perUser == 0 {
			truncated = true
		} else {
			for _, uid := range userFrontier {
				if truncated {
					break
				}
				if remainingEdges <= 0 || remainingNodes <= 0 {
					truncated = true
					break
				}

				q := fmt.Sprintf(cypher.UserToEntityTemplate, quotedEdgeTypes)
				rows, err := s.Repo.QueryRows(ctx, q, map[string]any{
					"user_id": uid,
					"limit":   perUser,
				})
				if err != nil {
					return model.SubgraphResponse{}, fmt.Errorf("graph query hop3 failed: %v", err)
				}

				for _, r := range rows {
					if truncated {
						break
					}
					fromType := fmt.Sprint(r["from_type"])
					fromKey := fmt.Sprint(r["from_key"])
					toType := fmt.Sprint(r["to_type"])
					toKey := fmt.Sprint(r["to_key"])
					et := fmt.Sprint(r["edge_type"])

					if toType == "UNKNOWN" || toKey == "" {
						continue
					}

					_ = putNode(fromType, fromKey)
					_ = putNode(toType, toKey)
					_ = putEdge(fromType, fromKey, toType, toKey, et)

					if remainingEdges <= 0 || remainingNodes <= 0 {
						truncated = true
						break
					}
				}
			}
		}
	}

	return model.SubgraphResponse{
		Version:   "1.0",
		Root:      rootID,
		Nodes:     mapToSlice(nodes),
		Edges:     mapToSliceEdges(edges),
		Truncated: truncated,
	}, nil
}

func (s *GraphService) resolveEntityInternalID(ctx context.Context, typ, key string) (int64, bool) {
	var matchExpr string
	switch typ {
	case "MERCHANT":
		matchExpr = "n:Merchant AND n.merchant_id_mpan = $key"
	case "EXCHANGE":
		matchExpr = "n:Exchange AND n.exchange = $key"
	case "WALLET":
		matchExpr = "n:Wallet AND n.wallet_address = $key"
	case "PAYMENT_METHOD":
		matchExpr = "n:PaymentMethod AND n.payment_method = $key"
	case "BANK":
		matchExpr = "n:Bank AND n.issuing_bank = $key"
	case "DEVICE":
		matchExpr = "n:Device AND n.device_id = $key"
	default:
		return 0, false
	}

	q := fmt.Sprintf(cypher.EntityInternalIDByKey, matchExpr)
	rows, err := s.Repo.QueryRows(ctx, q, map[string]any{"key": key})
	if err != nil || len(rows) == 0 {
		return 0, false
	}
	idv, ok := rows[0]["entity_id"]
	if !ok {
		return 0, false
	}
	return toInt64(idv)
}

func buildMetrics(r map[string]any) map[string]any {
	eventCount, _ := toInt64(r["event_count"])
	totalAmount := mustFloat64(r["total_amount"])
	avg := 0.0
	if eventCount > 0 {
		avg = totalAmount / float64(eventCount)
	}
	return map[string]any{
		"event_count":           eventCount,
		"event_count_30d":       mustInt64(r["event_count_30d"]),
		"distinct_ip_count_30d": mustInt64(r["distinct_ip_count_30d"]),
		"first_seen":            mustInt64(r["first_seen"]),
		"last_seen":             mustInt64(r["last_seen"]),
		"total_amount":          totalAmount,
		"max_amount":            mustFloat64(r["max_amount"]),
		"avg_amount":            avg,
	}
}

func toInt64(v any) (int64, bool) {
	switch x := v.(type) {
	case int64:
		return x, true
	case int:
		return int64(x), true
	case float64:
		return int64(x), true
	case string:
		var i int64
		_, err := fmt.Sscan(x, &i)
		return i, err == nil
	default:
		return 0, false
	}
}

func mustInt64(v any) int64 {
	i, _ := toInt64(v)
	return i
}

func mustFloat64(v any) float64 {
	switch x := v.(type) {
	case float64:
		return x
	case int64:
		return float64(x)
	case int:
		return float64(x)
	case string:
		var f float64
		_, _ = fmt.Sscan(x, &f)
		return f
	default:
		return 0
	}
}

func mapToSlice(m map[string]model.GraphNode) []model.GraphNode {
	out := make([]model.GraphNode, 0, len(m))
	for _, v := range m {
		out = append(out, v)
	}
	return out
}

func mapToSliceEdges(m map[string]model.GraphEdge) []model.GraphEdge {
	out := make([]model.GraphEdge, 0, len(m))
	for _, v := range m {
		out = append(out, v)
	}
	return out
}
