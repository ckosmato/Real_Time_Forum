package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"real-time-forum/models"
	"log"
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
		INSERT INTO sessions (user_id, session_id, csrf_token, created_at, expires_at)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(user_id) DO UPDATE SET
			session_id = excluded.session_id,
			csrf_token = excluded.csrf_token,
			created_at = excluded.created_at,
			expires_at = excluded.expires_at
	`, session.UserID, session.ID, session.CreatedAt, session.ExpiresAt)

	return err
}

func (r *SessionRepository) GetSessionByID(ctx context.Context, sessionID string) (*models.Session, error) {
	s := models.Session{}
	err := r.db.QueryRowContext(ctx, "SELECT session_id,csrf_token , user_id, created_at, expires_at FROM sessions WHERE session_id = ?", sessionID).Scan(&s.ID,  &s.UserID, &s.CreatedAt, &s.ExpiresAt)
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
