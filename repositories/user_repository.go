package repositories

import (
	"context"
	"database/sql"

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
	err := r.db.QueryRowContext(ctx, `SELECT u.nickname, u.email FROM sessions s JOIN users u ON u.id = s.user_id WHERE s.session_id = ?`, SessionID).Scan(&user.Nickname, &user.Email)
	if err != nil {

		return nil, err // return the raw DB error
	}
	return &user, nil
}

func (r *UserRepository) GetAllUsers(ctx context.Context) ([]models.User, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT id, nickname, email FROM users")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		if err := rows.Scan(&user.ID, &user.Nickname); err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, nil
}



