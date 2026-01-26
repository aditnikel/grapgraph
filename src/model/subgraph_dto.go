package model

type SubgraphRequest struct {
	Root struct {
		Type string `json:"type"`
		Key  string `json:"key"`
	} `json:"root"`

	Hops int `json:"hops"`

	TimeWindow struct {
		From string `json:"from"`
		To   string `json:"to"`
	} `json:"time_window"`

	EdgeTypes       []string `json:"edge_types"`
	MinEventCount   int      `json:"min_event_count"`
	RankNeighborsBy string   `json:"rank_neighbors_by"`

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
