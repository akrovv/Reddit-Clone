package rest

import (
	"github.com/akrovv/redditclone/internal/domain"
	"github.com/akrovv/redditclone/internal/service"
)

type UserService interface {
	Get(dto *service.GetUser) (*domain.User, error)
	Save(dto *service.SaveUser) (*domain.User, error)
}

type PostService interface {
	Save(post *domain.Post) (*domain.Post, error)
	Get() ([]*domain.Post, error)
	GetOne(dto *service.GetOnePost) (*domain.Post, error)
	GetBy(dto *service.GetByPost) ([]*domain.Post, error)
	Delete(dto *service.DeletePost) error
	IncrViews(dto *service.IncrViewsPost) error
	UpdateMetrics(dto *service.UpdateMetricsPost) error
}

type CommentService interface {
	Add(dto *service.AddComment) error
	Delete(dto *service.DeleteComment) error
}

type SessionService interface {
	Create(dto *service.CreateSession) (*domain.Session, error)
	Get(dto *service.GetSession) (*domain.User, error)
}
