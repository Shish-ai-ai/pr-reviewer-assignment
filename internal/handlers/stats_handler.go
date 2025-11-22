package handlers

import (
	"prReviewerAssignment/internal/models"
	"prReviewerAssignment/internal/services"
	"github.com/gin-gonic/gin"
)

type StatsHandler struct {
	statsService *services.StatsService
}

func NewStatsHandler() *StatsHandler {
	return &StatsHandler{
		statsService: services.NewStatsService(),
	}
}

func (h *StatsHandler) GetReviewerStats(c *gin.Context) {
	stats, err := h.statsService.GetReviewerStats()
	if err != nil {
		h.sendError(c, "NOT_FOUND", "Internal server error", 500)
		return
	}

	c.JSON(200, stats)
}

func (h *StatsHandler) sendError(c *gin.Context, code, message string, statusCode int) {
	errorResponse := models.ErrorResponse{}
	errorResponse.Error.Code = code
	errorResponse.Error.Message = message
	c.JSON(statusCode, errorResponse)
}