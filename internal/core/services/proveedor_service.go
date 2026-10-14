package services

import (
	"context"
	"post/internal/core/domain"
	"post/internal/core/ports"
)

type providerService struct {
	repo ports.ProviderRepository
}

func NewProviderService(repo ports.ProviderRepository) ports.ProviderService {
	return &providerService{repo: repo}
}

func (s *providerService) Create(ctx context.Context, dto domain.CreateProveedorDTO) (*domain.Proveedor, error) {
	if err := dto.Validate(); err != nil {
		return nil, err
	}

	prov := &domain.Proveedor{
		Nombre: dto.Nombre,
	}

	return s.repo.Create(ctx, prov)
}

func (s *providerService) GetAll(ctx context.Context, search string) ([]domain.Proveedor, error) {
	return s.repo.FindAll(ctx, search)
}

func (s *providerService) GetByID(ctx context.Context, id int64) (*domain.Proveedor, error) {
	if id <= 0 {
		return nil, domain.ErrNotFound
	}
	return s.repo.FindByID(ctx, id)
}

func (s *providerService) Update(ctx context.Context, id int64, dto domain.UpdateProveedorDTO) (*domain.Proveedor, error) {
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

	return s.repo.Update(ctx, id, existing)
}

func (s *providerService) Delete(ctx context.Context, id int64) error {
	if id <= 0 {
		return domain.ErrNotFound
	}
	return s.repo.Delete(ctx, id)
}

func (s *providerService) GetProducts(ctx context.Context, providerID int64) ([]domain.Producto, error) {
	if providerID <= 0 {
		return nil, domain.ErrNotFound
	}
	// Verificar que el proveedor exista
	if _, err := s.repo.FindByID(ctx, providerID); err != nil {
		return nil, err
	}
	return s.repo.FindProducts(ctx, providerID)
}
