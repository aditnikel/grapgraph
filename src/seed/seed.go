package seed

import (
	"context"
	"time"

	"github.com/aditnikel/grapgraph/src/ingest"
	"github.com/aditnikel/grapgraph/src/model"
)

type repo interface {
	UpsertAggregated(ctx context.Context, ev model.CustomerEvent, et model.EventType, targetLabel, targetKeyProp, targetKey string, tsMillis int64, amount *float64) error
}

// SeedDemo creates deterministic demo data for FE visualization.
// Exercises: fanout, shared entities, multiple edge types, and money-bearing aggregates.
func SeedDemo(ctx context.Context, r repo) error {
	base := time.Now().Add(-48 * time.Hour).UnixMilli()

	str := func(s string) *string { return &s }
	f64 := func(v float64) *float64 { return &v }

	events := []model.CustomerEvent{
		{UserID: "u_123", DeviceID: str("dev_a"), EventType: "LOGIN", EventTimestamp: float64(base + 1_000)},
		{UserID: "u_123", MerchantIDMPAN: str("m_777"), EventType: "PAYMENT", EventTimestamp: float64(base + 10_000), TotalAmount: f64(42.50)},
		{UserID: "u_123", PaymentMethod: str("card_visa_x1"), EventType: "PAYMENT", EventTimestamp: float64(base + 20_000), TotalAmount: f64(15.00)},
		{UserID: "u_123", WalletAddress: str("0xabc123"), EventType: "TRANSACTION", EventTimestamp: float64(base + 30_000), TotalAmount: f64(250.00)},
		{UserID: "u_123", Exchange: str("binance"), EventType: "LOGIN", EventTimestamp: float64(base + 40_000)},
		{UserID: "u_123", IssuingBank: str("bank_a"), EventType: "KYC_UPDATE", EventTimestamp: float64(base + 50_000)},

		{UserID: "u_555", DeviceID: str("dev_a"), EventType: "LOGIN", EventTimestamp: float64(base + 60_000)},
		{UserID: "u_555", MerchantIDMPAN: str("m_777"), EventType: "PAYMENT", EventTimestamp: float64(base + 70_000), TotalAmount: f64(99.99)},
		{UserID: "u_555", Exchange: str("binance"), EventType: "LOGIN", EventTimestamp: float64(base + 80_000)},

		{UserID: "u_999", PaymentMethod: str("card_visa_x1"), EventType: "PAYMENT", EventTimestamp: float64(base + 90_000), TotalAmount: f64(5.00)},
		{UserID: "u_999", WalletAddress: str("0xabc123"), EventType: "WITHDRAWAL", EventTimestamp: float64(base + 100_000), TotalAmount: f64(10.00)},

		{UserID: "u_123", MerchantIDMPAN: str("m_777"), EventType: "PAYMENT", EventTimestamp: float64(base + 110_000), TotalAmount: f64(200.00)},
		{UserID: "u_123", MerchantIDMPAN: str("m_777"), EventType: "PAYMENT", EventTimestamp: float64(base + 120_000), TotalAmount: f64(300.00)},
		{UserID: "u_123", DeviceID: str("dev_a"), EventType: "LOGIN", EventTimestamp: float64(base + 130_000)},
		{UserID: "u_123", DeviceID: str("dev_a"), EventType: "LOGIN", EventTimestamp: float64(base + 140_000)},

		{UserID: "u_777", IssuingBank: str("bank_a"), EventType: "ACCOUNT_UPDATE", EventTimestamp: float64(base + 150_000)},
		{UserID: "u_777", MerchantIDMPAN: str("m_777"), EventType: "PAYMENT", EventTimestamp: float64(base + 160_000), TotalAmount: f64(75.00)},
	}

	for _, ev := range events {
		et, err := model.ParseEventType(ev.EventType)
		if err != nil {
			return err
		}
		ts := int64(ev.EventTimestamp.(float64))

		label, prop, key, ok := ingest.ChooseTarget(ev)
		if !ok {
			continue
		}
		if err := r.UpsertAggregated(ctx, ev, et, label, prop, key, ts, ev.TotalAmount); err != nil {
			return err
		}
	}
	return nil
}
