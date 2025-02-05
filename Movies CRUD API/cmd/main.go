package main

import (
	"MoviesCRUDAPI/internal/database"
	"MoviesCRUDAPI/internal/handlers"
	"MoviesCRUDAPI/internal/repositories"
	"MoviesCRUDAPI/internal/routes"
	"MoviesCRUDAPI/internal/services"
	"context"
	"log"
)

func main() {
	connString := "postgres://postgres:test@localhost:5432/movies"
	log.Println("Starting Server on port 8000")

	db, err := database.ConnectDB(connString)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close(context.Background())

	// Dependency Injection
	movieRepo := repositories.NewMovieRepository(db)
	directorRepo := repositories.NewDirectorRepository(db)

	movieService := services.NewMovieService(movieRepo)
	directorService := services.NewDirectorService(directorRepo)

	movieHandler := handlers.NewMovieHandler(movieService)
	directorHandler := handlers.NewDirectorHandler(directorService)

	// Set up routes
	router := routes.SetupRouter(movieHandler, directorHandler)
	router.Run(":8000")
}
