package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/aditnikel/grapgraph/src/model"
	"github.com/aditnikel/grapgraph/src/service"
)

type IngestHandler struct {
	Ingest *service.IngestService
}

func (h *IngestHandler) PostEvent(w http.ResponseWriter, r *http.Request) {
	var ev model.CustomerEvent
	if err := json.NewDecoder(r.Body).Decode(&ev); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	if err := h.Ingest.AcceptEvent(r.Context(), ev); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	if err := json.NewEncoder(w).Encode(map[string]any{"accepted": true}); err != nil {
		http.Error(w, "encoding error", http.StatusInternalServerError)
	}
}
