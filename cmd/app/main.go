package main

import (
	"github.com/joho/godotenv"
	"log"
	"prReviewerAssignment/internal/db"
	"prReviewerAssignment/internal/handlers"
	"prReviewerAssignment/internal/routes"
)

func main() {
	godotenv.Load()

	if err := db.InitDB(); err != nil {
		log.Fatal("failed to connect to the database: ", err)
	}

	teamHandler := handlers.NewTeamHandler()
	userHandler := handlers.NewUserHandler()

	router := routes.SetupRouter(teamHandler, userHandler)

	if err := router.Run(":8080"); err != nil {
		log.Fatal("failed to start server: ", err)
	}
}
