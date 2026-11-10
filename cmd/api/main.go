package main

import (
	"context"
	"log"
	"todo-app/internal/config"
	"todo-app/internal/http"
	"todo-app/internal/storage/postgres"
	"todo-app/internal/todo"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	ctx := context.Background()

	if err := godotenv.Load(); err != nil {
		log.Printf("No .env file found: %v", err)
	}

	cfg, err := config.LoadFromEnv()
	if err != nil {
		log.Fatal(err)
	}

	if err := postgres.RunMigrations(cfg.DatabaseURL); err != nil {
		log.Fatal(err)
	}

	pool, err := postgres.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	repo := postgres.NewTodoRepository(pool)
	todoService := todo.NewService(repo)
	todoHandler := todo.NewHandler(todoService)

	router := gin.Default()

	api := router.Group("/api")
	{
		v1 := api.Group("v1")

		http.NewRouter(v1, todoHandler)
	}

	router.Run(":8080")
}
