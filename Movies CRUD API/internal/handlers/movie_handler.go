package handlers

import (
	"MoviesCRUDAPI/internal/models"
	"MoviesCRUDAPI/internal/services"
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type MovieHandler struct {
	MovieService *services.MovieService
}

func NewMovieHandler(service *services.MovieService) *MovieHandler {
	return &MovieHandler{MovieService: service}
}

func (h *MovieHandler) GetMovies(ctx *gin.Context) {
	movies, err := h.MovieService.GetAllMovies(context.Background())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, movies)
}

func (h *MovieHandler) GetMovieByID(ctx *gin.Context) {
	id, _ := strconv.Atoi(ctx.Param("id"))
	movie, err := h.MovieService.GetMovieByID(context.Background(), id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, movie)
}

func (h *MovieHandler) CreateMovie(ctx *gin.Context) {
	var movie models.Movie
	if err := ctx.ShouldBindJSON(&movie); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}
	id, err := h.MovieService.CreateMovie(context.Background(), movie)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusCreated, gin.H{"id": id})
}

func (h *MovieHandler) UpdateMovie(ctx *gin.Context) {
	var movie models.Movie
	id, _ := strconv.Atoi(ctx.Param("id"))
	movie.ID = id
	if err := ctx.ShouldBindJSON(&movie); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.MovieService.UpdateMovie(context.Background(), movie); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update movie"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "Movie updated successfully"})
}

func (h *MovieHandler) DeleteMovie(ctx *gin.Context) {
	id, _ := strconv.Atoi(ctx.Param("id"))
	if err := h.MovieService.DeleteMovie(context.Background(), id); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete movie"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "Movie deleted successfully"})
}
