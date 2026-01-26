package goa_services

import (
	"context"

	"github.com/aditnikel/grapgraph/gen/health"
	"github.com/aditnikel/grapgraph/src/observability"
	"github.com/aditnikel/grapgraph/src/service"
)

type HealthService struct {
	Log   *observability.Logger
	Graph *service.GraphService
}

func (s *HealthService) Get(ctx context.Context) (*health.HealthResponse, error) {
	if err := s.Graph.Ping(ctx); err != nil {
		return &health.HealthResponse{OK: false, Error: strPtr(err.Error())}, nil
	}
	return &health.HealthResponse{OK: true}, nil
}

func strPtr(s string) *string { return &s }
