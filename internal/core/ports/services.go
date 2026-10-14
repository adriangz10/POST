package ports

import (
	"context"
	"io"
	"post/internal/core/domain"
)

type CategoryService interface {
	Create(ctx context.Context, dto domain.CreateCategoriaDTO) (*domain.Categoria, error)
	GetAll(ctx context.Context, search string) ([]domain.Categoria, error)
	GetByID(ctx context.Context, id int64) (*domain.Categoria, error)
	Update(ctx context.Context, id int64, dto domain.UpdateCategoriaDTO) (*domain.Categoria, error)
	Delete(ctx context.Context, id int64) error
}

type ProviderService interface {
	Create(ctx context.Context, dto domain.CreateProveedorDTO) (*domain.Proveedor, error)
	GetAll(ctx context.Context, search string) ([]domain.Proveedor, error)
	GetByID(ctx context.Context, id int64) (*domain.Proveedor, error)
	Update(ctx context.Context, id int64, dto domain.UpdateProveedorDTO) (*domain.Proveedor, error)
	Delete(ctx context.Context, id int64) error
	GetProducts(ctx context.Context, providerID int64) ([]domain.Producto, error)
}

type ProductService interface {
	Create(ctx context.Context, dto domain.CreateProductoDTO) (*domain.Producto, error)
	GetAll(ctx context.Context, filter domain.ProductoFilter) (*domain.PagedResult[domain.Producto], error)
	GetByID(ctx context.Context, id int64) (*domain.Producto, error)
	Update(ctx context.Context, id int64, dto domain.UpdateProductoDTO) (*domain.Producto, error)
	Delete(ctx context.Context, id int64) error
}

type ComboService interface {
	Create(ctx context.Context, dto domain.CreateComboDTO) (*domain.Combo, error)
	GetAll(ctx context.Context, filter domain.ComboFilter) (*domain.PagedResult[domain.Combo], error)
	GetByID(ctx context.Context, id int64) (*domain.Combo, error)
	Update(ctx context.Context, id int64, dto domain.UpdateComboDTO) (*domain.Combo, error)
	Delete(ctx context.Context, id int64) error
	AddComponent(ctx context.Context, comboID int64, dto domain.ComboComponentInputDTO) (*domain.Combo, error)
	UpdateComponent(ctx context.Context, comboID int64, prodID int64, dto domain.ComboComponentUpdateDTO) (*domain.Combo, error)
	RemoveComponent(ctx context.Context, comboID int64, prodID int64) (*domain.Combo, error)
	RecalculateCombosForProduct(ctx context.Context, productID int64) error
}

type PriceService interface {
	Preview(ctx context.Context, dto domain.ActualizacionPreciosDTO) (*domain.PreciosPreviewResult, error)
	Update(ctx context.Context, dto domain.ActualizacionPreciosDTO) (*domain.PreciosUpdateResult, error)
}

type CSVService interface {
	Import(ctx context.Context, r io.Reader, modo domain.CSVImportMode, previsualizar bool) (*domain.CSVImportResult, error)
	Export(ctx context.Context, w io.Writer, filter domain.ProductoFilter) error
}
