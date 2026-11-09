package main

import (
	"todo-app/internal/todo"

	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()

	api := router.Group("/api")

	{
		v1 := api.Group("v1")
		v1.GET("/todo", todo.TodoList)
	}

	router.Run(":8080")
}
