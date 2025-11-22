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

func (h *PRHandler) MergePullRequest(c *gin.Context) {
	var request models.MergePRRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		h.sendError(c, "NOT_FOUND", "Invalid JSON data", 400)
		return
	}

	pr, err := h.prService.MergePullRequest(request.PullRequestID)
	if err != nil {
		if err.Error() == "PR not found" {
			h.sendError(c, "NOT_FOUND", "PR not found", 404)
			return
		}
		h.sendError(c, "NOT_FOUND", "Internal server error", 500)
		return
	}

	c.JSON(200, gin.H{
		"pr": pr,
	})
}

func (h *PRHandler) ReassignReviewer(c *gin.Context) {
	var request models.ReassignPRRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		h.sendError(c, "NOT_FOUND", "Invalid JSON data", 400)
		return
	}

	pr, newReviewer, err := h.prService.ReassignReviewer(request.PullRequestID, request.OldReviewerID)
	if err != nil {
		switch err.Error() {
		case "PR not found":
			h.sendError(c, "NOT_FOUND", "PR not found", 404)
		case "old reviewer not found or inactive":
			h.sendError(c, "NOT_FOUND", "reviewer not found or inactive", 404)
		case "cannot reassign on merged PR":
			h.sendError(c, "PR_MERGED", "cannot reassign on merged PR", 409)
		case "reviewer is not assigned to this PR":
			h.sendError(c, "NOT_ASSIGNED", "reviewer is not assigned to this PR", 409)
		case "no active replacement candidate in team":
			h.sendError(c, "NO_CANDIDATE", "no active replacement candidate in team", 409)
		default:
			h.sendError(c, "NOT_FOUND", "Internal server error", 500)
		}
		return
	}

	c.JSON(200, gin.H{
		"pr":          pr,
		"replaced_by": newReviewer,
	})
}

func (h *PRHandler) sendError(c *gin.Context, code, message string, statusCode int) {
	errorResponse := models.ErrorResponse{}
	errorResponse.Error.Code = code
	errorResponse.Error.Message = message
	c.JSON(statusCode, errorResponse)
}
