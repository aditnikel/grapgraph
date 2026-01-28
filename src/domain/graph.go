package domain

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aditnikel/grapgraph/src/infra/config"
	"github.com/aditnikel/grapgraph/src/infra/graph"
	"github.com/aditnikel/grapgraph/src/infra/graph/cypher"
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
	if req.Hops < 1 {
		return model.SubgraphResponse{}, fmt.Errorf("hops must be >= 1")
	}
	if req.MinEventCount < 0 {
		return model.SubgraphResponse{}, fmt.Errorf("min_event_count must be >= 0")
	}

	if req.Limit.MaxNodes <= 0 {
		req.Limit.MaxNodes = s.Cfg.DefaultMaxNodes
	}
	if req.Limit.MaxEdges <= 0 {
		req.Limit.MaxEdges = s.Cfg.DefaultMaxEdges
	}

	// If specific types are requested, filter by them. Otherwise, include all.
	whereClause := "true"
	if len(req.EdgeTypes) > 0 {
		edgeTypes := make([]string, 0, len(req.EdgeTypes))
		for _, t := range req.EdgeTypes {
			// Relaxed validation already allows dynamic types
			et, err := model.ParseEventType(t)
			if err != nil {
				return model.SubgraphResponse{}, err
			}
			edgeTypes = append(edgeTypes, string(et))
		}
		whereClause = fmt.Sprintf("type(r) IN [%s]", graph.QuoteEdgeTypes(toEventTypes(edgeTypes)))
	}
	if req.MinEventCount > 0 {
		whereClause = fmt.Sprintf("(%s) AND coalesce(r.event_count, 0) >= $min_event_count", whereClause)
	}

	windowStart := int64(0)
	if req.TimeWindowMs > 0 {
		windowStart = time.Now().UnixMilli() - req.TimeWindowMs
		if windowStart < 0 {
			windowStart = 0
		}
	}

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

	asBool := func(v any) bool {
		switch x := v.(type) {
		case bool:
			return x
		case int:
			return x != 0
		case int64:
			return x != 0
		case float64:
			return x != 0
		case string:
			return x == "true" || x == "1"
		default:
			return false
		}
	}

	putEdge := func(fromType, fromKey, toType, toKey, edgeType string, manual bool) bool {
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
			Manual:   manual,
		}
		remainingEdges--
		return true
	}

	// Tight budget-aware hop limits
	perFrontierLimit := func(frontierLen, total, cap int) int {
		if frontierLen <= 0 {
			return 0
		}
		if total < 1 {
			return 0
		}
		per := total / frontierLen
		if per < 1 {
			per = 1
		}
		if per > cap {
			per = cap
		}
		return per
	}
	totalForHop := func(hop int) (int, int) {
		switch {
		case hop == 1:
			return remainingEdges / 2, 200
		case hop%2 == 0:
			return remainingEdges * 30 / 100, 50
		default:
			return remainingEdges * 20 / 100, 30
		}
	}

	type entityRef struct {
		id int64
	}
	userFrontier := []string{req.Root.Key}
	entityFrontier := []entityRef{}

	for hop := 1; hop <= req.Hops && !truncated; hop++ {
		if hop%2 == 1 {
			if len(userFrontier) == 0 {
				break
			}
			total, cap := totalForHop(hop)
			perUser := perFrontierLimit(len(userFrontier), total, cap)
			if perUser == 0 {
				truncated = true
				break
			}

			nextEntities := []entityRef{}
			seenEntities := map[int64]struct{}{}
			q := fmt.Sprintf(cypher.UserToEntityTemplate, whereClause)
			for _, uid := range userFrontier {
				if truncated {
					break
				}
				if remainingEdges <= 0 || remainingNodes <= 0 {
					truncated = true
					break
				}

				rows, err := s.Repo.QueryRows(ctx, q, map[string]any{
					"user_id":         uid,
					"limit":           perUser,
					"window_start":    windowStart,
					"min_event_count": req.MinEventCount,
				})
				if err != nil {
					return model.SubgraphResponse{}, fmt.Errorf("graph query hop%d failed: %v", hop, err)
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
					manual := asBool(r["edge_manual"])

					if toType == "UNKNOWN" || toKey == "" {
						continue
					}

					_ = putNode(fromType, fromKey)
					_ = putNode(toType, toKey)
					_ = putEdge(fromType, fromKey, toType, toKey, et, manual)

					if eid, ok := s.resolveEntityInternalID(ctx, toType, toKey); ok {
						if _, seen := seenEntities[eid]; !seen {
							seenEntities[eid] = struct{}{}
							nextEntities = append(nextEntities, entityRef{id: eid})
						}
					}

					if remainingEdges <= 0 || remainingNodes <= 0 {
						truncated = true
						break
					}
				}
			}
			userFrontier = nil
			entityFrontier = nextEntities
		} else {
			if len(entityFrontier) == 0 {
				break
			}
			total, cap := totalForHop(hop)
			perEntity := perFrontierLimit(len(entityFrontier), total, cap)
			if perEntity == 0 {
				truncated = true
				break
			}

			nextUsers := []string{}
			seenUsers := map[string]struct{}{}
			q := fmt.Sprintf(cypher.EntityToUserTemplate, whereClause)
			for _, e := range entityFrontier {
				if truncated {
					break
				}
				if remainingEdges <= 0 || remainingNodes <= 0 {
					truncated = true
					break
				}

				rows, err := s.Repo.QueryRows(ctx, q, map[string]any{
					"entity_id":       e.id,
					"limit":           perEntity,
					"window_start":    windowStart,
					"min_event_count": req.MinEventCount,
				})
				if err != nil {
					return model.SubgraphResponse{}, fmt.Errorf("graph query hop%d failed: %v", hop, err)
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
					manual := asBool(r["edge_manual"])

					if toKey == "" || fromType == "UNKNOWN" || fromKey == "" {
						continue
					}

					_ = putNode(fromType, fromKey)
					_ = putNode(toType, toKey)
					_ = putEdge(fromType, fromKey, toType, toKey, et, manual)

					if _, ok := seenUsers[toKey]; !ok {
						seenUsers[toKey] = struct{}{}
						nextUsers = append(nextUsers, toKey)
					}

					if remainingEdges <= 0 || remainingNodes <= 0 {
						truncated = true
						break
					}
				}
			}
			entityFrontier = nil
			userFrontier = nextUsers
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

func (s *GraphService) GetMetadata(ctx context.Context) (model.MetadataResponse, error) {
	nodeTypes, err := s.Repo.QueryRows(ctx, cypher.QueryNodeLabels, nil)
	if err != nil {
		return model.MetadataResponse{}, err
	}
	edgeTypes, err := s.Repo.QueryRows(ctx, cypher.QueryRelationshipTypes, nil)
	if err != nil {
		return model.MetadataResponse{}, err
	}

	nt := make([]string, 0, len(nodeTypes))
	for _, r := range nodeTypes {
		if v, ok := r["label"].(string); ok {
			nt = append(nt, v)
		}
	}

	et := make([]string, 0, len(edgeTypes))
	for _, r := range edgeTypes {
		if v, ok := r["relationshipType"].(string); ok {
			et = append(et, v)
		}
	}

	return model.MetadataResponse{
		NodeTypes: nt,
		EdgeTypes: et,
	}, nil
}

func (s *GraphService) CreateManualEdge(ctx context.Context, req model.ManualEdgeRequest) (model.GraphEdge, error) {
	fromType := strings.TrimSpace(strings.ToUpper(req.From.Type))
	toType := strings.TrimSpace(strings.ToUpper(req.To.Type))
	fromKey := strings.TrimSpace(req.From.Key)
	toKey := strings.TrimSpace(req.To.Key)
	edgeType, err := validateEdgeType(req.EdgeType)
	if err != nil {
		return model.GraphEdge{}, err
	}
	if fromType == "" || toType == "" {
		return model.GraphEdge{}, fmt.Errorf("from.type and to.type required")
	}
	if fromKey == "" || toKey == "" {
		return model.GraphEdge{}, fmt.Errorf("from.key and to.key required")
	}

	fromLabel, fromKeyProp, fromNodeType, ok := nodeSpecForType(fromType)
	if !ok {
		return model.GraphEdge{}, fmt.Errorf("invalid from.type: %s", fromType)
	}
	toLabel, toKeyProp, toNodeType, ok := nodeSpecForType(toType)
	if !ok {
		return model.GraphEdge{}, fmt.Errorf("invalid to.type: %s", toType)
	}

	if err := s.Repo.UpsertManualEdge(ctx, fromLabel, fromKeyProp, fromKey, toLabel, toKeyProp, toKey, edgeType); err != nil {
		return model.GraphEdge{}, err
	}

	fromID := graph.StableNodeID(fromNodeType, fromKey)
	toID := graph.StableNodeID(toNodeType, toKey)

	return model.GraphEdge{
		ID:       graph.StableEdgeID(fromID, toID, edgeType),
		Type:     edgeType,
		From:     fromID,
		To:       toID,
		Directed: true,
		Manual:   true,
	}, nil
}

func nodeSpecForType(t string) (label, keyProp string, nodeType model.NodeType, ok bool) {
	switch t {
	case string(model.NodeUser):
		return "User", "user_id", model.NodeUser, true
	case string(model.NodeMerchant):
		return "Merchant", "merchant_id_mpan", model.NodeMerchant, true
	case string(model.NodeExchange):
		return "Exchange", "exchange", model.NodeExchange, true
	case string(model.NodeWallet):
		return "Wallet", "wallet_address", model.NodeWallet, true
	case string(model.NodePaymentMethod):
		return "PaymentMethod", "payment_method", model.NodePaymentMethod, true
	case string(model.NodeBank):
		return "Bank", "issuing_bank", model.NodeBank, true
	case string(model.NodeDevice):
		return "Device", "device_id", model.NodeDevice, true
	default:
		return "", "", "", false
	}
}

func validateEdgeType(edgeType string) (string, error) {
	et := strings.TrimSpace(strings.ToUpper(edgeType))
	if et == "" {
		return "", fmt.Errorf("edge_type required")
	}
	for _, r := range et {
		if (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' || r == '-' {
			continue
		}
		return "", fmt.Errorf("invalid edge_type: %s (must be A-Z, 0-9, _, -)", edgeType)
	}
	return et, nil
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

func toEventTypes(s []string) []model.EventType {
	out := make([]model.EventType, len(s))
	for i, v := range s {
		out[i] = model.EventType(v)
	}
	return out
}
