package design

import (
	. "goa.design/goa/v3/dsl"
)

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

	Method("post_manual_edge", func() {
		Description("Creates a manual relationship between two nodes.")
		Payload(ManualEdgeRequest)
		Result(GraphEdge)
		HTTP(func() {
			POST("/v1/graph/edge")
			Response(StatusCreated)
			Response("bad_request", StatusBadRequest)
		})
	})
})

var SubgraphRequest = Type("SubgraphRequest", func() {
	Description("Parameters for extracting a localized network subgraph.")
	Attribute("root", func() {
		Description("The starting node for the traversal.")
		Attribute("type", String, "Type of the root node (usually USER).", func() { Example("USER") })
		Attribute("key", String, "The unique key of the root node.", func() { Example("u_123") })
		Required("type", "key")
	})
	Attribute("hops", Int, "Number of hops to traverse (>=1).", func() { Default(2); Minimum(1); Example(2) })
	Attribute("edge_types", ArrayOf(String), "Filter to only include these relationship types.", func() { Example([]string{"PAYMENT", "LOGIN"}) })
	Attribute("min_event_count", Int, "Only include edges with at least this event_count. Set to 0 to disable.", func() {
		Default(0)
		Minimum(0)
		Example(2)
	})
	Attribute("time_window_ms", Int64, "Only include edges observed within the last N milliseconds. Omit or set to 0 for all time.", func() {
		Default(0)
		Minimum(0)
		Example(int64(2592000000))
	})
	Attribute("limit", func() {
		Description("Resource budget for the response.")
		Attribute("max_nodes", Int, "Maximum number of nodes to return.", func() { Default(100); Example(50) })
		Attribute("max_edges", Int, "Maximum number of edges to return.", func() { Default(200); Example(100) })
		Required("max_nodes", "max_edges")
	})
	Required("root", "limit")
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
	Description("A relationship between two entities.")
	Attribute("id", String, "Unique ID for the specific relationship.", func() { Example("e123") })
	Attribute("type", String, "The type of connection (e.g. PAYMENT).", func() { Example("PAYMENT") })
	Attribute("from", String, "ID of the source node.", func() { Example("USER:u_123") })
	Attribute("to", String, "ID of the target node.", func() { Example("MERCHANT:m_777") })
	Attribute("directed", Boolean, "Whether the relationship has a specific flow direction.", func() { Example(true) })
	Attribute("manual", Boolean, "Whether the relationship was manually added.", func() { Example(false) })
	Required("id", "type", "from", "to", "directed", "manual")
})

var NodeRef = Type("NodeRef", func() {
	Description("A reference to a specific node in the graph.")
	Attribute("type", String, "Type of the node.", func() { Example("USER") })
	Attribute("key", String, "The unique key of the node.", func() { Example("u_123") })
	Required("type", "key")
})

var ManualEdgeRequest = Type("ManualEdgeRequest", func() {
	Description("Defines a manually created relationship between two nodes.")
	Attribute("from", NodeRef, "Source node.")
	Attribute("to", NodeRef, "Target node.")
	Attribute("edge_type", String, "Relationship type (e.g. PAYMENT, MANUAL).", func() { Example("MANUAL") })
	Required("from", "to", "edge_type")
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
	Required("node_types", "edge_types")
})
