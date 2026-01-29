package ingest

import (
	"testing"
	"time"

	"github.com/aditnikel/grapgraph/src/model"
)

func strPtr(s string) *string {
	return &s
}

func TestChooseTargetPrecedence(t *testing.T) {
	cases := []struct {
		name    string
		ev      model.CustomerEvent
		label   string
		keyProp string
		key     string
		ok      bool
	}{
		{
			name: "merchant_overrides_others",
			ev: model.CustomerEvent{
				MerchantIDMPAN: strPtr("m1"),
				Exchange:       strPtr("ex1"),
				WalletAddress:  strPtr("w1"),
				PaymentMethod:  strPtr("pm1"),
				IssuingBank:    strPtr("b1"),
				DeviceID:       strPtr("d1"),
			},
			label:   "Merchant",
			keyProp: "merchant_id_mpan",
			key:     "m1",
			ok:      true,
		},
		{
			name: "exchange_when_only_exchange",
			ev: model.CustomerEvent{
				Exchange: strPtr("ex1"),
			},
			label:   "Exchange",
			keyProp: "exchange",
			key:     "ex1",
			ok:      true,
		},
		{
			name: "wallet_when_only_wallet",
			ev: model.CustomerEvent{
				WalletAddress: strPtr("w1"),
			},
			label:   "Wallet",
			keyProp: "wallet_address",
			key:     "w1",
			ok:      true,
		},
		{
			name: "payment_method_when_only_payment_method",
			ev: model.CustomerEvent{
				PaymentMethod: strPtr("pm1"),
			},
			label:   "PaymentMethod",
			keyProp: "payment_method",
			key:     "pm1",
			ok:      true,
		},
		{
			name: "issuing_bank_when_only_issuing_bank",
			ev: model.CustomerEvent{
				IssuingBank: strPtr("b1"),
			},
			label:   "Bank",
			keyProp: "issuing_bank",
			key:     "b1",
			ok:      true,
		},
		{
			name: "device_when_only_device",
			ev: model.CustomerEvent{
				DeviceID: strPtr("d1"),
			},
			label:   "Device",
			keyProp: "device_id",
			key:     "d1",
			ok:      true,
		},
		{
			name:  "no_match",
			ev:    model.CustomerEvent{},
			label: "",
			ok:    false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			label, keyProp, keyValue, ok := ChooseTarget(tc.ev)
			if ok != tc.ok {
				t.Fatalf("unexpected ok: %v", ok)
			}
			if label != tc.label || keyProp != tc.keyProp || keyValue != tc.key {
				t.Fatalf("unexpected target: %s %s %s", label, keyProp, keyValue)
			}
		})
	}
}

func TestParseEventTimestamp(t *testing.T) {
	base := time.Date(2024, 1, 2, 3, 4, 5, 0, time.UTC)
	want := base.UnixMilli()

	cases := []struct {
		name    string
		in      any
		want    int64
		wantErr bool
	}{
		{"float64", float64(want), want, false},
		{"int64", int64(want), want, false},
		{"int", int(want), want, false},
		{"rfc3339", base.Format(time.RFC3339), want, false},
		{"invalid_rfc3339", "not-a-date", 0, true},
		{"unsupported_type", true, 0, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ParseEventTimestamp(tc.in)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Fatalf("got %d want %d", got, tc.want)
			}
		})
	}
}
