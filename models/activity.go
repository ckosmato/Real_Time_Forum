package models

import (
	"time"
)

type ActivityItem struct {
	Type       string
	Timestamp  time.Time
	UpdatedAt  time.Time
	UserID     string
	AuthorName string
	PostID     string
	PostTitle  string
	Value      int8
	TargetType string
	CommentID  string
	Comment    string
}

type ProfilePageData struct {
	// User      *User
	// CSRFToken string

	All       []ActivityItem
	Posts     []ActivityItem
	// CommReactions []ActivityItem
	Comments []ActivityItem

	AllPage      int
	PostsPage    int
	CommentsPage int

	PageSize         int
	PostsPageSize    int
	CommentsPageSize int
}
