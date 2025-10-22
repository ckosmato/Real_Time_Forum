package services

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"real-time-forum/models"
	repositories "real-time-forum/repositories"
	"time"

	"github.com/gofrs/uuid"
)

type SessionService struct {
	repo repositories.SessionRepository
}

func NewSessionService(repo repositories.SessionRepository) *SessionService {
	return &SessionService{repo: repo}
}

func (s *SessionService) GenerateSession(ctx context.Context, user *models.User) (models.Session, error) {

	u1, err := uuid.NewV4()
	if err != nil {
		log.Printf("GenerateSession: failed to generate session ID: %v", err)
		return models.Session{}, errors.New("failed to generate session ID")
	}

	createdAt := time.Now()
	expiresAt := time.Now().Add(time.Hour)

	session := models.Session{
		ID:        u1.String(),
		UserID:    user.ID,
		CreatedAt: createdAt,
		ExpiresAt: expiresAt,
	}

	if err := s.repo.CreateSession(ctx, session); err != nil {
		log.Printf("GenerateSession: failed to save session: %v", err)
		return models.Session{}, errors.New("failed to save session")
	}

	return session, nil
}

func (s *SessionService) GetSession(ctx context.Context, sessionID string) (*models.Session, error) {
	session, err := s.repo.GetSessionByID(ctx, sessionID)
	if errors.Is(err, sql.ErrNoRows) {
		log.Printf("GetSession: session %s not found", sessionID)
		return nil, errors.New("session not found")
	}
	if err != nil {
		log.Printf("GetSession: internal server error for session %s: %v", sessionID, err)
		return nil, errors.New("internal server error")
	}

	if time.Now().After(session.ExpiresAt) {
		log.Printf("GetSession: session %s expired", sessionID)
		return nil, errors.New("session expired")
	}

	return session, nil
}

func (s *SessionService) ExpireSession(ctx context.Context, sessionID string) error {
	err := s.repo.DeleteSession(ctx, sessionID)
	if err != nil {
		log.Printf("ExpireSession: failed to expire session %s: %v", sessionID, err)
		return errors.New("failed to expire session")
	}

	return nil
}

func (s *SessionService) CleanupExpiredSessions(ctx context.Context) error {
	err := s.repo.CleanupExpiredSessions(ctx)
	if err != nil {
		log.Printf("CleanupExpiredSessions: failed to cleanup expired sessions: %v", err)
		return errors.New("failed to cleanup expired sessions")
	}
	return nil
}
