package routes

import (
	"MoviesCRUDAPI/internal/handlers"

	"github.com/gin-gonic/gin"
)

func SetupRouter(movieHandler *handlers.MovieHandler, directorHandler *handlers.DirectorHandler) *gin.Engine {
	r := gin.Default()

	r.GET("/movies", movieHandler.GetMovies)
	r.GET("/movies/:id", movieHandler.GetMovieByID)
	r.POST("/movies", movieHandler.CreateMovie)
	r.PUT("/movies/:id", movieHandler.UpdateMovie)
	r.DELETE("/movies/:id", movieHandler.DeleteMovie)

	r.GET("/directors", directorHandler.GetDirectors)
	r.GET("/directors/:id", directorHandler.GetDirectorByID)
	r.POST("/directors", directorHandler.CreateDirector)
	r.PUT("/directors/:id", directorHandler.UpdateDirector)
	r.DELETE("/directors/:id", directorHandler.DeleteDirector)

	return r
}
