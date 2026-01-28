package goa_services

import (
	"context"

	"github.com/aditnikel/grapgraph/gen/graph"
	"github.com/aditnikel/grapgraph/src/domain"
	"github.com/aditnikel/grapgraph/src/model"
)

type GraphService struct {
	Graph *domain.GraphService
}

func (s *GraphService) GetMetadata(ctx context.Context) (*graph.MetadataResponse, error) {
	ets := make([]string, 0, len(model.AllEventTypes()))
	for _, e := range model.AllEventTypes() {
		ets = append(ets, string(e))
	}

	return &graph.MetadataResponse{
		NodeTypes: []string{
			string(model.NodeUser),
			string(model.NodeMerchant),
			string(model.NodeExchange),
			string(model.NodeWallet),
			string(model.NodePaymentMethod),
			string(model.NodeBank),
			string(model.NodeDevice),
		},
		EdgeTypes: ets,
	}, nil
}

func (s *GraphService) PostSubgraph(ctx context.Context, p *graph.SubgraphRequest) (*graph.SubgraphResponse, error) {
	req := model.SubgraphRequest{
		Hops:      p.Hops,
		EdgeTypes: p.EdgeTypes,
	}

	req.Root.Type = p.Root.Type
	req.Root.Key = p.Root.Key

	req.Limit.MaxNodes = p.Limit.MaxNodes
	req.Limit.MaxEdges = p.Limit.MaxEdges

	resp, err := s.Graph.Subgraph(ctx, req)
	if err != nil {
		return nil, graph.BadRequest(err.Error())
	}

	nodes := make([]*graph.GraphNode, len(resp.Nodes))
	for i, n := range resp.Nodes {
		nodes[i] = &graph.GraphNode{
			ID:    n.ID,
			Type:  n.Type,
			Key:   n.Key,
			Label: n.Label,
			Props: n.Props,
		}
	}

	edges := make([]*graph.GraphEdge, len(resp.Edges))
	for i, e := range resp.Edges {
		edges[i] = &graph.GraphEdge{
			ID:       e.ID,
			Type:     e.Type,
			From:     e.From,
			To:       e.To,
			Directed: e.Directed,
			Manual:   e.Manual,
		}
	}

	return &graph.SubgraphResponse{
		Version:   resp.Version,
		Root:      resp.Root,
		Nodes:     nodes,
		Edges:     edges,
		Truncated: resp.Truncated,
	}, nil
}

func (s *GraphService) PostManualEdge(ctx context.Context, p *graph.ManualEdgeRequest) (*graph.GraphEdge, error) {
	req := model.ManualEdgeRequest{
		EdgeType: p.EdgeType,
	}
	req.From.Type = p.From.Type
	req.From.Key = p.From.Key
	req.To.Type = p.To.Type
	req.To.Key = p.To.Key

	edge, err := s.Graph.CreateManualEdge(ctx, req)
	if err != nil {
		return nil, graph.BadRequest(err.Error())
	}

	return &graph.GraphEdge{
		ID:       edge.ID,
		Type:     edge.Type,
		From:     edge.From,
		To:       edge.To,
		Directed: edge.Directed,
		Manual:   edge.Manual,
	}, nil
}
