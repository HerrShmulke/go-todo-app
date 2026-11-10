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
		todos.DELETE("/:id", todoHandler.Delete)
		todos.PUT("", todoHandler.Update)
		todos.PATCH("/:id/done", todoHandler.Done)
		todos.PATCH("/:id/undone", todoHandler.Undone)
	}
}
