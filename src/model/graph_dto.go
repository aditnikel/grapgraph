package model

type GraphNode struct {
	ID    string         `json:"id"`
	Type  string         `json:"type"`
	Key   string         `json:"key"`
	Label string         `json:"label"`
	Props map[string]any `json:"props,omitempty"`
}

type GraphEdge struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	From     string `json:"from"`
	To       string `json:"to"`
	Directed bool   `json:"directed"`
	Manual   bool   `json:"manual"`
}

type MetadataResponse struct {
	NodeTypes []string `json:"node_types"`
	EdgeTypes []string `json:"edge_types"`
}
