package sqlite

import (
	"context"
	"database/sql"
	"real-time-forum/models"
	
)

type CategoriesRepository struct {
	db *sql.DB
}

func NewCategoriesRepository(db *sql.DB) *CategoriesRepository {
	return &CategoriesRepository{db: db}
}

func (c *CategoriesRepository) CreateCategory(ctx context.Context, category models.Category) error {
	_, err := c.db.ExecContext(ctx, "INSERT INTO categories (id, name) VALUES (?,?)", category.ID, category.Name)
	if err != nil {
		return err
	}

	return nil
}


func (p *CategoriesRepository) GetAllCategories(ctx context.Context) ([]models.Category, error) {
	rows, err := p.db.QueryContext(ctx, "SELECT id, name, is_deleted FROM categories")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []models.Category
	for rows.Next() {
		var category models.Category
		if err := rows.Scan(&category.ID, &category.Name); err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return categories, nil
}

