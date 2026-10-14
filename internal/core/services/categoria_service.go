package services

import (
	"context"
	"post/internal/core/domain"
	"post/internal/core/ports"
)

type categoryService struct {
	repo ports.CategoryRepository
}

func NewCategoryService(repo ports.CategoryRepository) ports.CategoryService {
	return &categoryService{repo: repo}
}

func (s *categoryService) Create(ctx context.Context, dto domain.CreateCategoriaDTO) (*domain.Categoria, error) {
	if err := dto.Validate(); err != nil {
		return nil, err
	}

	cat := &domain.Categoria{
		Nombre: dto.Nombre,
		Color:  dto.Color,
	}

	return s.repo.Create(ctx, cat)
}

func (s *categoryService) GetAll(ctx context.Context, search string) ([]domain.Categoria, error) {
	return s.repo.FindAll(ctx, search)
}

func (s *categoryService) GetByID(ctx context.Context, id int64) (*domain.Categoria, error) {
	if id <= 0 {
		return nil, domain.ErrNotFound
	}
	return s.repo.FindByID(ctx, id)
}

func (s *categoryService) Update(ctx context.Context, id int64, dto domain.UpdateCategoriaDTO) (*domain.Categoria, error) {
	if id <= 0 {
		return nil, domain.ErrNotFound
	}
	if err := dto.Validate(); err != nil {
		return nil, err
	}

	existing, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if dto.Nombre != nil {
		existing.Nombre = *dto.Nombre
	}
	if dto.Color != nil {
		existing.Color = *dto.Color
	}

	return s.repo.Update(ctx, id, existing)
}

func (s *categoryService) Delete(ctx context.Context, id int64) error {
	if id <= 0 {
		return domain.ErrNotFound
	}
	return s.repo.Delete(ctx, id)
}
