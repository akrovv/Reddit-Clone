package service

import (
	"github.com/akrovv/redditclone/internal/domain"
)

type userService struct {
	storage UserStorage
}

func NewUserService(storage UserStorage) *userService {
	return &userService{storage: storage}
}

func (s userService) Get(dto *GetUser) (*domain.User, error) {
	return s.storage.Get(dto.Username, dto.Password)
}

func (s userService) Save(dto *SaveUser) (*domain.User, error) {
	return s.storage.Save(dto.Username, dto.Password)
}
