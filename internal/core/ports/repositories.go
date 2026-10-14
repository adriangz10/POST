package ports

import (
	"context"
	"database/sql"
	"post/internal/core/domain"
)

type CategoryRepository interface {
	Create(ctx context.Context, cat *domain.Categoria) (*domain.Categoria, error)
	FindAll(ctx context.Context, search string) ([]domain.Categoria, error)
	FindByID(ctx context.Context, id int64) (*domain.Categoria, error)
	FindByName(ctx context.Context, name string) (*domain.Categoria, error)
	Update(ctx context.Context, id int64, cat *domain.Categoria) (*domain.Categoria, error)
	Delete(ctx context.Context, id int64) error
	CountProducts(ctx context.Context, id int64) (int64, error)
	CountCombos(ctx context.Context, id int64) (int64, error)
}

type ProviderRepository interface {
	Create(ctx context.Context, prov *domain.Proveedor) (*domain.Proveedor, error)
	FindAll(ctx context.Context, search string) ([]domain.Proveedor, error)
	FindByID(ctx context.Context, id int64) (*domain.Proveedor, error)
	FindByName(ctx context.Context, name string) (*domain.Proveedor, error)
	Update(ctx context.Context, id int64, prov *domain.Proveedor) (*domain.Proveedor, error)
	Delete(ctx context.Context, id int64) error
	FindProducts(ctx context.Context, providerID int64) ([]domain.Producto, error)
}

type ProductRepository interface {
	Create(ctx context.Context, p *domain.Producto) (*domain.Producto, error)
	CreateTx(ctx context.Context, tx *sql.Tx, p *domain.Producto) (*domain.Producto, error)
	FindAll(ctx context.Context, filter domain.ProductoFilter) (*domain.PagedResult[domain.Producto], error)
	FindAllUnpaged(ctx context.Context, filter domain.ProductoFilter) ([]domain.Producto, error)
	FindByID(ctx context.Context, id int64) (*domain.Producto, error)
	FindByCode(ctx context.Context, code string) (*domain.Producto, error)
	Update(ctx context.Context, id int64, p *domain.Producto) (*domain.Producto, error)
	UpdateTx(ctx context.Context, tx *sql.Tx, id int64, p *domain.Producto) (*domain.Producto, error)
	Delete(ctx context.Context, id int64) error
	GetProductsForPriceUpdate(ctx context.Context, alcance domain.PreciosAlcance, id *int64) ([]domain.Producto, error)
	BulkUpdatePrices(ctx context.Context, operacion domain.PreciosOperacion, porcentaje float64, alcance domain.PreciosAlcance, id *int64) (int64, error)
	BeginTx(ctx context.Context) (*sql.Tx, error)
}

type ComboRepository interface {
	Create(ctx context.Context, combo *domain.Combo) (*domain.Combo, error)
	FindAll(ctx context.Context, filter domain.ComboFilter) (*domain.PagedResult[domain.Combo], error)
	FindByID(ctx context.Context, id int64) (*domain.Combo, error)
	FindByCode(ctx context.Context, code string) (*domain.Combo, error)
	Update(ctx context.Context, id int64, combo *domain.Combo) (*domain.Combo, error)
	Delete(ctx context.Context, id int64) error
	GetComponents(ctx context.Context, comboID int64) ([]domain.ComboProducto, error)
	AddComponent(ctx context.Context, comboID int64, prodID int64, cantidad float64) error
	UpdateComponent(ctx context.Context, comboID int64, prodID int64, cantidad float64) error
	RemoveComponent(ctx context.Context, comboID int64, prodID int64) error
	UpdateStoredCost(ctx context.Context, comboID int64, cost float64) error
	FindCombosByProductID(ctx context.Context, prodID int64) ([]int64, error)
	BeginTx(ctx context.Context) (*sql.Tx, error)
}
