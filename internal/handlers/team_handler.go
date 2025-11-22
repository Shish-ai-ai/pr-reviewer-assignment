package handlers

import (
	"github.com/gin-gonic/gin"
	"prReviewerAssignment/internal/models"
	"prReviewerAssignment/internal/services"
)

type TeamHandler struct {
	teamService *services.TeamService
}

func NewTeamHandler() *TeamHandler {
	return &TeamHandler{
		teamService: services.NewTeamService(),
	}
}

func (h *TeamHandler) AddTeam(c *gin.Context) {
	var team models.Team
	if err := c.ShouldBindJSON(&team); err != nil {
		h.sendError(c, "TEAM_EXISTS", "Invalid JSON", 400)
		return
	}

	if team.TeamName == "" {
		h.sendError(c, "TEAM_EXISTS", "Team name is required", 400)
		return
	}
	if len(team.Members) == 0 {
		h.sendError(c, "TEAM_EXISTS", "Team must have at least one member", 400)
		return
	}

	createdTeam, err := h.teamService.CreateTeam(team)
	if err != nil {
		if err.Error() == "team already exists" {
			h.sendError(c, "TEAM_EXISTS", "team_name already exists", 400)
			return
		}
		h.sendError(c, "TEAM_EXISTS", "Internal server error", 500)
		return
	}

	c.JSON(201, gin.H{
		"team": createdTeam,
	})
}

func (h *TeamHandler) GetTeam(c *gin.Context) {
	teamName := c.Query("team_name")
	if teamName == "" {
		h.sendError(c, "NOT_FOUND", "team_name parameter is required", 400)
		return
	}

	team, err := h.teamService.GetTeam(teamName)
	if err != nil {
		if err.Error() == "team not found" {
			h.sendError(c, "NOT_FOUND", "team not found", 404)
			return
		}
		h.sendError(c, "NOT_FOUND", "Internal server error", 500)
		return
	}

	c.JSON(200, team)
}

func (h *TeamHandler) sendError(c *gin.Context, code, message string, statusCode int) {
	errorResponse := models.ErrorResponse{}
	errorResponse.Error.Code = code
	errorResponse.Error.Message = message
	c.JSON(statusCode, errorResponse)
}
