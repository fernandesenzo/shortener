package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/fernandesenzo/shortener/internal/auth"
	"github.com/fernandesenzo/shortener/internal/jwt"
	platform "github.com/fernandesenzo/shortener/internal/platform/cache"
	"github.com/fernandesenzo/shortener/internal/platform/postgres"
	"github.com/fernandesenzo/shortener/internal/shortener"
	"github.com/fernandesenzo/shortener/internal/user"
	"github.com/joho/godotenv"
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
	redisURL := os.Getenv("REDIS_URL")
	if dbURL == "" || redisURL == "" {
		return errors.New("DATABASE_URL and REDIS_URL must be set")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	db, err := postgres.NewConnection(dbURL)
	if err != nil {
		slog.Error("postgres connection failed", "error", err)
		return err
	}
	defer func() {
		if err := db.Close(); err != nil {
			slog.Error("failed to close postgres", "error", err)
		} else {
			slog.Info("postgres connection closed gracefully")
		}
	}()

	if err = postgres.RunMigrations(db); err != nil {
		slog.Error("failed to run migrations", "error", err)
		return err
	}

	redisClient, err := platform.NewRedisClient(redisURL)
	if err != nil {
		slog.Error("redis connection failed", "error", err)
		return err
	}
	defer func() {
		if err := redisClient.Close(); err != nil {
			slog.Error("failed to close redis", "error", err)
		} else {
			slog.Info("redis connection closed gracefully")
		}
	}()

	slog.Info("infrastructure connected")

	pgRepo := shortener.NewPostgresRepository(db)
	redisRepo := shortener.NewRedisRepository(redisClient)
	repo := shortener.NewHybridLinkRepository(pgRepo, redisRepo)
	service := shortener.NewService(repo)
	handler := shortener.NewHandler(service)

	pgRepoUser := user.NewPostgresRepository(db)
	serviceUser := user.NewService(pgRepoUser)
	handlerUser := user.NewHandler(serviceUser)

	jwtManager := jwt.NewManager(os.Getenv("JWT_SECRET_KEY"), time.Hour)
	pgRepoAuth := auth.NewPostgresRepository(db)
	serviceAuth := auth.NewService(pgRepoAuth, jwtManager)
	handlerAuth := auth.NewHandler(serviceAuth)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/links", handler.Shorten)
	mux.HandleFunc("GET /{code}", handler.Get)
	mux.HandleFunc("POST /api/users", handlerUser.Create)
	mux.HandleFunc("POST /api/login", handlerAuth.Login)
	mux.Handle("DELETE /api/links/{code}", RequireAuthMiddleware(http.HandlerFunc(handler.Delete)))

	handlerStack := AuthMiddleware(mux, jwtManager)
	handlerStack = RateLimitMiddleware(handlerStack, redisClient, 10, time.Hour)
	handlerStack = CORSMiddleware(handlerStack)
	handlerStack = RecoverMiddleware(handlerStack)
	handlerStack = LoggingMiddleware(handlerStack)

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      handlerStack,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	serverErrors := make(chan error, 1)

	go func() {
		slog.Info("server starting", "port", port)
		serverErrors <- srv.ListenAndServe()
	}()

	select {
	case err := <-serverErrors:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("server error: %w", err)
		}
	case <-ctx.Done():
		slog.Info("shutting down OS signal received")

		shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancelShutdown()

		if err := srv.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("graceful shutdown failed: %w", err)
		}
	}

	slog.Info("server stopped")
	return nil
}
