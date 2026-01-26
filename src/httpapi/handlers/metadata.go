package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/aditnikel/grapgraph/src/model"
)

type MetadataHandler struct{}

func (h *MetadataHandler) Get(w http.ResponseWriter, r *http.Request) {
	ets := make([]string, 0, len(model.AllEventTypes()))
	for _, e := range model.AllEventTypes() {
		ets = append(ets, string(e))
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"node_types": []string{
			string(model.NodeUser),
			string(model.NodeMerchant),
			string(model.NodeExchange),
			string(model.NodeWallet),
			string(model.NodePaymentMethod),
			string(model.NodeBank),
			string(model.NodeDevice),
		},
		"edge_types": ets,
		"rank_metrics": []string{
			"event_count_30d",
			"event_count",
			"total_amount",
		},
	})
}
