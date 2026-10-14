package services

import (
	"context"
	"post/internal/core/domain"
	"post/internal/core/ports"
)

type productService struct {
	productRepo  ports.ProductRepository
	categoryRepo ports.CategoryRepository
	providerRepo ports.ProviderRepository
	comboRepo    ports.ComboRepository
}

func NewProductService(
	productRepo ports.ProductRepository,
	categoryRepo ports.CategoryRepository,
	providerRepo ports.ProviderRepository,
	comboRepo ports.ComboRepository,
) ports.ProductService {
	return &productService{
		productRepo:  productRepo,
		categoryRepo: categoryRepo,
		providerRepo: providerRepo,
		comboRepo:    comboRepo,
	}
}

func (s *productService) Create(ctx context.Context, dto domain.CreateProductoDTO) (*domain.Producto, error) {
	if err := dto.Validate(); err != nil {
		return nil, err
	}

	// Verificar categoría si se proporcionó
	if dto.CategoriaID != nil && *dto.CategoriaID > 0 {
		if _, err := s.categoryRepo.FindByID(ctx, *dto.CategoriaID); err != nil {
			return nil, &domain.ValidationError{Field: "categoria_id", Message: "la categoría especificada no existe"}
		}
	}

	// Verificar proveedor si se proporcionó
	if dto.ProveedorID != nil && *dto.ProveedorID > 0 {
		if _, err := s.providerRepo.FindByID(ctx, *dto.ProveedorID); err != nil {
			return nil, &domain.ValidationError{Field: "proveedor_id", Message: "el proveedor especificado no existe"}
		}
	}

	prod := &domain.Producto{
		Nombre:      dto.Nombre,
		Codigo:      dto.Codigo,
		Descripcion: dto.Descripcion,
		CategoriaID: dto.CategoriaID,
		Unidad:      dto.Unidad,
		CostoUnidad: dto.CostoUnidad,
		VentaUnidad: dto.VentaUnidad,
		Costo:       dto.Costo,
		PrecioVenta: dto.PrecioVenta,
		Stock:       dto.Stock,
		ProveedorID: dto.ProveedorID,
		Vencimiento: dto.Vencimiento,
		Imagen:      dto.Imagen,
		Estado:      dto.Estado,
	}

	return s.productRepo.Create(ctx, prod)
}

func (s *productService) GetAll(ctx context.Context, filter domain.ProductoFilter) (*domain.PagedResult[domain.Producto], error) {
	return s.productRepo.FindAll(ctx, filter)
}

func (s *productService) GetByID(ctx context.Context, id int64) (*domain.Producto, error) {
	if id <= 0 {
		return nil, domain.ErrNotFound
	}
	return s.productRepo.FindByID(ctx, id)
}

func (s *productService) Update(ctx context.Context, id int64, dto domain.UpdateProductoDTO) (*domain.Producto, error) {
	if id <= 0 {
		return nil, domain.ErrNotFound
	}
	if err := dto.Validate(); err != nil {
		return nil, err
	}

	existing, err := s.productRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	oldCosto := existing.Costo
	oldCostoUnidad := existing.CostoUnidad
	oldUnidad := existing.Unidad

	if dto.Nombre != nil {
		existing.Nombre = *dto.Nombre
	}
	if dto.Codigo != nil {
		existing.Codigo = *dto.Codigo
	}
	if dto.Descripcion != nil {
		existing.Descripcion = dto.Descripcion
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
	if dto.Unidad != nil {
		existing.Unidad = *dto.Unidad
	}
	if dto.CostoUnidad != nil {
		existing.CostoUnidad = *dto.CostoUnidad
	}
	if dto.VentaUnidad != nil {
		existing.VentaUnidad = *dto.VentaUnidad
	}
	if dto.Costo != nil {
		existing.Costo = *dto.Costo
	}
	if dto.PrecioVenta != nil {
		existing.PrecioVenta = *dto.PrecioVenta
	}
	if dto.Stock != nil {
		existing.Stock = *dto.Stock
	}
	if dto.ProveedorID != nil {
		if *dto.ProveedorID > 0 {
			if _, err := s.providerRepo.FindByID(ctx, *dto.ProveedorID); err != nil {
				return nil, &domain.ValidationError{Field: "proveedor_id", Message: "el proveedor especificado no existe"}
			}
			existing.ProveedorID = dto.ProveedorID
		} else {
			existing.ProveedorID = nil
		}
	}
	if dto.Vencimiento != nil {
		existing.Vencimiento = dto.Vencimiento
	}
	if dto.Imagen != nil {
		existing.Imagen = dto.Imagen
	}
	if dto.Estado != nil {
		existing.Estado = *dto.Estado
	}

	updated, err := s.productRepo.Update(ctx, id, existing)
	if err != nil {
		return nil, err
	}

	// Si el costo o las unidades cambiaron, sincronizar y recalcular el costo en los combos que contienen este producto
	if s.comboRepo != nil && (oldCosto != updated.Costo || oldCostoUnidad != updated.CostoUnidad || oldUnidad != updated.Unidad) {
		comboIDs, err := s.comboRepo.FindCombosByProductID(ctx, id)
		if err == nil {
			for _, cid := range comboIDs {
				components, err := s.comboRepo.GetComponents(ctx, cid)
				if err == nil {
					totalCost := domain.CalculateTotalCost(components)
					_ = s.comboRepo.UpdateStoredCost(ctx, cid, totalCost)
				}
			}
		}
	}

	return updated, nil
}

func (s *productService) Delete(ctx context.Context, id int64) error {
	if id <= 0 {
		return domain.ErrNotFound
	}

	// Si está en combo_productos, repo.Delete devolverá ErrConflict debido a ON DELETE RESTRICT
	return s.productRepo.Delete(ctx, id)
}
