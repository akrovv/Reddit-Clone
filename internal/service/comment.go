package service

import "github.com/akrovv/redditclone/internal/domain"

type commentService struct {
	storage CommentStorage
}

func NewCommentService(storage CommentStorage) *commentService {
	return &commentService{storage: storage}
}

func (s commentService) Add(dto *AddComment) error {
	author := &domain.Profile{Username: dto.User.Password, ID: dto.User.ID}
	return s.storage.Add(author, dto.Body, dto.PostID)
}

func (s commentService) Delete(dto *DeleteComment) error {
	return s.storage.Delete(dto.PostID, dto.CommentID)
}
