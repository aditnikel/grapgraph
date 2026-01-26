package design

import (
	. "goa.design/goa/v3/dsl"
)

var _ = API("grapgraph", func() {
	Title("Grapgraph API")
	Description("High-performance fraud detection and financial network analysis platform built with Goa and FalkorDB.")
	Server("grapgraph", func() {
		Host("localhost", func() {
			URI("http://localhost:8080")
		})
	})
})

var _ = Service("openapi", func() {
	Description("The openapi service serves the OpenAPI specification and interactive documentation.")
	HTTP(func() {
		Path("/")
	})

	Method("index", func() {
		Description("Provides a simple landing page.")
		Payload(Empty)
		Result(String)
		HTTP(func() {
			GET("/")
			Response(StatusOK)
		})
	})

	Method("docs", func() {
		Description("Serves the interactive Swagger UI.")
		Payload(Empty)
		Result(String)
		HTTP(func() {
			GET("/docs")
			Response(StatusOK, func() {
				ContentType("text/html")
			})
		})
	})

	Files("/openapi.json", "gen/http/openapi3.json")
})

var _ = Service("health", func() {
	Description("Health check service for monitoring service and database connectivity.")
	Method("get", func() {
		Description("Returns the health status of the API and its underlying graph database.")
		Payload(Empty)
		Result(HealthResponse)
		HTTP(func() {
			GET("/healthz")
			Response(StatusOK)
			Response(StatusServiceUnavailable)
		})
	})
})

var _ = Service("ingest", func() {
	Description("High-speed financial event ingestion service.")
	Error("bad_request", String, "Error returned when the request payload is malformed or invalid.")

	Method("post_event", func() {
		Description("Accepts a new financial event (Payment, Login, etc.) and updates the relationship graph.")
		Payload(CustomerEvent)
		Result(IngestResponse)
		HTTP(func() {
			POST("/v1/ingest/event")
			Response(StatusAccepted)
			Response("bad_request", StatusBadRequest)
		})
	})
})

var _ = Service("graph", func() {
	Description("Graph traversal service for fraud pattern analysis and subgraph extraction.")
	Error("bad_request", String, "Error returned when the traversal parameters or root node are invalid.")

	Method("get_metadata", func() {
		Description("Returns valid node types, edge types, and supported ranking metrics.")
		Payload(Empty)
		Result(MetadataResponse)
		HTTP(func() {
			GET("/v1/graph/metadata")
			Response(StatusOK)
		})
	})

	Method("post_subgraph", func() {
		Description("Extracts a surrounding subgraph for a specific root node using multi-hop analysis.")
		Payload(SubgraphRequest)
		Result(SubgraphResponse)
		HTTP(func() {
			POST("/v1/graph/subgraph")
			Response(StatusOK)
			Response("bad_request", StatusBadRequest)
		})
	})
})

var HealthResponse = Type("HealthResponse", func() {
	Description("Health status of the system components.")
	Attribute("ok", Boolean, "Whether the service is fully operational.", func() { Example(true) })
	Attribute("error", String, "Error message if the service is unhealthy.", func() { Example("Database connection failed") })
	Required("ok")
})

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

var SubgraphRequest = Type("SubgraphRequest", func() {
	Description("Parameters for extracting a localized network subgraph.")
	Attribute("root", func() {
		Description("The starting node for the traversal.")
		Attribute("type", String, "Type of the root node (usually USER).", func() { Example("USER") })
		Attribute("key", String, "The unique key of the root node.", func() { Example("u_123") })
		Required("type", "key")
	})
	Attribute("hops", Int, "Number of hops to traverse (1-3).", func() { Default(2); Minimum(1); Maximum(3); Example(2) })
	Attribute("time_window", func() {
		Description("Optional time range to filter relationship metrics.")
		Attribute("from", String, "Start of the window (RFC3339).", func() { Example("2024-01-01T00:00:00Z") })
		Attribute("to", String, "End of the window (RFC3339).", func() { Example("2024-12-31T23:59:59Z") })
		Required("from", "to")
	})
	Attribute("edge_types", ArrayOf(String), "Filter to only include these relationship types.", func() { Example([]string{"PAYMENT", "LOGIN"}) })
	Attribute("min_event_count", Int, "Minimum number of aggregate events to include an edge.", func() { Default(1); Example(2) })
	Attribute("rank_neighbors_by", String, "Metric used to sort and truncate neighbor nodes.", func() { Default("event_count_30d"); Example("total_amount") })
	Attribute("limit", func() {
		Description("Resource budget for the response.")
		Attribute("max_nodes", Int, "Maximum number of nodes to return.", func() { Default(100); Example(50) })
		Attribute("max_edges", Int, "Maximum number of edges to return.", func() { Default(200); Example(100) })
		Required("max_nodes", "max_edges")
	})
	Required("root", "time_window", "limit")
})

var GraphNode = Type("GraphNode", func() {
	Description("A single entity (User, Merchant, Device) in the resulting subgraph.")
	Attribute("id", String, "Stable ID generated for visualization.", func() { Example("USER:u_123") })
	Attribute("type", String, "The category of the entity.", func() { Example("USER") })
	Attribute("key", String, "The domain-specific key (e.g. u_123).", func() { Example("u_123") })
	Attribute("label", String, "Human-friendly display name.", func() { Example("User u_123") })
	Attribute("props", MapOf(String, Any), "Additional key-value properties.")
	Required("id", "type", "key", "label")
})

var GraphEdge = Type("GraphEdge", func() {
	Description("A relationship between two entities with aggregated metrics.")
	Attribute("id", String, "Unique ID for the specific relationship.", func() { Example("e123") })
	Attribute("type", String, "The type of connection (e.g. PAYMENT).", func() { Example("PAYMENT") })
	Attribute("from", String, "ID of the source node.", func() { Example("USER:u_123") })
	Attribute("to", String, "ID of the target node.", func() { Example("MERCHANT:m_777") })
	Attribute("directed", Boolean, "Whether the relationship has a specific flow direction.", func() { Example(true) })
	Attribute("metrics", MapOf(String, Any), "Statistical snapshots (count, amount, first_seen, etc).")
	Required("id", "type", "from", "to", "directed")
})

var SubgraphResponse = Type("SubgraphResponse", func() {
	Description("Result of the graph traversal containing the extracted network.")
	Attribute("version", String, "Format version of the response.", func() { Example("1.0") })
	Attribute("root", String, "The ID of the requested starting node.", func() { Example("USER:u_123") })
	Attribute("nodes", ArrayOf(GraphNode), "List of all entities in the network.")
	Attribute("edges", ArrayOf(GraphEdge), "List of all connections found.")
	Attribute("truncated", Boolean, "Indicates if the result was clipped by performance budgets.", func() { Example(false) })
	Required("version", "root", "nodes", "edges", "truncated")
})

var MetadataResponse = Type("MetadataResponse", func() {
	Description("Supported constants and schema definitions for the current system.")
	Attribute("node_types", ArrayOf(String), "All valid entity types.", func() { Example([]string{"USER", "MERCHANT", "DEVICE"}) })
	Attribute("edge_types", ArrayOf(String), "All valid event types.", func() { Example([]string{"PAYMENT", "LOGIN", "WITHDRAWAL"}) })
	Attribute("rank_metrics", ArrayOf(String), "Valid keys for the rank_neighbors_by parameter.", func() { Example([]string{"event_count", "total_amount"}) })
	Required("node_types", "edge_types", "rank_metrics")
})
