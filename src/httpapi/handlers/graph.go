package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/aditnikel/grapgraph/src/model"
	"github.com/aditnikel/grapgraph/src/service"
)

type GraphHandler struct {
	Graph *service.GraphService
}

func (h *GraphHandler) PostSubgraph(w http.ResponseWriter, r *http.Request) {
	var req model.SubgraphRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	resp, err := h.Graph.Subgraph(r.Context(), req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "encoding error", http.StatusInternalServerError)
	}
}
