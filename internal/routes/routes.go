package routes

import (
	"github.com/gin-gonic/gin"
	"prReviewerAssignment/internal/handlers"
)

func SetupRouter(teamHandler *handlers.TeamHandler, userHandler *handlers.UserHandler, prHandler *handlers.PRHandler) *gin.Engine {
	router := gin.Default()

	router.POST("/team/add", teamHandler.AddTeam)
	router.GET("/team/get", teamHandler.GetTeam)
	router.POST("/users/setIsActive", userHandler.SetUserActive)
	router.POST("/pullRequest/create", prHandler.CreatePullRequest)
	router.POST("/pullRequest/merge", prHandler.MergePullRequest)

	return router
}
