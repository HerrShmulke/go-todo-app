package http

import (
	"todo-app/internal/todo"

	"github.com/gin-gonic/gin"
)

func NewRouter(router *gin.RouterGroup, todoHandler *todo.Handler) {
	todos := router.Group("/todos")
	{
		todos.POST("", todoHandler.Create)
		todos.GET("", todoHandler.List)
		todos.GET("/:id", todoHandler.GetById)
		todos.PUT("", todoHandler.Update)
	}
}
