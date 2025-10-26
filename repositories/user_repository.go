package repositories

import (
	"context"
	"database/sql"
	"time"

	"real-time-forum/models"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) CreateUser(ctx context.Context, user *models.User) error {

	query := `INSERT INTO users (id, nickname, age, gender, first_name, last_name,  email, password)
	        VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := r.db.Exec(query, user.ID, user.Nickname, user.Age, user.Gender, user.FirstName, user.LastName, user.Email, user.Password)
	if err != nil {
		return err
	}
	return nil
}

func (r *UserRepository) CheckUser(ctx context.Context, user *models.User) (bool, error) {
	var exists int
	err := r.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM users WHERE nickname = ? OR email = ?)", user.Nickname, user.Email).Scan(&exists)
	if err != nil {
		return false, err // something went wrong with the query
	}
	return exists == 1, nil // true if user exists
}

func (r *UserRepository) GetUserByEmailorName(ctx context.Context, email, name string) (*models.User, error) {
	user := models.User{}

	err := r.db.QueryRowContext(ctx, `SELECT id, nickname, email, password FROM users WHERE nickname = ? OR email = ?`, name, email).Scan(&user.ID, &user.Nickname, &user.Email, &user.Password)

	if err != nil {

		return nil, err
	}

	return &user, nil
}

func (r *UserRepository) GetUserBySessionID(ctx context.Context, SessionID string) (*models.User, error) {
	user := models.User{}
	err := r.db.QueryRowContext(ctx, `SELECT u.id, u.nickname, u.email FROM sessions s JOIN users u ON u.id = s.user_id WHERE s.session_id = ?`, SessionID).Scan(&user.ID, &user.Nickname, &user.Email)
	if err != nil {

		return nil, err // return the raw DB error
	}
	return &user, nil
}

func (r *UserRepository) GetActiveUsers(ctx context.Context, sessionID string) ([]models.User, error) {
	now := time.Now()
	rows, err := r.db.QueryContext(ctx, "SELECT u.id, u.nickname, u.email FROM users u JOIN sessions s ON u.id = s.user_id WHERE s.expires_at > ? AND s.session_id != ?", now, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		if err := rows.Scan(&user.ID, &user.Nickname, &user.Email); err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, nil
}
