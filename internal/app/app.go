package app

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"todo-app/internal/auth"
	"todo-app/internal/auth/password"
	"todo-app/internal/auth/token"
	"todo-app/internal/config"
	httpserver "todo-app/internal/http"
	"todo-app/internal/redis"
	"todo-app/internal/storage/cached"
	"todo-app/internal/storage/postgres"
	"todo-app/internal/todo"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type App struct {
	server      *http.Server
	redisClient *redis.Client
	pool        *pgxpool.Pool
}

func New(cfg config.Config) (*App, error) {
	ctx := context.Background()

	application := &App{}

	// PostgreSQL migrations
	if err := postgres.RunMigrations(cfg.DatabaseURL); err != nil {
		return nil, fmt.Errorf("run migrations: %w", err)
	}

	// PostgreSQL connection pool
	pool, err := postgres.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("create postgres pool: %w", err)
	}
	application.pool = pool

	// Redis client
	redisClient, err := redis.NewClient(cfg.RedisAddr)
	if err != nil {
		log.Printf("Redis connection error: %v", err)
	} else {
		application.redisClient = redisClient
	}

	// Todo cache
	var todoCache *redis.Todo
	if redisClient != nil {
		todoCache, err = redis.NewTodo(redisClient, 60*time.Second)
		if err != nil {
			log.Printf("Redis cache initialization error: %v", err)
		}
	}

	// Repository
	repo := postgres.NewTodoRepository(pool)
	userRepo := postgres.NewUserRepository(pool)
	tokenRepo := postgres.NewRefreshTokenRepository(pool)

	// Cached repository
	var todoRepository todo.Repository = repo
	if todoCache != nil {
		todoRepository = cached.NewTodoRepository(repo, todoCache)
	}

	// Service
	hasher := password.NewArgon2Hasher()
	tokens, err := token.NewAccessToken([]byte(cfg.JWTSecret), 15*time.Minute, "todo-app")
	if err != nil {
		return nil, err
	}

	todoService := todo.NewService(todoRepository)
	authService := auth.NewService(userRepo, hasher, tokens, tokenRepo)

	// Handler
	todoHandler := todo.NewHandler(todoService)
	authHandler := auth.NewHandler(authService)

	// Router
	router := gin.Default()

	api := router.Group("/api")
	v1 := api.Group("/v1")

	httpserver.NewRouter(v1, tokens, todoHandler, authHandler)

	// HTTP server
	application.server = httpserver.NewServer(router, cfg.HTTPPort)

	return application, nil
}

func (a *App) Run() error {
	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	errCh := make(chan error, 1)

	go func() {
		log.Printf("HTTP server listening on %s", a.server.Addr)

		if err := a.server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	select {
	case <-ctx.Done():
		log.Println("Shutdown signal received")
	case err := <-errCh:
		return fmt.Errorf("HTTP server error: %w", err)
	}

	shutdownCtx, cancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)
	defer cancel()

	if err := a.server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("HTTP server shutdown: %w", err)
	}

	log.Println("HTTP server stopped")

	return nil
}

func (a *App) Close() {
	if a.redisClient != nil {
		if err := a.redisClient.Close(); err != nil {
			log.Printf("Redis close error: %v", err)
		}
	}

	if a.pool != nil {
		a.pool.Close()
	}
}
