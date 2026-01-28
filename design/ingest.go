package design

import (
	. "goa.design/goa/v3/dsl"
)

var _ = Service("ingest", func() {
	Description("High-speed financial event ingestion service.")
	Error("bad_request", String, "Error returned when the request payload is malformed or invalid.")

	Method("post_event", func() {
		Description("Accepts one or more financial events (Payment, Login, etc.) and updates the relationship graph.")
		Payload(BulkCustomerEvents)
		Result(BulkIngestResponse)
		HTTP(func() {
			POST("/v1/ingest/event")
			Response(StatusAccepted)
			Response("bad_request", StatusBadRequest)
		})
	})
})

var BulkIngestResponse = Type("BulkIngestResponse", func() {
	Description("Result of the bulk ingestion attempt.")
	Attribute("accepted", Boolean, "Whether all events were processed successfully.", func() { Example(true) })
	Attribute("accepted_count", Int, "Number of events accepted in this batch.", func() { Example(3) })
	Attribute("failed_count", Int, "Number of events rejected in this batch.", func() { Example(0) })
	Required("accepted", "accepted_count", "failed_count")
})

// Deprecated: IngestResponse is replaced by BulkIngestResponse now that post_event supports batches.
var IngestResponse = Type("IngestResponse", func() {
	Description("Result of the event ingestion attempt.")
	Attribute("accepted", Boolean, "Whether the event was successfully queued or processed.", func() { Example(true) })
	Required("accepted")
})

var CustomerEvent = Type("CustomerEvent", func() {
	Description("Information about a financial activity or user action.")
	Attribute("user_id", String, "Unique identifier of the user (e.g. u_123).", func() { Example("u_123") })
	Attribute("merchant_id_mpan", String, "Target merchant ID or card MPAN.", func() { Example("m_777") })
	Attribute("event_type", String, "The type of event (PAYMENT, LOGIN, WITHDRAWAL, etc).", func() { Example("PAYMENT") })
	Attribute("event_timestamp", Any, "Timestamp of the activity (RFC3339 string or Epoch MS).", func() { Example("2024-03-20T10:00:00Z") })
	Attribute("total_transaction_amount", Float64, "Monetary value of the transaction.", func() { Example(150.50) })
	Attribute("device_id", String, "Unique hardware ID where the activity originated.", func() { Example("d_888") })
	Attribute("payment_method", String, "Method used (VISA, CRYPTO, etc).", func() { Example("VISA") })
	Attribute("issuing_bank", String, "The bank that issued the instrument.", func() { Example("JP_MORGAN") })
	Attribute("wallet_address", String, "Blockchain wallet address if applicable.", func() { Example("0xabc123") })
	Attribute("exchange", String, "Crypto exchange name if applicable.", func() { Example("BINANCE") })
	Attribute("ip_address", String, "Remote IP address (not stored directly in graph).", func() { Example("192.168.1.1") })
	Required("user_id", "event_type", "event_timestamp")
})

var BulkCustomerEvents = Type("BulkCustomerEvents", func() {
	Description("Batch of financial events for ingestion.")
	Attribute("events", ArrayOf(CustomerEvent), "List of events to ingest in-order.", func() {
		MinLength(1)
		Example([]any{
			map[string]any{"user_id": "u_1", "event_type": "PAYMENT", "event_timestamp": "2024-03-20T10:00:00Z", "total_transaction_amount": 125.0, "merchant_id_mpan": "m_7"},
			map[string]any{"user_id": "u_2", "event_type": "LOGIN", "event_timestamp": 1710930030000},
		})
	})
	Required("events")
})
