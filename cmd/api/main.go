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

	"github.com/fernandesenzo/shortener/internal/jobs/linkprune"
	"github.com/joho/godotenv"

	"github.com/fernandesenzo/shortener/internal/platform/postgres"
	"github.com/fernandesenzo/shortener/internal/shortener"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	if err := run(); err != nil {
		slog.Error("application startup failed", "error", err)
		os.Exit(1)
	}
}

func run() error {
	if err := godotenv.Load(); err != nil {
		slog.Warn("cannot load .env file. assuming env variables are set")
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		return errors.New("could not get database URL from environment variable DATABASE_URL")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	db, err := postgres.NewConnection(dbURL)
	if err != nil {
		return err
	}

	defer func() {
		if err := db.Close(); err != nil {
			slog.Error("failed to close database", "error", err)
		}
	}()

	slog.Info("connected to db successfully")

	repo := shortener.NewPostgresRepository(db)

	// activating job
	expirationRule := time.Hour * 24
	jobInterval := time.Minute * 10
	prunerJob := linkprune.NewJob(repo, jobInterval, expirationRule)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	slog.Info("starting link pruner job")
	go prunerJob.Run(ctx)

	service := shortener.NewService(repo)
	handler := shortener.NewHandler(service)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/links", handler.Shorten)
	mux.HandleFunc("GET /{code}", handler.Get)

	rateLimiter := NewRateLimiter(100, 200, 5, 10, 60)

	ratedMux := rateLimiter.Middleware(mux)

	loggedMux := LoggingMiddleware(ratedMux)

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      loggedMux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	shutdownError := make(chan error)

	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

		s := <-quit
		slog.Info("shutting down server", "signal", s.String())

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		shutdownError <- srv.Shutdown(ctx)
	}()

	slog.Info("server starting", "port", port)

	err = srv.ListenAndServe()

	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	err = <-shutdownError
	if err != nil {
		return err
	}

	slog.Info("server stopped gracefully")
	return nil
}
