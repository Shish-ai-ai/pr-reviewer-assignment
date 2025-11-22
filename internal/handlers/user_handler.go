package handlers

import (
	"github.com/gin-gonic/gin"
	"prReviewerAssignment/internal/models"
	"prReviewerAssignment/internal/services"
)

type UserHandler struct {
	userService *services.UserService
}

func NewUserHandler() *UserHandler {
	return &UserHandler{
		userService: services.NewUserService(),
	}
}

func (h *UserHandler) SetUserActive(c *gin.Context) {
	var request struct {
		UserID   string `json:"user_id"`
		IsActive bool   `json:"is_active"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		h.sendError(c, "NOT_FOUND", "Invalid JSON data", 400)
		return
	}

	if request.UserID == "" {
		h.sendError(c, "NOT_FOUND", "user_id is required", 400)
		return
	}

	user, err := h.userService.SetUserActive(request.UserID, request.IsActive)
	if err != nil {
		if err.Error() == "user not found" {
			h.sendError(c, "NOT_FOUND", "user not found", 404)
			return
		}
		h.sendError(c, "NOT_FOUND", "Internal server error", 500)
		return
	}

	c.JSON(200, gin.H{
		"user": user,
	})
}

func (h *UserHandler) sendError(c *gin.Context, code, message string, statusCode int) {
	errorResponse := models.ErrorResponse{}
	errorResponse.Error.Code = code
	errorResponse.Error.Message = message
	c.JSON(statusCode, errorResponse)
}
