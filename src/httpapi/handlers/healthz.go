package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/aditnikel/grapgraph/src/observability"
	"github.com/aditnikel/grapgraph/src/service"
)

type HealthzHandler struct {
	Log   *observability.Logger
	Graph *service.GraphService
}

func (h *HealthzHandler) Get(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if err := h.Graph.Ping(r.Context()); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"ok":    false,
			"error": err.Error(),
		})
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]any{"ok": true})
}
