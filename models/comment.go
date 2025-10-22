package models

import (
	"time"
)

type Comment struct {
	ID            string
	PostID        string
	AuthorID      string
	PostTitle     string
	AuthorName    string
	Content       string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}


