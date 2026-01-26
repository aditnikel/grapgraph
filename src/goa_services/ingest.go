package goa_services

import (
	"context"

	"github.com/aditnikel/grapgraph/gen/ingest"
	"github.com/aditnikel/grapgraph/src/model"
	"github.com/aditnikel/grapgraph/src/service"
)

type IngestService struct {
	Ingest *service.IngestService
}

func (s *IngestService) PostEvent(ctx context.Context, p *ingest.CustomerEvent) (*ingest.IngestResponse, error) {
	ev := model.CustomerEvent{
		UserID:         p.UserID,
		MerchantIDMPAN: p.MerchantIDMpan,
		EventType:      p.EventType,
		EventTimestamp: p.EventTimestamp,
		TotalAmount:    p.TotalTransactionAmount,
		DeviceID:       p.DeviceID,
		PaymentMethod:  p.PaymentMethod,
		IssuingBank:    p.IssuingBank,
		WalletAddress:  p.WalletAddress,
		Exchange:       p.Exchange,
		IPAddress:      p.IPAddress,
	}

	if err := s.Ingest.AcceptEvent(ctx, ev); err != nil {
		return nil, ingest.BadRequest(err.Error())
	}

	return &ingest.IngestResponse{Accepted: true}, nil
}
