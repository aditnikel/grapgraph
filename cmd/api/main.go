package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/redis/rueidis"
	goahttp "goa.design/goa/v3/http"

	"github.com/aditnikel/grapgraph/gen/graph"
	"github.com/aditnikel/grapgraph/gen/health"
	graphsvr "github.com/aditnikel/grapgraph/gen/http/graph/server"
	healthsvr "github.com/aditnikel/grapgraph/gen/http/health/server"
	ingestsvr "github.com/aditnikel/grapgraph/gen/http/ingest/server"
	openapisvr "github.com/aditnikel/grapgraph/gen/http/openapi/server"
	"github.com/aditnikel/grapgraph/gen/ingest"
	"github.com/aditnikel/grapgraph/gen/openapi"
	"github.com/aditnikel/grapgraph/src/config"
	"github.com/aditnikel/grapgraph/src/domain"
	repo "github.com/aditnikel/grapgraph/src/graph"
	custmid "github.com/aditnikel/grapgraph/src/httpapi/middleware"
	"github.com/aditnikel/grapgraph/src/observability"
	goa_services "github.com/aditnikel/grapgraph/src/services"
)

func main() {
	_ = godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}

	log := observability.New(cfg.LogLevel)

	rdb, err := rueidis.NewClient(rueidis.ClientOption{
		InitAddress: cfg.RedisAddrs,
		Password:    cfg.RedisPassword,
	})
	if err != nil {
		panic(err)
	}
	defer rdb.Close()

	gRepo := repo.New(rdb, cfg.GraphName, cfg.DBTimeout)
	gRepo.EnsureSchema(context.Background())

	// Initialize domain services
	graphSvcBase := &domain.GraphService{Repo: gRepo, Cfg: cfg}
	ingestSvcBase := &domain.IngestService{Repo: gRepo}

	// Initialize Goa service wrappers
	handler := buildHandler(log, graphSvcBase, ingestSvcBase)

	srv := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Info("server_start", observability.Fields{"addr": cfg.HTTPAddr, "framework": "goa"})
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("server_error", observability.Fields{"err": err.Error()})
			os.Exit(1)
		}
	}()

	handleGracefulShutdown(log, srv)
}

func buildHandler(log *observability.Logger, graphSvcBase *domain.GraphService, ingestSvcBase *domain.IngestService) http.Handler {
	mux := goahttp.NewMuxer()
	dec := goahttp.RequestDecoder
	enc := goahttp.ResponseEncoder

	// Goa Services
	healthSvc := &goa_services.HealthService{Log: log, Graph: graphSvcBase}
	ingestSvc := &goa_services.IngestService{Ingest: ingestSvcBase}
	graphSvc := &goa_services.GraphService{Graph: graphSvcBase}
	openapiSvc := &goa_services.OpenapiService{}

	// Goa Endpoints
	healthEndpoints := health.NewEndpoints(healthSvc)
	ingestEndpoints := ingest.NewEndpoints(ingestSvc)
	graphEndpoints := graph.NewEndpoints(graphSvc)
	openapiEndpoints := openapi.NewEndpoints(openapiSvc)

	// Goa HTTP Servers
	healthServer := healthsvr.New(healthEndpoints, mux, dec, enc, nil, nil)
	ingestServer := ingestsvr.New(ingestEndpoints, mux, dec, enc, nil, nil)
	graphServer := graphsvr.New(graphEndpoints, mux, dec, enc, nil, nil)
	openapiServer := openapisvr.New(openapiEndpoints, mux, dec, enc, nil, nil, nil)

	// Mount servers

	healthsvr.Mount(mux, healthServer)
	ingestsvr.Mount(mux, ingestServer)
	graphsvr.Mount(mux, graphServer)
	openapisvr.Mount(mux, openapiServer)

	// Apply CORS
	return custmid.CORS(mux)
}

func handleGracefulShutdown(log *observability.Logger, srv *http.Server) {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	log.Info("server_shutdown", nil)
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error("shutdown_error", observability.Fields{"err": err.Error()})
	}
}
