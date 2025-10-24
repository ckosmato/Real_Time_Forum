package services

import (
	"context"
	"errors"
	"log"
	"real-time-forum/models"
	repos "real-time-forum/repositories"
	"strings"

	"github.com/gofrs/uuid"
)

type CategoriesService struct {
	repo repos.CategoriesRepository
}

func NewCategoriesService(repo repos.CategoriesRepository) *CategoriesService {
	return &CategoriesService{repo: repo}
}

func (s *CategoriesService) CreateCategory(ctx context.Context, categoryName string) error {

	if strings.TrimSpace(categoryName) == "" {
		log.Printf("CreateCategory: category name is required")
		return errors.New("category name is required")
	}

	categoryID, err := uuid.NewV4()
	if err != nil {
		log.Printf("CreateCategory: failed to generate category ID: %v", err)
		return errors.New("failed to generate category ID")
	}

	category := models.Category{
		ID:   categoryID.String(),
		Name: categoryName,
	}

	if err := s.repo.CreateCategory(ctx, category); err != nil {
		log.Printf("CreateCategory: failed to create category: %v", err)
		return errors.New("failed to create category")
	}
	return nil
}

func (s *CategoriesService) GetAllCategories(ctx context.Context) ([]models.Category, error) {
	categories, err := s.repo.GetAllCategories(ctx)
	if err != nil {
		log.Printf("GetAllCategories: failed to retrieve categories AYTO EDW PERNOYME: %v", err)
		return nil, errors.New("failed to retrieve categories")
	}
	return categories, nil
}
