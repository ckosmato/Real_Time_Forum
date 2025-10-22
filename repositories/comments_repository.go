package repositories

import (
	"context"
	"database/sql"
	"real-time-forum/models"
)

type CommentRepository struct {
	db *sql.DB
}

func NewCommentRepository(db *sql.DB) *CommentRepository {
	return &CommentRepository{db: db}
}

func (r *CommentRepository) CreateComment(ctx context.Context, comment *models.Comment) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO comments 
		(id, post_id, author_id, content) 
		VALUES (?, ?, ?, ?)`,
		comment.ID, comment.PostID, comment.AuthorID, comment.Content)
	if err != nil {
		return err
	}

	return nil
}

func (r *CommentRepository) GetCommentByID(ctx context.Context, commentID string) (*models.Comment, error) {
	var comment models.Comment
	query := "SELECT id, post_id, author_id, content, created_at, updated_at FROM comments WHERE id = ? AND is_deleted = 0"
	err := r.db.QueryRowContext(ctx, query, commentID).Scan(&comment.ID, &comment.PostID, &comment.AuthorID, &comment.Content, &comment.CreatedAt, &comment.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &comment, nil
}

func (r *CommentRepository) UpdateComment(ctx context.Context, comment *models.Comment) error {
	_, err := r.db.ExecContext(ctx, "UPDATE comments SET content = ? WHERE id = ?", comment.Content, comment.ID)
	if err != nil {
		return err
	}

	return nil
}



func (r *CommentRepository) GetPostComments(ctx context.Context, postID string) ([]models.Comment, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT c.id, u.username, c.content, c.created_at, c.likes_count, c.dislikes_count
		FROM comments c
		JOIN users u ON c.author_id = u.id
		WHERE c.post_id = ? AND is_deleted = 0
		ORDER BY created_at DESC;
	`, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []models.Comment
	for rows.Next() {
		var comment models.Comment
		if err := rows.Scan(&comment.ID, &comment.AuthorName, &comment.Content, &comment.CreatedAt); err != nil {
			return nil, err
		}
		comments = append(comments, comment)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return comments, nil
}

func (r *CommentRepository) CommentsListByUser(ctx context.Context, userID string, limit, offset int) ([]models.Comment, error) {
	const query = `
        SELECT c.id, c.post_id, p.title, c.content, c.created_at
        FROM comments c
        JOIN posts p ON p.id = c.post_id
        WHERE c.author_id = $1 AND c.is_deleted = 0
        ORDER BY c.created_at DESC
        LIMIT $2 OFFSET $3;
    `
	rows, err := r.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []models.Comment
	for rows.Next() {
		var c models.Comment
		if err := rows.Scan(
			&c.ID,
			&c.PostID,
			&c.PostTitle,
			&c.Content,
			&c.CreatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}
