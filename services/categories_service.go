package services

import (
	"context"
	"errors"
	"log"
	"real-time-forum/models"
	repos "real-time-forum/repositories"
)

type CategoriesService struct {
	repo repos.CategoriesRepository
}

func NewCategoriesService(repo repos.CategoriesRepository) *CategoriesService {
	return &CategoriesService{repo: repo}
}

func (s *CategoriesService) GetAllCategories(ctx context.Context) ([]models.Category, error) {
	categories, err := s.repo.GetAllCategories(ctx)
	if err != nil {
		log.Printf("GetAllCategories: failed to retrieve categories AYTO EDW PERNOYME: %v", err)
		return nil, errors.New("failed to retrieve categories")
	}
	return categories, nil
}
