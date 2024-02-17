package service

import (
	"github.com/akrovv/redditclone/internal/domain"
)

type PostStorage interface {
	Save(post *domain.Post) (*domain.Post, error)
	GetOne(id string) (*domain.Post, error)
	Get() ([]*domain.Post, error)
	GetBy(category, data, sortField string) ([]*domain.Post, error)
	UpdateMetrics(postID string, inc int8, authorID string) error
	IncrViews(postID string) error
	Delete(postID string) error
}

type UserStorage interface {
	Get(username, password string) (*domain.User, error)
	Save(username, password string) (*domain.User, error)
}

type SessionStorage interface {
	Create(ID, username string) (*domain.Session, error)
	Get(key string) (*domain.User, error)
}

type CommentStorage interface {
	Add(author *domain.Profile, body, postID string) error
	Delete(postID, commentID string) error
}
