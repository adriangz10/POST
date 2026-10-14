package postgres

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/lib/pq"
	"post/internal/core/domain"
	"post/internal/core/ports"
)

type providerRepository struct {
	db *sql.DB
}

func NewProviderRepository(db *sql.DB) ports.ProviderRepository {
	return &providerRepository{db: db}
}

func (r *providerRepository) Create(ctx context.Context, prov *domain.Proveedor) (*domain.Proveedor, error) {
	query := `INSERT INTO proveedores (nombre) VALUES ($1) RETURNING id`
	err := r.db.QueryRowContext(ctx, query, prov.Nombre).Scan(&prov.ID)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			if pqErr.Code == "23505" {
				return nil, domain.ErrDuplicateKey
			}
		}
		return nil, err
	}
	return prov, nil
}

func (r *providerRepository) FindAll(ctx context.Context, search string) ([]domain.Proveedor, error) {
	var query string
	var args []interface{}

	search = strings.TrimSpace(search)
	if search != "" {
		query = `SELECT id, nombre FROM proveedores WHERE LOWER(nombre) LIKE LOWER($1) ORDER BY nombre ASC`
		args = append(args, "%"+search+"%")
	} else {
		query = `SELECT id, nombre FROM proveedores ORDER BY nombre ASC`
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var providers []domain.Proveedor
	for rows.Next() {
		var p domain.Proveedor
		if err := rows.Scan(&p.ID, &p.Nombre); err != nil {
			return nil, err
		}
		providers = append(providers, p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return providers, nil
}

func (r *providerRepository) FindByID(ctx context.Context, id int64) (*domain.Proveedor, error) {
	query := `SELECT id, nombre FROM proveedores WHERE id = $1`
	var p domain.Proveedor
	err := r.db.QueryRowContext(ctx, query, id).Scan(&p.ID, &p.Nombre)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &p, nil
}

func (r *providerRepository) FindByName(ctx context.Context, name string) (*domain.Proveedor, error) {
	query := `SELECT id, nombre FROM proveedores WHERE LOWER(nombre) = LOWER($1)`
	var p domain.Proveedor
	err := r.db.QueryRowContext(ctx, query, name).Scan(&p.ID, &p.Nombre)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &p, nil
}

func (r *providerRepository) Update(ctx context.Context, id int64, prov *domain.Proveedor) (*domain.Proveedor, error) {
	query := `UPDATE proveedores SET nombre = $1 WHERE id = $2 RETURNING id, nombre`
	var updated domain.Proveedor
	err := r.db.QueryRowContext(ctx, query, prov.Nombre, id).Scan(&updated.ID, &updated.Nombre)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			if pqErr.Code == "23505" {
				return nil, domain.ErrDuplicateKey
			}
		}
		return nil, err
	}
	return &updated, nil
}

func (r *providerRepository) Delete(ctx context.Context, id int64) error {
	// Nota: según la FK ON DELETE SET NULL, en PostgreSQL al eliminar el proveedor sus productos quedan con proveedor_id = NULL
	query := `DELETE FROM proveedores WHERE id = $1`
	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *providerRepository) FindProducts(ctx context.Context, providerID int64) ([]domain.Producto, error) {
	query := `
		SELECT p.id, p.nombre, p.codigo, p.descripcion, p.categoria_id, c.nombre AS categoria_nombre,
		       p.unidad, p.costo_unidad, p.venta_unidad, p.costo, p.precio_venta, p.stock,
		       p.proveedor_id, pr.nombre AS proveedor_nombre, TO_CHAR(p.vencimiento, 'YYYY-MM-DD') AS vencimiento,
		       p.imagen, p.estado, p.creado_en, p.actualizado_en
		FROM productos p
		LEFT JOIN categorias c ON p.categoria_id = c.id
		LEFT JOIN proveedores pr ON p.proveedor_id = pr.id
		WHERE p.proveedor_id = $1
		ORDER BY p.nombre ASC
	`
	rows, err := r.db.QueryContext(ctx, query, providerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []domain.Producto
	for rows.Next() {
		var prod domain.Producto
		var desc, venc, img, catNom, provNom sql.NullString
		var catID, provID sql.NullInt64

		err := rows.Scan(
			&prod.ID, &prod.Nombre, &prod.Codigo, &desc, &catID, &catNom,
			&prod.Unidad, &prod.CostoUnidad, &prod.VentaUnidad, &prod.Costo, &prod.PrecioVenta, &prod.Stock,
			&provID, &provNom, &venc, &img, &prod.Estado, &prod.CreadoEn, &prod.ActualizadoEn,
		)
		if err != nil {
			return nil, err
		}
		if desc.Valid {
			prod.Descripcion = &desc.String
		}
		if catID.Valid {
			prod.CategoriaID = &catID.Int64
		}
		if catNom.Valid {
			prod.CategoriaNombre = &catNom.String
		}
		if provID.Valid {
			prod.ProveedorID = &provID.Int64
		}
		if provNom.Valid {
			prod.ProveedorNombre = &provNom.String
		}
		if venc.Valid {
			prod.Vencimiento = &venc.String
		}
		if img.Valid {
			prod.Imagen = &img.String
		}
		products = append(products, prod)
	}
	return products, nil
}
