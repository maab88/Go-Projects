package services

import (
	"MoviesCRUDAPI/internal/models"
	"MoviesCRUDAPI/internal/repositories"
	"context"
)

type DirectorService struct {
	DirectorRepo *repositories.DirectorRepository
}

func NewDirectorService(repo *repositories.DirectorRepository) *DirectorService {
	return &DirectorService{DirectorRepo: repo}
}

func (s *DirectorService) GetAllDirectors(ctx context.Context) ([]models.Director, error) {
	return s.DirectorRepo.GetAllDirectors(ctx)
}

func (s *DirectorService) GetDirectorByID(ctx context.Context, id int) (models.Director, error) {
	return s.DirectorRepo.GetDirectorByID(ctx, id)
}

func (s *DirectorService) CreateDirector(ctx context.Context, director models.Director) (int, error) {
	return s.DirectorRepo.CreateDirector(ctx, director)
}

func (s *DirectorService) UpdateDirector(ctx context.Context, director models.Director) error {
	return s.DirectorRepo.UpdateDirector(ctx, director)
}

func (s *DirectorService) DeleteDirector(ctx context.Context, id int) error {
	return s.DirectorRepo.DeleteDirector(ctx, id)
}
