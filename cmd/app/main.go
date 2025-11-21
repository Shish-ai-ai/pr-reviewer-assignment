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

	router := routes.SetupRouter(teamHandler)

	if err := router.Run(":8080"); err != nil {
		log.Fatal("failed to start server: ", err)
	}
}
