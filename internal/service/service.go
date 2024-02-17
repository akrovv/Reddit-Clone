package service

import (
	"github.com/akrovv/redditclone/internal/domain"
)

// Post
type GetOnePost struct {
	PostID string
}

type GetByPost struct {
	Category  string
	Data      string
	SortField string
}

type UpdateMetricsPost struct {
	PostID   string
	Inc      int8
	AuthorID string
}

type IncrViewsPost struct {
	PostID string
}

type DeletePost struct {
	PostID string
}

// User
type GetUser struct {
	Username string `json:"username"`
	Password string `json:"password"`
	ID       string `json:"id"`
}

type SaveUser struct {
	Username string `json:"username"`
	Password string `json:"password"`
	ID       string `json:"id"`
}

// Session
type CreateSession struct {
	ID       string
	Username string
}

type GetSession struct {
	Key string
}

// Comment
type AddComment struct {
	User   *domain.User
	Body   string
	PostID string
}

type DeleteComment struct {
	PostID    string
	CommentID string
}
