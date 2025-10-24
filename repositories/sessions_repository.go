package repositories

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"real-time-forum/models"
	"time"
)

type SessionRepository struct {
	db *sql.DB
}

func NewSessionRepository(db *sql.DB) *SessionRepository {
	return &SessionRepository{db: db}
}

func (r *SessionRepository) CreateSession(ctx context.Context, session models.Session) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO sessions (session_id, user_id, created_at, expires_at)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(session_id) DO UPDATE SET
			user_id = excluded.user_id,
			created_at = excluded.created_at,
			expires_at = excluded.expires_at
	`, session.ID, session.UserID, session.CreatedAt, session.ExpiresAt)

	return err
}

func (r *SessionRepository) GetSessionByID(ctx context.Context, sessionID string) (*models.Session, error) {
	s := models.Session{}
	err := r.db.QueryRowContext(ctx, "SELECT session_id, user_id, created_at, expires_at FROM sessions WHERE session_id = ?", sessionID).Scan(&s.ID, &s.UserID, &s.CreatedAt, &s.ExpiresAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return &s, nil
}

func (r *SessionRepository) CleanupExpiredSessions(ctx context.Context) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM sessions WHERE expires_at < ?`, time.Now())
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	log.Printf("Deleted %d expired sessions", n)
	return err
}

func (r *SessionRepository) DeleteSession(ctx context.Context, sessionID string) error {
	res, err := r.db.Exec(`DELETE FROM sessions WHERE session_id = ?`, sessionID)
	if err != nil {
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("no session found to delete")
	}
	return nil
}
