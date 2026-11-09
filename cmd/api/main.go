package main

import (
	"todo-app/internal/http"
	"todo-app/internal/todo"

	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()

	api := router.Group("/api")
	{
		v1 := api.Group("v1")

		repo := todo.NewMemoryRepository()
		todoService := todo.NewService(repo)
		todoHandler := todo.NewHandler(todoService)
		http.NewRouter(v1, todoHandler)
	}

	router.Run(":8080")
}
