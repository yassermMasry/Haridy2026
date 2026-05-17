package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"haridy2026/configs"
	"haridy2026/internal/database"
	"haridy2026/internal/jobs"
	"haridy2026/internal/routes"
)

func main() {
	cfg := configs.Load()
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: cfg.LogLevel()}))
	slog.SetDefault(logger)

	db, err := configs.ConnectDatabase(cfg)
	if err != nil {
		logger.Error("database connection failed", "error", err)
		os.Exit(1)
	}

	if err := database.AutoMigrate(db); err != nil {
		logger.Error("database migration failed", "error", err)
		os.Exit(1)
	}
	if err := database.Seed(db); err != nil {
		logger.Error("database seed failed", "error", err)
		os.Exit(1)
	}

	jobCtx, stopJobs := context.WithCancel(context.Background())
	defer stopJobs()
	jobs.Start(jobCtx, db)

	router := routes.Setup(db, cfg)
	server := &http.Server{
		Addr:         ":" + cfg.AppPort,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 20 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		logger.Info("server started", "addr", server.Addr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("server failed", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	logger.Info("shutting down server")
	if err := server.Shutdown(ctx); err != nil {
		logger.Error("graceful shutdown failed", "error", err)
	}
}
