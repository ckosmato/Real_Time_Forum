package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"real-time-forum/models"
	"strings"
)

type PostRepository struct {
	db *sql.DB
}

func NewPostRepository(db *sql.DB) *PostRepository {
	return &PostRepository{db: db}
}

func (r *PostRepository) CreatePost(ctx context.Context, user *models.User, post *models.Post, categoryIDs []string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("CreatePost: could not begin transaction: %w", err)
	}
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		} else if err != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				err = fmt.Errorf("CreatePost: original error: %v, rollback error: %v", err, rbErr)
			}
		} else {
			if commitErr := tx.Commit(); commitErr != nil {
				err = fmt.Errorf("CreatePost: could not commit transaction: %w", commitErr)
			}
		}
	}()

	insertPostQuery := `
		INSERT INTO posts (id, author_id, title, content, created_at, image_name) 
		VALUES (?, ?, ?, ?, ?, ?)`
	_, err = tx.ExecContext(ctx, insertPostQuery, post.ID, user.ID, post.Title, post.Content, post.CreatedAt, post.Image)
	if err != nil {
		return fmt.Errorf("CreatePost: inserting post: %w", err)
	}

	insertCategoryQuery := `
		INSERT INTO post_categories (post_id, category_id)
		VALUES (?, ?)
		ON CONFLICT DO NOTHING`

	stmt, err := tx.PrepareContext(ctx, insertCategoryQuery)
	if err != nil {
		return fmt.Errorf("CreatePost: preparing post_categories statement: %w", err)
	}

	for _, cid := range categoryIDs {
		if _, err = stmt.ExecContext(ctx, post.ID, cid); err != nil {
			stmt.Close()
			return fmt.Errorf("CreatePost: inserting post_category link for category %s: %w", cid, err)
		}
	}

	if closeErr := stmt.Close(); closeErr != nil {
		return fmt.Errorf("CreatePost: closing post_categories statement: %w", closeErr)
	}

	return nil
}

func (r *PostRepository) GetAllPosts(ctx context.Context) ([]models.Post, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT
		p.id,
		u.username,
		p.title,
		p.content,
		p.created_at,
		p.likes_count,
		p.dislikes_count,
		p.image_name,
		GROUP_CONCAT(c.name, ',')
		FROM posts p
		JOIN users u ON p.author_id = u.id
		LEFT JOIN post_categories pc ON p.id = pc.post_id
		LEFT JOIN categories c     ON pc.category_id = c.id
		WHERE p.is_deleted = 0
		GROUP BY p.id
		ORDER BY p.created_at DESC;
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []models.Post

	for rows.Next() {
		var pv models.Post
		var cats sql.NullString

		if err := rows.Scan(
			&pv.ID,
			&pv.AuthorName,
			&pv.Title,
			&pv.Content,
			&pv.CreatedAt,
			&pv.Image,
			&cats,
		); err != nil {
			return nil, err
		}

		if cats.Valid {
			pv.Categories = strings.Split(cats.String, ",")
		} else {
			pv.Categories = []string{}
		}

		posts = append(posts, pv)
	}

	return posts, nil
}

func (r *PostRepository) GetPostByID(ctx context.Context, postID string) (*models.Post, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT
		p.id,
		p.author_id,
		u.username,
		p.title,
		p.content,
		p.created_at,
		p.likes_count,
		p.dislikes_count,
		p.image_name,
		GROUP_CONCAT(c.name, ',')
		FROM posts p
		JOIN users u ON p.author_id = u.id
		LEFT JOIN post_categories pc ON p.id = pc.post_id
		LEFT JOIN categories c     ON pc.category_id = c.id
		WHERE p.id = ? AND p.is_deleted = 0
		GROUP BY p.id
	`, postID)

	var pv models.Post
	var cats sql.NullString

	if err := row.Scan(
		&pv.ID,
		&pv.AuthorID,
		&pv.AuthorName,
		&pv.Title,
		&pv.Content,
		&pv.CreatedAt,
		&pv.Image,
		&cats,
	); err != nil {
        if err == sql.ErrNoRows {
            log.Printf("GetPostByID: post with ID %s not found", postID)
            return nil, fmt.Errorf("post with ID %s not found", postID)
        }
        return nil, err
	}
	if cats.Valid {
		pv.Categories = strings.Split(cats.String, ",")
	}
	return &pv, nil
}

func (r *PostRepository) GetPostsByCategory(ctx context.Context, categoryID string) ([]models.Post, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT
		p.id,
		u.username,
		p.title,
		p.content,
		p.created_at,
		p.likes_count,
		p.dislikes_count,
		p.image_name,
		GROUP_CONCAT(c.name, ',')
		FROM posts p
		JOIN users u ON p.author_id = u.id
		JOIN post_categories pc ON p.id = pc.post_id
		LEFT JOIN categories c ON pc.category_id = c.id
		WHERE pc.category_id = ? AND p.is_deleted = 0
		GROUP BY p.id
		ORDER BY p.created_at DESC;
	`, categoryID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []models.Post
	for rows.Next() {
		var pv models.Post
		var cats sql.NullString

		if err := rows.Scan(
			&pv.ID,
			&pv.AuthorName,
			&pv.Title,
			&pv.Content,
			&pv.CreatedAt,
			&pv.Image,
			&cats,
		); err != nil {
			return nil, err
		}

		if cats.Valid {
			pv.Categories = strings.Split(cats.String, ",")
		} else {
			pv.Categories = []string{}
		}

		posts = append(posts, pv)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return posts, nil
}

func (r *PostRepository) GetUserPosts(ctx context.Context, userID string) ([]models.Post, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT
		p.id,
		u.username,
		p.title,
		p.content,
		p.created_at,
		p.likes_count,
		p.dislikes_count,
		p.image_name,
		GROUP_CONCAT(c.name, ',')
		FROM posts p
		JOIN users u ON p.author_id = u.id
		JOIN post_categories pc ON p.id = pc.post_id
		LEFT JOIN categories c ON pc.category_id = c.id
		WHERE p.author_id = ? AND p.is_deleted = 0
		GROUP BY p.id
		ORDER BY p.created_at DESC;
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []models.Post
	for rows.Next() {
		var pv models.Post
		var cats sql.NullString

		if err := rows.Scan(
			&pv.ID,
			&pv.AuthorName,
			&pv.Title,
			&pv.Content,
			&pv.CreatedAt,
			&pv.Image,
			&cats,
		); err != nil {
			return nil, err
		}

		if cats.Valid {
			pv.Categories = strings.Split(cats.String, ",")
		} else {
			pv.Categories = []string{}
		}
		posts = append(posts, pv)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return posts, nil
}


func (r *PostRepository) ListByAuthor(ctx context.Context, authorID string, limit, offset int) ([]models.Post, error) {
	const query = `
		SELECT id, title, created_at, updated_at
        FROM posts
        WHERE author_id = $1 AND is_deleted = 0
        ORDER BY created_at DESC
        LIMIT $2 OFFSET $3;
    `
	rows, err := r.db.QueryContext(ctx, query, authorID, limit, offset)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var results []models.Post
	for rows.Next() {
		var p models.Post
		if err := rows.Scan(&p.ID, &p.Title, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		results = append(results, p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return results, nil
}


