package services

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"real-time-forum/models"
	repositories "real-time-forum/repositories"
	"strings"

	"github.com/gofrs/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	repo repositories.UserRepository
}

func NewAuthService(repo repositories.UserRepository) *AuthService {
	return &AuthService{repo: repo}
}

func (s *AuthService) Register(ctx context.Context, user *models.User) error {
	err := validateUser(user, true)
	if err != nil {
		log.Printf("RegisterUser: validation error: %v", err)
		return err
	}
	exists, err := s.repo.CheckUser(ctx, user)
	if err != nil {
		log.Printf("RegisterUser: error checking user: %v", err)
		return err
	}
	if exists {
		log.Printf("RegisterUser: user already exists: %s", user.Nickname)
		return errors.New("user already exists")
	}

	hashedPass, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("RegisterUser: error hashing password: %v", err)
		return err
	}

	u1, err := uuid.NewV4()
	if err != nil {
		log.Printf("RegisterUser: error generating user ID: %v", err)
		return err
	}
	user.ID = u1.String()
	user.Password = string(hashedPass)
	if err = s.repo.CreateUser(ctx, user); err != nil {
		log.Printf("RegisterUser: error creating user: %v", err)
		return err
	}
	return nil
}


func (s *AuthService) LoginUser(ctx context.Context, input *models.User) (*models.User, error) {
	err := validateUser(input, false)
	if err != nil {
		log.Printf("LoginUser: validation error: %v", err)
		return nil, err
	}
	user, err := s.repo.GetUserByEmailorName(ctx, input.Email, input.Nickname)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Printf("LoginUser: user not found: %v", err)
			return nil,err
		}
		log.Printf("LoginUser: error retrieving user: %v", err)
		return nil,err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password))
	if err != nil {
		log.Printf("LoginUser: invalid credentials for user %s: %v", user.Nickname, err)
		return nil, err
	}
	return user, nil
}


func validateUser(user *models.User, strict bool) error {
	switch strict {
	case false:
		if strings.TrimSpace(user.Nickname) == "" && strings.TrimSpace(user.Email) == "" {
			return errors.New("must provide nickname or email")
		}
		if strings.TrimSpace(user.Password) == "" {
			return errors.New("password cannot be empty")
		}

	case true:
		if len(user.Nickname) < 4 {
			return errors.New("nickname must be longer that four character")
		} else if len(user.Nickname) > 20 {
			return errors.New("nickname must be shorter than twenty characters")
		} else if strings.TrimSpace(user.Email) == "" || !strings.Contains(user.Email, "@") {
			return errors.New("invalid email address")
		} else if len(user.Password) < 6 {
			return errors.New("password must be at least six characters long")
		}
	}
	return nil
}