package todo

import (
	"errors"
	"net/http"
	"todo-app/internal/auth"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) Create(c *gin.Context) {
	userID, ok := auth.UserIDFromContext(c.Request.Context())
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "unauthorized",
		})
		return
	}

	var req CreateTodoRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body",
		})
		return
	}

	todo, err := h.service.Create(c.Request.Context(), userID, req.Title, req.Description)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, ToTodoResponse(todo))
}

func (h *Handler) List(c *gin.Context) {
	userID, ok := auth.UserIDFromContext(c.Request.Context())
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "unauthorized",
		})
		return
	}

	todos, err := h.service.GetAll(c.Request.Context(), userID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response := make([]TodoResponse, 0, len(todos))

	for i := range todos {
		response = append(response, ToTodoResponse(&todos[i]))
	}

	c.JSON(http.StatusOK, response)
}

func (h *Handler) GetByID(c *gin.Context) {
	userID, ok := auth.UserIDFromContext(c.Request.Context())
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "unauthorized",
		})
		return
	}

	id := c.Param("id")

	todo, err := h.service.GetByID(c.Request.Context(), userID, id)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, ToTodoResponse(todo))
}

func (h *Handler) Update(c *gin.Context) {
	userID, ok := auth.UserIDFromContext(c.Request.Context())
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "unauthorized",
		})
		return
	}

	var req UpdateTodoRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body",
		})
		return
	}

	todo, err := h.service.Update(c.Request.Context(), userID, req.ID, req.Title, req.Description)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, ToTodoResponse(todo))
}

func (h *Handler) Delete(c *gin.Context) {
	userID, ok := auth.UserIDFromContext(c.Request.Context())
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "unauthorized",
		})
		return
	}

	id := c.Param("id")

	if err := h.service.Delete(c.Request.Context(), userID, id); err != nil {
		h.handleError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *Handler) Done(c *gin.Context) {
	id := c.Param("id")
	userID, ok := auth.UserIDFromContext(c.Request.Context())
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "unauthorized",
		})
		return
	}

	todo, err := h.service.Done(c.Request.Context(), userID, id)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, ToTodoResponse(todo))
}

func (h *Handler) Undone(c *gin.Context) {
	id := c.Param("id")
	userID, ok := auth.UserIDFromContext(c.Request.Context())
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "unauthorized",
		})
		return
	}

	todo, err := h.service.Undone(c.Request.Context(), userID, id)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, ToTodoResponse(todo))
}

func (h *Handler) handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrEmptyTitle):
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})

	case errors.Is(err, ErrTodoNotFound):
		c.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})

	default:
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "internal server error",
		})
	}
}
