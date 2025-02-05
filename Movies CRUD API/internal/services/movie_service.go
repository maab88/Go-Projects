package services

import (
	"MoviesCRUDAPI/internal/models"
	"MoviesCRUDAPI/internal/repositories"
	"context"
)

type MovieService struct {
	MovieRepo *repositories.MovieRepository
}

func NewMovieService(repo *repositories.MovieRepository) *MovieService {
	return &MovieService{MovieRepo: repo}
}

func (s *MovieService) GetAllMovies(ctx context.Context) ([]models.Movie, error) {
	return s.MovieRepo.GetAllMovies(ctx)
}

func (s *MovieService) GetMovieByID(ctx context.Context, id int) (models.Movie, error) {
	return s.MovieRepo.GetMovieByID(ctx, id)
}

func (s *MovieService) CreateMovie(ctx context.Context, movie models.Movie) (int, error) {
	return s.MovieRepo.CreateMovie(ctx, movie)
}

func (s *MovieService) UpdateMovie(ctx context.Context, movie models.Movie) error {
	return s.MovieRepo.UpdateMovie(ctx, movie)
}

func (s *MovieService) DeleteMovie(ctx context.Context, id int) error {
	return s.MovieRepo.DeleteMovie(ctx, id)
}
