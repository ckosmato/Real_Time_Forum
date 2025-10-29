package repositories

import (
	"context"
	"database/sql"
	"real-time-forum/models"
)

type MessageRepository struct {
	db *sql.DB
}

func NewMessageRepository(db *sql.DB) *MessageRepository {
	return &MessageRepository{db: db}
}

// SaveMessage saves a chat message to the database
func (r *MessageRepository) SaveMessage(ctx context.Context, message *models.Message) error {
	query := `
		INSERT INTO messages (from_user, to_user, body, created_at)
		VALUES (?, ?, ?, ?)
	`

	_, err := r.db.ExecContext(ctx, query, message.From, message.To, message.Content, message.Timestamp)
	return err
}

// GetMessages retrieves chat messages between two users
func (r *MessageRepository) GetMessages(ctx context.Context, user1, user2 string, limit int) ([]models.Message, error) {
	query := `
		SELECT id, from_user, to_user, body, created_at
		FROM messages 
		WHERE from_user = ? AND to_user = ? OR from_user = ? AND to_user = ?
		ORDER BY created_at ASC
		LIMIT ?
	`

	rows, err := r.db.QueryContext(ctx, query, user1, user2, user2, user1, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []models.Message
	for rows.Next() {
		var msg models.Message
		err := rows.Scan(&msg.ID, &msg.From, &msg.To, &msg.Content, &msg.Timestamp)
		if err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}

	return messages, rows.Err()
}

// // GetRecentConversations gets the most recent conversations for a user
// func (r *MessageRepository) GetRecentConversations(ctx context.Context, userID string, limit int) ([]models.Message, error) {
// 	query := `
// 		SELECT id, from_user, to_user, body, created_at
// 		FROM messages m1
// 		WHERE m1.created_at = (
// 			SELECT MAX(m2.created_at)
// 			FROM messages m2
// 			WHERE (
// 				(m2.from_user = ? AND m2.to_user = m1.to_user) OR
// 				(m2.to_user = ? AND m2.from_user = m1.from_user)
// 			) AND m1.from_user = ? OR m1.to_user = ?
// 		)
// 		ORDER BY created_at DESC
// 		LIMIT ?
// 	`

// 	rows, err := r.db.QueryContext(ctx, query, userID, userID, userID, userID, limit)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer rows.Close()

// 	var messages []models.Message
// 	for rows.Next() {
// 		var msg models.Message
// 		err := rows.Scan(&msg.ID, &msg.From, &msg.To, &msg.Content, &msg.Timestamp)
// 		if err != nil {
// 			return nil, err
// 		}
// 		messages = append(messages, msg)
// 	}

// 	return messages, rows.Err()
// }
