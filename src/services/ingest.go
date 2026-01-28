package goa_services

import (
	"context"
	"fmt"

	"github.com/aditnikel/grapgraph/gen/ingest"
	"github.com/aditnikel/grapgraph/src/domain"
	"github.com/aditnikel/grapgraph/src/model"
)

type IngestService struct {
	Ingest *domain.IngestService
}

func (s *IngestService) PostEvent(ctx context.Context, p *ingest.BulkCustomerEvents) (*ingest.BulkIngestResponse, error) {
	if p == nil || len(p.Events) == 0 {
		return nil, ingest.BadRequest("events must contain at least one item")
	}

	events := make([]model.CustomerEvent, 0, len(p.Events))
	for i, e := range p.Events {
		if e == nil {
			return nil, ingest.BadRequest(fmt.Sprintf("event[%d] is null", i))
		}
		events = append(events, model.CustomerEvent{
			UserID:         e.UserID,
			MerchantIDMPAN: e.MerchantIDMpan,
			EventType:      e.EventType,
			EventTimestamp: e.EventTimestamp,
			TotalAmount:    e.TotalTransactionAmount,
			DeviceID:       e.DeviceID,
			PaymentMethod:  e.PaymentMethod,
			IssuingBank:    e.IssuingBank,
			WalletAddress:  e.WalletAddress,
			Exchange:       e.Exchange,
			IPAddress:      e.IPAddress,
		})
	}

	acceptedCount, err := s.Ingest.AcceptEvents(ctx, events)
	if err != nil {
		return nil, ingest.BadRequest(err.Error())
	}

	return &ingest.BulkIngestResponse{
		Accepted:      true,
		AcceptedCount: acceptedCount,
		FailedCount:   len(events) - acceptedCount,
	}, nil
}
