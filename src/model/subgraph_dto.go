package model

type SubgraphRequest struct {
	Root struct {
		Type string `json:"type"`
		Key  string `json:"key"`
	} `json:"root"`

	Hops int `json:"hops"`

	EdgeTypes []string `json:"edge_types"`

	MinEventCount int `json:"min_event_count"`

	TimeWindowMs int64 `json:"time_window_ms"`

	Limit struct {
		MaxNodes int `json:"max_nodes"`
		MaxEdges int `json:"max_edges"`
	} `json:"limit"`
}

type SubgraphResponse struct {
	Version   string      `json:"version"`
	Root      string      `json:"root"`
	Nodes     []GraphNode `json:"nodes"`
	Edges     []GraphEdge `json:"edges"`
	Truncated bool        `json:"truncated"`
}

type ManualEdgeRequest struct {
	From struct {
		Type string `json:"type"`
		Key  string `json:"key"`
	} `json:"from"`
	To struct {
		Type string `json:"type"`
		Key  string `json:"key"`
	} `json:"to"`
	EdgeType string `json:"edge_type"`
}
