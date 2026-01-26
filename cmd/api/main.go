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

	"github.com/aditnikel/grapgraph/src/config"
	"github.com/aditnikel/grapgraph/src/graph"
	"github.com/aditnikel/grapgraph/src/httpapi"
	"github.com/aditnikel/grapgraph/src/httpapi/handlers"
	"github.com/aditnikel/grapgraph/src/observability"
	"github.com/aditnikel/grapgraph/src/service"
)

func main() {
	// Local-dev convenience; production should inject env vars via runtime.
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

	repo := graph.New(rdb, cfg.GraphName, cfg.DBTimeout)
	repo.EnsureSchema(context.Background())

	graphSvc := &service.GraphService{
		Repo: repo,
		Cfg:  cfg,
	}
	ingestSvc := &service.IngestService{Repo: repo}

	router := httpapi.Router(httpapi.Deps{
		Log:      log,
		Graph:    &handlers.GraphHandler{Graph: graphSvc},
		Ingest:   &handlers.IngestHandler{Ingest: ingestSvc},
		Metadata: &handlers.MetadataHandler{},
		Healthz:  &handlers.HealthzHandler{Log: log, Graph: graphSvc},
	})

	srv := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Info("server_start", observability.Fields{"addr": cfg.HTTPAddr})
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("server_error", observability.Fields{"err": err.Error()})
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	log.Info("server_shutdown", nil)
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)
}
