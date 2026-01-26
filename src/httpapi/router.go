package httpapi

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/aditnikel/grapgraph/src/httpapi/handlers"
	"github.com/aditnikel/grapgraph/src/httpapi/middleware"
	"github.com/aditnikel/grapgraph/src/observability"
)

type Deps struct {
	Log      *observability.Logger
	Graph    *handlers.GraphHandler
	Ingest   *handlers.IngestHandler
	Metadata *handlers.MetadataHandler
	Healthz  *handlers.HealthzHandler
}

func Router(d Deps) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.NewRecoverer(d.Log))
	r.Use(middleware.NewRequestLog(d.Log))

	r.Get("/healthz", d.Healthz.Get)

	r.Route("/v1", func(r chi.Router) {
		r.Post("/ingest/event", d.Ingest.PostEvent)
		r.Get("/graph/metadata", d.Metadata.Get)
		r.Post("/graph/subgraph", d.Graph.PostSubgraph)
	})

	return r
}
