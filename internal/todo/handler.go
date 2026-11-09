package todo

import (
	"errors"
	"net/http"

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
	var req CreateTodoRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body",
		})
		return
	}

	todo, err := h.service.Create(c.Request.Context(), req.Title, req.Description)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, ToTodoResponse(todo))
}

func (h *Handler) List(c *gin.Context) {
	todos, err := h.service.GetAll(c.Request.Context())
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

func (h *Handler) handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrEmptyTitle):
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})

	default:
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "internal server error",
		})
	}
}
