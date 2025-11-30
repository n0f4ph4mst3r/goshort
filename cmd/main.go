package main

import (
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/n0f4ph4mst3r/goshort/internal/config"
	"github.com/n0f4ph4mst3r/goshort/internal/http-server/handlers/erase"
	"github.com/n0f4ph4mst3r/goshort/internal/http-server/handlers/redirect"
	"github.com/n0f4ph4mst3r/goshort/internal/http-server/handlers/save"
	"github.com/n0f4ph4mst3r/goshort/internal/http-server/mwlogger"
	"github.com/n0f4ph4mst3r/goshort/internal/storage"
	"github.com/n0f4ph4mst3r/goshort/internal/storage/postgres"
	rds "github.com/n0f4ph4mst3r/goshort/internal/storage/redis"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("PWD:", dir)

	cfg, dbUrl, cacheUrl := config.MustLoad()
	log := setupLogger(cfg.Env)

	log.Info("Starting application...", slog.String("env", cfg.Env))
	log.Debug("Debugging is enabled")

	postgres, err := postgres.New(dbUrl)
	if err != nil {
		log.Error("Failed to initialize storage", "err", err)
		os.Exit(1)
	}

	rdsStorage, err := rds.New(cacheUrl, &cfg.Cache)
	if err != nil {
		log.Warn(err.Error())
	}

	url_storage := storage.New(log, postgres, rdsStorage)

	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.Logger)
	router.Use(mwlogger.New(log))
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)

	router.Route("/api", func(api_routes chi.Router) {
		api_routes.Get("/url/{alias}", redirect.New(log, url_storage))

		api_routes.Route("/url", func(auth_routes chi.Router) {
			auth_routes.Use(middleware.BasicAuth("goshort", map[string]string{
				cfg.HTTPServer.User: cfg.HTTPServer.Password,
			}))

			auth_routes.Post("/", save.New(log, url_storage, nil))
			auth_routes.Delete("/{alias}", erase.New(log, url_storage))
		})
	})

	log.Info("starting server", slog.String("address", cfg.Address+":"+fmt.Sprint(cfg.Port)))

	srv := &http.Server{
		Addr:         cfg.Address + ":" + fmt.Sprint(cfg.Port),
		Handler:      router,
		ReadTimeout:  cfg.HTTPServer.Timeout,
		WriteTimeout: cfg.HTTPServer.Timeout,
		IdleTimeout:  cfg.HTTPServer.IdleTimeout,
	}

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Error("failed to start server", slog.Any("err", err))
		os.Exit(1)
	}
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case envDev:
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case envProd:
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	}

	return log
}
