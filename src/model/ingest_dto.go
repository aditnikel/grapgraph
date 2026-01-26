package model

type CustomerEvent struct {
	UserID         string  `json:"user_id"`
	MerchantIDMPAN *string `json:"merchant_id_mpan,omitempty"`
	EventType      string  `json:"event_type"`
	EventTimestamp any     `json:"event_timestamp"` // RFC3339 string OR epoch ms number

	TotalAmount   *float64 `json:"total_transaction_amount,omitempty"`
	DeviceID      *string  `json:"device_id,omitempty"`
	PaymentMethod *string  `json:"payment_method,omitempty"`
	IssuingBank   *string  `json:"issuing_bank,omitempty"`
	WalletAddress *string  `json:"wallet_address,omitempty"`
	Exchange      *string  `json:"exchange,omitempty"`

	IPAddress *string `json:"ip_address,omitempty"` // never stored as node; raw omitted from API response
}
