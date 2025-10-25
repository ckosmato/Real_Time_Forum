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

func (s *UserService) GetUserBySessionID(ctx context.Context, sessionID string) (*models.User, error) {
	user, err := s.repo.GetUserBySessionID(ctx, sessionID)
	if errors.Is(err, sql.ErrNoRows) {
		log.Printf("GetUserBySessionID: user with session ID %s not found", sessionID)
		return nil, errors.New("user not found")
	}
	if err != nil {
		log.Printf("GetUserBySessionID: failed to retrieve user %s: %v", sessionID, err)
		return nil, errors.New("failed to retrieve user")
	}
	return user, nil
}

func (s *UserService) GetActiveUsers(ctx context.Context) ([]models.User, error) {
	users, err := s.repo.GetActiveUsers(ctx)
	if err != nil {
		log.Printf("GetActiveUsers: internal server error: %v", err)
		return nil, errors.New("internal server error")
	}
	return users, nil
}
