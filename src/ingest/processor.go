package ingest

import (
	"fmt"
	"time"

	"github.com/aditnikel/grapgraph/src/model"
)

// Choose target entity by precedence.
// Returns (label, keyProp, keyValue, ok)
func ChooseTarget(ev model.CustomerEvent) (string, string, string, bool) {
	if ev.MerchantIDMPAN != nil && *ev.MerchantIDMPAN != "" {
		return "Merchant", "merchant_id_mpan", *ev.MerchantIDMPAN, true
	}
	if ev.Exchange != nil && *ev.Exchange != "" {
		return "Exchange", "exchange", *ev.Exchange, true
	}
	if ev.WalletAddress != nil && *ev.WalletAddress != "" {
		return "Wallet", "wallet_address", *ev.WalletAddress, true
	}
	if ev.PaymentMethod != nil && *ev.PaymentMethod != "" {
		return "PaymentMethod", "payment_method", *ev.PaymentMethod, true
	}
	if ev.IssuingBank != nil && *ev.IssuingBank != "" {
		return "Bank", "issuing_bank", *ev.IssuingBank, true
	}
	if ev.DeviceID != nil && *ev.DeviceID != "" {
		return "Device", "device_id", *ev.DeviceID, true
	}
	return "", "", "", false
}

// Accept RFC3339 string or epoch ms number.
func ParseEventTimestamp(ts any) (int64, error) {
	switch v := ts.(type) {
	case float64:
		return int64(v), nil
	case int64:
		return v, nil
	case int:
		return int64(v), nil
	case string:
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			return 0, fmt.Errorf("invalid RFC3339 timestamp")
		}
		return t.UnixMilli(), nil
	default:
		return 0, fmt.Errorf("event_timestamp must be RFC3339 string or epoch ms number")
	}
}
