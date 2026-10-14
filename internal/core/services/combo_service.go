package services

import (
	"context"
	"post/internal/core/domain"
	"post/internal/core/ports"
)

type comboService struct {
	comboRepo    ports.ComboRepository
	productRepo  ports.ProductRepository
	categoryRepo ports.CategoryRepository
}

func NewComboService(
	comboRepo ports.ComboRepository,
	productRepo ports.ProductRepository,
	categoryRepo ports.CategoryRepository,
) ports.ComboService {
	return &comboService{
		comboRepo:    comboRepo,
		productRepo:  productRepo,
		categoryRepo: categoryRepo,
	}
}

func (s *comboService) Create(ctx context.Context, dto domain.CreateComboDTO) (*domain.Combo, error) {
	if err := dto.Validate(); err != nil {
		return nil, err
	}

	if dto.CategoriaID != nil && *dto.CategoriaID > 0 {
		if _, err := s.categoryRepo.FindByID(ctx, *dto.CategoriaID); err != nil {
			return nil, &domain.ValidationError{Field: "categoria_id", Message: "la categoría especificada no existe"}
		}
	}

	// Validar que cada producto exista y calcular el costo inicial
	var calculatedCost float64
	for _, comp := range dto.Componentes {
		p, err := s.productRepo.FindByID(ctx, comp.ProductoID)
		if err != nil {
			return nil, &domain.ValidationError{Field: "componentes", Message: "el producto especificado en los componentes no existe"}
		}
		calculatedCost += p.CostPerBaseUnit() * comp.Cantidad
	}
	calculatedCost = domain.CalculateTotalCost([]domain.ComboProducto{}) // helper will be used below

	combo := &domain.Combo{
		Nombre:      dto.Nombre,
		Codigo:      dto.Codigo,
		CategoriaID: dto.CategoriaID,
		Imagen:      dto.Imagen,
		Estado:      dto.Estado,
		Costo:       0,
		PrecioVenta: dto.PrecioVenta,
	}

	created, err := s.comboRepo.Create(ctx, combo)
	if err != nil {
		return nil, err
	}

	// Insertar componentes si los hay
	for _, comp := range dto.Componentes {
		if err := s.comboRepo.AddComponent(ctx, created.ID, comp.ProductoID, comp.Cantidad); err != nil {
			return nil, err
		}
	}

	// Recalcular y persistir el costo exacto
	components, err := s.comboRepo.GetComponents(ctx, created.ID)
	if err == nil {
		totalCost := domain.CalculateTotalCost(components)
		_ = s.comboRepo.UpdateStoredCost(ctx, created.ID, totalCost)
		created.Costo = totalCost
		created.Componentes = components
	}

	return created, nil
}

func (s *comboService) GetAll(ctx context.Context, filter domain.ComboFilter) (*domain.PagedResult[domain.Combo], error) {
	return s.comboRepo.FindAll(ctx, filter)
}

func (s *comboService) GetByID(ctx context.Context, id int64) (*domain.Combo, error) {
	if id <= 0 {
		return nil, domain.ErrNotFound
	}
	combo, err := s.comboRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Verificar y asegurar sincronización de costo calculado vs almacenado
	totalCost := domain.CalculateTotalCost(combo.Componentes)
	if totalCost != combo.Costo {
		_ = s.comboRepo.UpdateStoredCost(ctx, combo.ID, totalCost)
		combo.Costo = totalCost
	}

	return combo, nil
}

func (s *comboService) Update(ctx context.Context, id int64, dto domain.UpdateComboDTO) (*domain.Combo, error) {
	if id <= 0 {
		return nil, domain.ErrNotFound
	}
	if err := dto.Validate(); err != nil {
		return nil, err
	}

	existing, err := s.comboRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if dto.Nombre != nil {
		existing.Nombre = *dto.Nombre
	}
	if dto.Codigo != nil {
		existing.Codigo = *dto.Codigo
	}
	if dto.CategoriaID != nil {
		if *dto.CategoriaID > 0 {
			if _, err := s.categoryRepo.FindByID(ctx, *dto.CategoriaID); err != nil {
				return nil, &domain.ValidationError{Field: "categoria_id", Message: "la categoría especificada no existe"}
			}
			existing.CategoriaID = dto.CategoriaID
		} else {
			existing.CategoriaID = nil
		}
	}
	if dto.Imagen != nil {
		existing.Imagen = dto.Imagen
	}
	if dto.Estado != nil {
		existing.Estado = *dto.Estado
	}
	if dto.PrecioVenta != nil {
		existing.PrecioVenta = *dto.PrecioVenta
	}

	updated, err := s.comboRepo.Update(ctx, id, existing)
	if err != nil {
		return nil, err
	}

	components, err := s.comboRepo.GetComponents(ctx, id)
	if err == nil {
		updated.Componentes = components
	}

	return updated, nil
}

func (s *comboService) Delete(ctx context.Context, id int64) error {
	if id <= 0 {
		return domain.ErrNotFound
	}
	return s.comboRepo.Delete(ctx, id)
}

func (s *comboService) AddComponent(ctx context.Context, comboID int64, dto domain.ComboComponentInputDTO) (*domain.Combo, error) {
	if comboID <= 0 {
		return nil, domain.ErrNotFound
	}
	if err := dto.Validate(); err != nil {
		return nil, err
	}

	// Verificar combo
	if _, err := s.comboRepo.FindByID(ctx, comboID); err != nil {
		return nil, err
	}

	// Verificar producto
	if _, err := s.productRepo.FindByID(ctx, dto.ProductoID); err != nil {
		return nil, &domain.ValidationError{Field: "producto_id", Message: "el producto especificado no existe"}
	}

	if err := s.comboRepo.AddComponent(ctx, comboID, dto.ProductoID, dto.Cantidad); err != nil {
		return nil, err
	}

	// Recalcular y actualizar costo almacenado
	components, err := s.comboRepo.GetComponents(ctx, comboID)
	if err != nil {
		return nil, err
	}
	totalCost := domain.CalculateTotalCost(components)
	if err := s.comboRepo.UpdateStoredCost(ctx, comboID, totalCost); err != nil {
		return nil, err
	}

	return s.GetByID(ctx, comboID)
}

func (s *comboService) UpdateComponent(ctx context.Context, comboID int64, prodID int64, dto domain.ComboComponentUpdateDTO) (*domain.Combo, error) {
	if comboID <= 0 || prodID <= 0 {
		return nil, domain.ErrNotFound
	}
	if err := dto.Validate(); err != nil {
		return nil, err
	}

	if err := s.comboRepo.UpdateComponent(ctx, comboID, prodID, dto.Cantidad); err != nil {
		return nil, err
	}

	// Recalcular y actualizar costo almacenado
	components, err := s.comboRepo.GetComponents(ctx, comboID)
	if err != nil {
		return nil, err
	}
	totalCost := domain.CalculateTotalCost(components)
	if err := s.comboRepo.UpdateStoredCost(ctx, comboID, totalCost); err != nil {
		return nil, err
	}

	return s.GetByID(ctx, comboID)
}

func (s *comboService) RemoveComponent(ctx context.Context, comboID int64, prodID int64) (*domain.Combo, error) {
	if comboID <= 0 || prodID <= 0 {
		return nil, domain.ErrNotFound
	}

	if err := s.comboRepo.RemoveComponent(ctx, comboID, prodID); err != nil {
		return nil, err
	}

	// Recalcular y actualizar costo almacenado
	components, err := s.comboRepo.GetComponents(ctx, comboID)
	if err != nil {
		return nil, err
	}
	totalCost := domain.CalculateTotalCost(components)
	if err := s.comboRepo.UpdateStoredCost(ctx, comboID, totalCost); err != nil {
		return nil, err
	}

	return s.GetByID(ctx, comboID)
}

func (s *comboService) RecalculateCombosForProduct(ctx context.Context, productID int64) error {
	comboIDs, err := s.comboRepo.FindCombosByProductID(ctx, productID)
	if err != nil {
		return err
	}
	for _, cid := range comboIDs {
		components, err := s.comboRepo.GetComponents(ctx, cid)
		if err != nil {
			continue
		}
		totalCost := domain.CalculateTotalCost(components)
		_ = s.comboRepo.UpdateStoredCost(ctx, cid, totalCost)
	}
	return nil
}
