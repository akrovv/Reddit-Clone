package service

import (
	"github.com/akrovv/redditclone/internal/domain"
)

type sessionService struct {
	storage SessionStorage
}

func NewSessionService(storage SessionStorage) *sessionService {
	return &sessionService{storage: storage}
}

func (s sessionService) Create(dto *CreateSession) (*domain.Session, error) {
	return s.storage.Create(dto.ID, dto.Username)
}

func (s sessionService) Get(dto *GetSession) (*domain.User, error) {
	return s.storage.Get(dto.Key)
}
