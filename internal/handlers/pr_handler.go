package handlers

import (
	"github.com/gin-gonic/gin"
	"prReviewerAssignment/internal/models"
	"prReviewerAssignment/internal/services"
)

type PRHandler struct {
	prService *services.PRService
}

func NewPRHandler() *PRHandler {
	return &PRHandler{
		prService: services.NewPRService(),
	}
}

func (h *PRHandler) CreatePullRequest(c *gin.Context) {
	var request models.CreatePRRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		h.sendError(c, "PR_EXISTS", "Invalid JSON data", 400)
		return
	}

	pr, err := h.prService.CreatePullRequest(request)
	if err != nil {
		if err.Error() == "PR already exists" {
			h.sendError(c, "PR_EXISTS", "PR id already exists", 409)
			return
		}
		if err.Error() == "author not found or inactive" {
			h.sendError(c, "NOT_FOUND", "author not found or inactive", 404)
			return
		}
		h.sendError(c, "PR_EXISTS", "Internal server error", 500)
		return
	}

	c.JSON(201, gin.H{
		"pr": pr,
	})
}

func (h *PRHandler) sendError(c *gin.Context, code, message string, statusCode int) {
	errorResponse := models.ErrorResponse{}
	errorResponse.Error.Code = code
	errorResponse.Error.Message = message
	c.JSON(statusCode, errorResponse)
}
