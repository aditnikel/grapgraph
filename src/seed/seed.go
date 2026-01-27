package seed

import (
	"context"
	"time"

	"github.com/aditnikel/grapgraph/src/model"
)

type ingester interface {
	AcceptEvent(ctx context.Context, ev model.CustomerEvent) error
}

// SeedDemo creates deterministic demo data for FE visualization.
// Exercises: fanout, shared entities, multiple edge types, and money-bearing aggregates.
func SeedDemo(ctx context.Context, ingestSvc ingester) error {
	base := time.Now().Add(-48 * time.Hour).UnixMilli()

	str := func(s string) *string { return &s }
	f64 := func(v float64) *float64 { return &v }

	events := []model.CustomerEvent{
		// -- Normal User: Alice (u_123) --
		{UserID: "u_123", DeviceID: str("iphone_15_pro"), EventType: "LOGIN", EventTimestamp: float64(base + 1_000)},
		{UserID: "u_123", MerchantIDMPAN: str("starbucks_sg_01"), EventType: "PAYMENT", EventTimestamp: float64(base + 10_000), TotalAmount: f64(5.50)},
		{UserID: "u_123", PaymentMethod: str("visa_9988"), EventType: "PAYMENT", EventTimestamp: float64(base + 20_000), TotalAmount: f64(5.50)},
		{UserID: "u_123", WalletAddress: str("0x71C765..."), EventType: "TRANSACTION", EventTimestamp: float64(base + 30_000), TotalAmount: f64(120.00)},
		{UserID: "u_123", Exchange: str("coinbase"), EventType: "LOGIN", EventTimestamp: float64(base + 40_000)},
		{UserID: "u_123", IssuingBank: str("dbs_bank"), EventType: "KYC_UPDATE", EventTimestamp: float64(base + 50_000)},
		{UserID: "u_123", DeviceID: str("iphone_15_pro"), EventType: "PAYMENT", EventTimestamp: float64(base + 3600_000), TotalAmount: f64(45.00)},

		// -- Suspicious Pattern: Multi-user Shared Device --
		{UserID: "u_bot_1", DeviceID: str("emulator_v3"), EventType: "REGISTER", EventTimestamp: float64(base + 100)},
		{UserID: "u_bot_2", DeviceID: str("emulator_v3"), EventType: "REGISTER", EventTimestamp: float64(base + 200)},
		{UserID: "u_bot_3", DeviceID: str("emulator_v3"), EventType: "REGISTER", EventTimestamp: float64(base + 300)},
		{UserID: "u_bot_4", DeviceID: str("emulator_v3"), EventType: "REGISTER", EventTimestamp: float64(base + 400)},
		{UserID: "u_bot_1", MerchantIDMPAN: str("global_casino_x"), EventType: "PAYMENT", EventTimestamp: float64(base + 5000), TotalAmount: f64(500.0)},
		{UserID: "u_bot_2", MerchantIDMPAN: str("global_casino_x"), EventType: "PAYMENT", EventTimestamp: float64(base + 6000), TotalAmount: f64(500.0)},

		// -- Shared Wallet Pattern --
		{UserID: "u_mule_1", WalletAddress: str("0xDEADBEEF..."), EventType: "WITHDRAWAL", EventTimestamp: float64(base + 100_000), TotalAmount: f64(1000.0)},
		{UserID: "u_mule_2", WalletAddress: str("0xDEADBEEF..."), EventType: "WITHDRAWAL", EventTimestamp: float64(base + 110_000), TotalAmount: f64(2000.0)},
		{UserID: "u_mule_3", WalletAddress: str("0xDEADBEEF..."), EventType: "WITHDRAWAL", EventTimestamp: float64(base + 120_000), TotalAmount: f64(1500.0)},

		// -- High Velocity Payment: Bob (u_555) --
		{UserID: "u_555", DeviceID: str("pixel_8"), EventType: "LOGIN", EventTimestamp: float64(base + 60_000)},
		{UserID: "u_555", MerchantIDMPAN: str("steam_games"), EventType: "PAYMENT", EventTimestamp: float64(base + 61_000), TotalAmount: f64(10.00)},
		{UserID: "u_555", MerchantIDMPAN: str("steam_games"), EventType: "PAYMENT", EventTimestamp: float64(base + 62_000), TotalAmount: f64(15.00)},
		{UserID: "u_555", MerchantIDMPAN: str("steam_games"), EventType: "PAYMENT", EventTimestamp: float64(base + 63_000), TotalAmount: f64(20.00)},
		{UserID: "u_555", MerchantIDMPAN: str("steam_games"), EventType: "PAYMENT", EventTimestamp: float64(base + 64_000), TotalAmount: f64(50.00)},
		{UserID: "u_555", Exchange: str("binance"), EventType: "LOGIN", EventTimestamp: float64(base + 80_000)},

		// -- Normal Merchant Activity: Charlie (u_999) --
		{UserID: "u_999", PaymentMethod: str("mastercard_1122"), EventType: "PAYMENT", EventTimestamp: float64(base + 90_000), TotalAmount: f64(5.00)},
		{UserID: "u_999", MerchantIDMPAN: str("starbucks_sg_01"), EventType: "PAYMENT", EventTimestamp: float64(base + 95_000), TotalAmount: f64(6.50)},
		{UserID: "u_999", WalletAddress: str("0xabc123..."), EventType: "WITHDRAWAL", EventTimestamp: float64(base + 100_000), TotalAmount: f64(10.00)},

		// -- ATO (Account Takeover) Simulation: Dave (u_001) --
		{UserID: "u_001", DeviceID: str("daves_macbook"), EventType: "LOGIN", EventTimestamp: float64(base + 10_000)},
		{UserID: "u_001", DeviceID: str("attacker_kali_linux"), EventType: "LOGIN", EventTimestamp: float64(base + 48*3600_000 - 10_000)},
		{UserID: "u_001", DeviceID: str("attacker_kali_linux"), EventType: "PASSWORD_CHANGE", EventTimestamp: float64(base + 48*3600_000 - 5_000)},
		{UserID: "u_001", DeviceID: str("attacker_kali_linux"), EventType: "WITHDRAWAL", EventTimestamp: float64(base + 48*3600_000 - 1_000), TotalAmount: f64(10000.0)},
	}

	for _, ev := range events {
		if err := ingestSvc.AcceptEvent(ctx, ev); err != nil {
			return err
		}
	}
	return nil
}
