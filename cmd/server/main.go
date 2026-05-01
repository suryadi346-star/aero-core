package main

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	_ "modernc.org/sqlite"
	"github.com/suryadi346-star/aero-core/internal/api"
	"github.com/suryadi346-star/aero-core/internal/cache"
	"github.com/suryadi346-star/aero-core/internal/config"
	"github.com/suryadi346-star/aero-core/internal/model"
	"github.com/suryadi346-star/aero-core/internal/session"
)

func main() {
	cfgPath := os.Getenv("CONFIG_PATH")
	if cfgPath == "" { cfgPath = "configs/app.yaml" }

	cfg, err := config.Load(cfgPath)
	if err != nil {
		slog.Error("Failed to load config", "error", err)
		os.Exit(1)
	}

	runtime.SetMemoryLimit(int64(cfg.System.MemoryLimitMB) << 20)
	slog.Info("Starting AeroCore AI", "port", cfg.Server.Port, "ram_limit_mb", cfg.System.MemoryLimitMB)

	db, err := initSQLite(cfg)
	if err != nil {
		slog.Error("SQLite init failed", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
		Handler:      setupRoutes(db, cfg),
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("HTTP server error", "error", err)
		}
	}()

	<-sigChan
	slog.Info("Shutdown signal received")

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		slog.Error("Graceful shutdown failed", "error", err)
	}
	slog.Info("Server stopped")
}

func initSQLite(cfg *config.AppConfig) (*sql.DB, error) {
	dsn := fmt.Sprintf("file:%s?_journal_mode=WAL&_busy_timeout=5000&_foreign_keys=ON", cfg.Cache.SQLitePath)
	db, err := sql.Open("sqlite", dsn)
	if err != nil { return nil, err }
	db.SetMaxOpenConns(4)
	db.SetMaxIdleConns(2)
	db.SetConnMaxLifetime(10 * time.Minute)
	if err := db.Ping(); err != nil { return nil, fmt.Errorf("ping failed: %w", err) }
	return db, nil
}

func setupRoutes(db *sql.DB, cfg *config.AppConfig) http.Handler {
	mux := http.NewServeMux()
	ollamaURL := fmt.Sprintf("http://%s", cfg.System.OllamaHost)
	ollamaClient := model.NewOllamaClient(ollamaURL)
	sessStore := session.NewStore(time.Duration(cfg.Cache.TTLMinutes) * time.Minute)
	sqlCache, err := cache.NewSQLiteCache(db)
	if err != nil {
		slog.Error("Cache init failed", "error", err)
		os.Exit(1)
	}
	chatHandler := api.NewChatHandler(ollamaClient, sessStore, sqlCache)

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"status":"ok"}`)
	})
	mux.HandleFunc("/health/deep", api.HealthDeep(db, ollamaURL))
	mux.HandleFunc("/metrics", api.MetricsHandler)
	mux.HandleFunc("/chat/stream", chatHandler.Stream)

	rl := api.NewRateLimiter(10, time.Second)
	return api.LoggingMiddleware(api.RecoveryMiddleware(rl.Middleware(mux)))
}
