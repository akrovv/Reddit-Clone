package service

import "github.com/akrovv/redditclone/internal/domain"

type postService struct {
	storage PostStorage
}

func NewPostService(storage PostStorage) *postService {
	return &postService{storage: storage}
}

func (s postService) Save(post *domain.Post) (*domain.Post, error) {
	return s.storage.Save(post)
}

func (s postService) Get() ([]*domain.Post, error) {
	return s.storage.Get()
}

func (s postService) GetOne(dto *GetOnePost) (*domain.Post, error) {
	return s.storage.GetOne(dto.PostID)
}

func (s postService) GetBy(dto *GetByPost) ([]*domain.Post, error) {
	return s.storage.GetBy(dto.Category, dto.Data, dto.SortField)
}

func (s postService) Delete(dto *DeletePost) error {
	return s.storage.Delete(dto.PostID)
}

func (s postService) IncrViews(dto *IncrViewsPost) error {
	return s.storage.IncrViews(dto.PostID)
}

func (s postService) UpdateMetrics(dto *UpdateMetricsPost) error {
	return s.storage.UpdateMetrics(dto.PostID, dto.Inc, dto.AuthorID)
}
