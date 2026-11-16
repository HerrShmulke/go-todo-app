package http

import (
	"todo-app/internal/auth"
	"todo-app/internal/auth/token"
	"todo-app/internal/todo"

	"github.com/gin-gonic/gin"
)

func NewRouter(
	router *gin.RouterGroup,
	accessToken *token.AccessToken,
	todoHandler *todo.Handler,
	authHandler *auth.Handler,
) {
	authMiddleware := auth.AuthMiddleware(accessToken)

	todos := router.Group("/todos")
	todos.Use(authMiddleware)
	{
		todos.POST("", todoHandler.Create)
		todos.GET("", todoHandler.List)
		todos.GET("/:id", todoHandler.GetByID)
		todos.DELETE("/:id", todoHandler.Delete)
		todos.PUT("", todoHandler.Update)
		todos.PATCH("/:id/done", todoHandler.Done)
		todos.PATCH("/:id/undone", todoHandler.Undone)
	}

	auth := router.Group("/auth")
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
		auth.POST("/refresh", authHandler.Refresh)
		auth.POST("/logout", authHandler.Logout)
	}
}
