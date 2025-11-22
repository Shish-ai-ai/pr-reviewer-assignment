package routes

import (
	"prReviewerAssignment/internal/handlers"
	"github.com/gin-gonic/gin"
)

func SetupRouter(teamHandler *handlers.TeamHandler, userHandler *handlers.UserHandler, prHandler *handlers.PRHandler, statsHandler *handlers.StatsHandler) *gin.Engine {
	router := gin.Default()

	router.POST("/team/add", teamHandler.AddTeam)
	router.GET("/team/get", teamHandler.GetTeam)
	router.POST("/users/setIsActive", userHandler.SetUserActive)
	router.GET("/users/getReview", userHandler.GetUserReviews)
	router.POST("/pullRequest/create", prHandler.CreatePullRequest)
	router.POST("/pullRequest/merge", prHandler.MergePullRequest)
	router.POST("/pullRequest/reassign", prHandler.ReassignReviewer)
	router.GET("/stats/reviewers", statsHandler.GetReviewerStats)

	return router
}