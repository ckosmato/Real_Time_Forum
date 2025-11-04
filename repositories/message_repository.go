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

// GetMessagesWithPagination retrieves chat messages with pagination support
func (r *MessageRepository) GetMessagesWithPagination(ctx context.Context, user1, user2 string, limit, offset int) ([]models.Message, error) {
	query := `
		SELECT id, from_user, to_user, body, created_at
		FROM messages 
		WHERE (from_user = ? AND to_user = ?) OR (from_user = ? AND to_user = ?)
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := r.db.QueryContext(ctx, query, user1, user2, user2, user1, limit, offset)
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

	// Reverse the order to show chronological order (oldest first in the returned batch)
	// This ensures that when we display messages, they appear in chronological order
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages, rows.Err()
}

// GetLastMessageTimestamps gets the last message timestamp for each user the current user has chatted with
func (r *MessageRepository) GetLastMessageTimestamps(ctx context.Context, currentUser string) (map[string]string, error) {
	query := `
		SELECT 
			CASE 
				WHEN from_user = ? THEN to_user 
				ELSE from_user 
			END as other_user,
			MAX(created_at) as last_message_time
		FROM messages 
		WHERE from_user = ? OR to_user = ?
		GROUP BY other_user
	`

	rows, err := r.db.QueryContext(ctx, query, currentUser, currentUser, currentUser)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	timestamps := make(map[string]string)
	for rows.Next() {
		var otherUser, lastMessageTime string
		err := rows.Scan(&otherUser, &lastMessageTime)
		if err != nil {
			return nil, err
		}
		timestamps[otherUser] = lastMessageTime
	}

	return timestamps, rows.Err()
}
