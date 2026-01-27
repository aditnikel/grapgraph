package domain

import (
	"context"
	"fmt"

	"github.com/aditnikel/grapgraph/src/graph"
	"github.com/aditnikel/grapgraph/src/ingest"
	"github.com/aditnikel/grapgraph/src/model"
)

type IngestService struct {
	Repo *graph.Repo
}

func (s *IngestService) AcceptEvent(ctx context.Context, ev model.CustomerEvent) error {
	if ev.UserID == "" {
		return fmt.Errorf("user_id required")
	}
	et, err := model.ParseEventType(ev.EventType)
	if err != nil {
		return err
	}
	tsMillis, err := ingest.ParseEventTimestamp(ev.EventTimestamp)
	if err != nil {
		return err
	}

	label, keyProp, keyValue, ok := ingest.ChooseTarget(ev)
	if !ok {
		// No edge target -> no edge. (Optionally update user features later.)
		return nil
	}

	return s.Repo.UpsertAggregated(ctx, ev, et, label, keyProp, keyValue, tsMillis, ev.TotalAmount)
}

func (s *IngestService) AcceptEvents(ctx context.Context, events []model.CustomerEvent) (int, error) {
	accepted := 0
	for idx, ev := range events {
		if err := s.AcceptEvent(ctx, ev); err != nil {
			return accepted, fmt.Errorf("event[%d]: %w", idx, err)
		}
		accepted++
	}
	return accepted, nil
}
