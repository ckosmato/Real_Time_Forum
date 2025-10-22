package services

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"real-time-forum/models"
	repo "real-time-forum/repositories"
)

type UserService struct {
	repo repo.UserRepository
}

func NewUserService(r repo.UserRepository) *UserService {
	return &UserService{repo: r}
}

func (s *UserService) GetUserByID(ctx context.Context, id string) (*models.User, error) {
	user, err := s.repo.GetUserByID(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		log.Printf("GetUserByID: user with ID %s not found", id)
		return nil, errors.New("user not found")
	}
	if err != nil {
		log.Printf("GetUserByID: failed to retrieve user %s: %v", id, err)
		return nil, errors.New("failed to retrieve user")
	}
	return user, nil
}

func (s *UserService) GetAllUsers(ctx context.Context) ([]models.User, error) {
	users, err := s.repo.GetAllUsers(ctx)
	if err != nil {
		log.Printf("GetAllUsers: internal server error: %v", err)
		return nil, errors.New("internal server error")
	}
	return users, nil
}
