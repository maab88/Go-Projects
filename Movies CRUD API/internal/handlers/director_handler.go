package handlers

import (
	"MoviesCRUDAPI/internal/models"
	"MoviesCRUDAPI/internal/services"
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type DirectorHandler struct {
	DirectorService *services.DirectorService
}

func NewDirectorHandler(service *services.DirectorService) *DirectorHandler {
	return &DirectorHandler{DirectorService: service}
}

func (h *DirectorHandler) GetDirectors(ctx *gin.Context) {
	directors, err := h.DirectorService.GetAllDirectors(context.Background())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}
	ctx.JSON(http.StatusOK, directors)
}

func (h *DirectorHandler) GetDirectorByID(ctx *gin.Context) {
	id, _ := strconv.Atoi(ctx.Param("id"))
	director, err := h.DirectorService.GetDirectorByID(context.Background(), id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Director not found"})
		return
	}
	ctx.JSON(http.StatusOK, director)
}

func (h *DirectorHandler) CreateDirector(ctx *gin.Context) {
	var director models.Director
	if err := ctx.ShouldBindJSON(&director); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}
	id, err := h.DirectorService.CreateDirector(context.Background(), director)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create director"})
		return
	}
	ctx.JSON(http.StatusCreated, gin.H{"id": id})
}

func (h *DirectorHandler) UpdateDirector(ctx *gin.Context) {
	var director models.Director
	id, _ := strconv.Atoi(ctx.Param("id"))
	director.ID = id
	if err := ctx.ShouldBindJSON(&director); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}
	if err := h.DirectorService.UpdateDirector(context.Background(), director); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update director"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "Director updated successfully"})
}

func (h *DirectorHandler) DeleteDirector(ctx *gin.Context) {
	id, _ := strconv.Atoi(ctx.Param("id"))
	if err := h.DirectorService.DeleteDirector(context.Background(), id); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete director"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "Director deleted successfully"})
}
