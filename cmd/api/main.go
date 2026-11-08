package main

import (
	"context"
	"errors"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	listingapp "go-prod-project/internal/application/listing"
	"go-prod-project/internal/config"
	"go-prod-project/internal/db"
	"go-prod-project/internal/infrastructure/postgres"
	"go-prod-project/internal/server"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	// 1. Load application configuration.
	cfg := config.MustLoad()

	// 2. Initialize logger.
	logger := slog.New(
		slog.NewJSONHandler(os.Stdout, nil),
	)
	slog.SetDefault(logger)

	// 3. Connect to PostgreSQL (Database Connection Pool).
	database, err := db.Connect(cfg.DatabaseUrl)
	if err != nil {
		return err
	}
	defer database.Close()

	// 4. [DDD Infrastructure Layer] Initialize secondary adapter (PostgreSQL repository).
	listingRepository := postgres.NewListingRepository(database)

	// 5. [DDD Application Layer] Initialize application service with the repository port.
	listingService := listingapp.NewService(listingRepository)

	// 6. [DDD Presentation Layer] Initialize primary adapter (HTTP REST handler).
	listingHandler := server.NewListingHandler(listingService, logger)

	// 7. [Composition Root] Initialize HTTP router and routes.
	router := server.NewRouter(listingHandler)

	// 8. Initialize HTTP server.
	addr := os.Getenv("HTTP_ADDR")
	if addr == "" {
		addr = ":" + cfg.Port
	}

	httpServer := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// 9. Handle OS shutdown signals for graceful shutdown.
	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	serverErrors := make(chan error, 1)

	go func() {
		logger.Info("HTTP server starting", "address", addr)

		serverErrors <- httpServer.ListenAndServe()
	}()

	// 10. Wait for server failure or shutdown signal.
	select {
	case err := <-serverErrors:
		if !errors.Is(err, http.ErrServerClosed) {
			return err
		}

	case <-ctx.Done():
		logger.Info("shutdown signal received")

		shutdownCtx, cancel := context.WithTimeout(
			context.Background(),
			10*time.Second,
		)
		defer cancel()

		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			_ = httpServer.Close()
			return err
		}

		err := <-serverErrors
		if !errors.Is(err, http.ErrServerClosed) {
			return err
		}

		logger.Info("HTTP server stopped gracefully")
	}

	return nil
}
