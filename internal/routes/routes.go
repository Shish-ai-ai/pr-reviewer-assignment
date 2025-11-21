package routes

import (
	"github.com/gin-gonic/gin"
	"prReviewerAssignment/internal/handlers"
)

func SetupRouter(teamHandler *handlers.TeamHandler) *gin.Engine {
	router := gin.Default()

	router.POST("/team/add", teamHandler.AddTeam)

	return router
}
