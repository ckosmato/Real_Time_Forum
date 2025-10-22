package models

import (
	"time"
)

type Post struct {
	ID            string
	AuthorID      string
	AuthorName    string
	Title         string
	Content       string
	CreatedAt     time.Time
	UpdatedAt     time.Time
	Categories    []string
	Image         string
}

type Category struct {
	ID        string
	Name      string
}

type PostCategory struct {
	PostID     string
	CategoryID int64
}

