package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/lib/pq"
	"post/internal/core/domain"
	"post/internal/core/ports"
)

type productRepository struct {
	db *sql.DB
}

func NewProductRepository(db *sql.DB) ports.ProductRepository {
	return &productRepository{db: db}
}

func (r *productRepository) BeginTx(ctx context.Context) (*sql.Tx, error) {
	return r.db.BeginTx(ctx, nil)
}

func (r *productRepository) Create(ctx context.Context, p *domain.Producto) (*domain.Producto, error) {
	return r.createInternal(ctx, r.db, p)
}

func (r *productRepository) CreateTx(ctx context.Context, tx *sql.Tx, p *domain.Producto) (*domain.Producto, error) {
	return r.createInternal(ctx, tx, p)
}

type queryExecutor interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

func (r *productRepository) createInternal(ctx context.Context, exec queryExecutor, p *domain.Producto) (*domain.Producto, error) {
	query := `
		INSERT INTO productos (
			nombre, codigo, descripcion, categoria_id, unidad, costo_unidad, venta_unidad,
			costo, precio_venta, stock, proveedor_id, vencimiento, imagen, estado
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14
		) RETURNING id, creado_en, actualizado_en
	`
	err := exec.QueryRowContext(
		ctx, query,
		p.Nombre, p.Codigo, p.Descripcion, p.CategoriaID, p.Unidad, p.CostoUnidad, p.VentaUnidad,
		p.Costo, p.PrecioVenta, p.Stock, p.ProveedorID, p.Vencimiento, p.Imagen, p.Estado,
	).Scan(&p.ID, &p.CreadoEn, &p.ActualizadoEn)

	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			if pqErr.Code == "23505" {
				return nil, domain.ErrDuplicateKey
			}
			if pqErr.Code == "23503" {
				return nil, domain.ErrForeignKeyViolated
			}
			if pqErr.Code == "23514" {
				return nil, domain.ErrInvalidInput
			}
		}
		return nil, err
	}
	return p, nil
}

func (r *productRepository) buildFilterQuery(filter domain.ProductoFilter) (string, []interface{}) {
	var conditions []string
	var args []interface{}
	idx := 1

	if filter.CategoriaID != nil {
		conditions = append(conditions, "p.categoria_id = $"+strconv.Itoa(idx))
		args = append(args, *filter.CategoriaID)
		idx++
	}

	if filter.ProveedorID != nil {
		conditions = append(conditions, "p.proveedor_id = $"+strconv.Itoa(idx))
		args = append(args, *filter.ProveedorID)
		idx++
	}

	if filter.Estado != nil && *filter.Estado != "" {
		conditions = append(conditions, "p.estado = $"+strconv.Itoa(idx))
		args = append(args, *filter.Estado)
		idx++
	}

	if filter.Buscar != nil && strings.TrimSpace(*filter.Buscar) != "" {
		searchTerm := "%" + strings.TrimSpace(*filter.Buscar) + "%"
		conditions = append(conditions, "(LOWER(p.nombre) LIKE LOWER($"+strconv.Itoa(idx)+") OR LOWER(p.codigo) LIKE LOWER($"+strconv.Itoa(idx)+"))")
		args = append(args, searchTerm)
		idx++
	}

	if filter.VencimientoDesde != nil && *filter.VencimientoDesde != "" {
		conditions = append(conditions, "p.vencimiento >= $"+strconv.Itoa(idx))
		args = append(args, *filter.VencimientoDesde)
		idx++
	}

	if filter.VencimientoHasta != nil && *filter.VencimientoHasta != "" {
		conditions = append(conditions, "p.vencimiento <= $"+strconv.Itoa(idx))
		args = append(args, *filter.VencimientoHasta)
		idx++
	}

	if filter.StockMin != nil {
		conditions = append(conditions, "p.stock >= $"+strconv.Itoa(idx))
		args = append(args, *filter.StockMin)
		idx++
	}

	if filter.StockMax != nil {
		conditions = append(conditions, "p.stock <= $"+strconv.Itoa(idx))
		args = append(args, *filter.StockMax)
		idx++
	}

	if filter.PrecioMin != nil {
		conditions = append(conditions, "p.precio_venta >= $"+strconv.Itoa(idx))
		args = append(args, *filter.PrecioMin)
		idx++
	}

	if filter.PrecioMax != nil {
		conditions = append(conditions, "p.precio_venta <= $"+strconv.Itoa(idx))
		args = append(args, *filter.PrecioMax)
		idx++
	}

	if filter.StockBajo != nil && *filter.StockBajo {
		umbral := 5.0
		if filter.UmbralStockBajo != nil && *filter.UmbralStockBajo >= 0 {
			umbral = *filter.UmbralStockBajo
		}
		conditions = append(conditions, "p.stock <= $"+strconv.Itoa(idx))
		args = append(args, umbral)
		idx++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}
	return whereClause, args
}

func (r *productRepository) FindAll(ctx context.Context, filter domain.ProductoFilter) (*domain.PagedResult[domain.Producto], error) {
	whereClause, args := r.buildFilterQuery(filter)

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM productos p %s", whereClause)
	var total int64
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, err
	}

	// Ordenamiento
	validSortColumns := map[string]string{
		"id":           "p.id",
		"nombre":       "p.nombre",
		"codigo":       "p.codigo",
		"precio_venta": "p.precio_venta",
		"costo":        "p.costo",
		"stock":        "p.stock",
		"vencimiento":  "p.vencimiento",
		"creado_en":    "p.creado_en",
	}
	orderBy := "p.nombre"
	if col, ok := validSortColumns[strings.ToLower(filter.OrdenPor)]; ok {
		orderBy = col
	}

	dir := "ASC"
	if strings.ToUpper(filter.Direccion) == "DESC" {
		dir = "DESC"
	}

	limit := filter.Limite
	if limit <= 0 {
		limit = 20
	}
	if limit > 5000 {
		limit = 5000
	}

	page := filter.Pagina
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * limit

	query := fmt.Sprintf(`
		SELECT p.id, p.nombre, p.codigo, p.descripcion, p.categoria_id, c.nombre AS categoria_nombre, c.color AS categoria_color,
		       p.unidad, p.costo_unidad, p.venta_unidad, p.costo, p.precio_venta, p.stock,
		       p.proveedor_id, pr.nombre AS proveedor_nombre, TO_CHAR(p.vencimiento, 'YYYY-MM-DD') AS vencimiento,
		       p.imagen, p.estado, p.creado_en, p.actualizado_en
		FROM productos p
		LEFT JOIN categorias c ON p.categoria_id = c.id
		LEFT JOIN proveedores pr ON p.proveedor_id = pr.id
		%s
		ORDER BY %s %s
		LIMIT $%d OFFSET $%d
	`, whereClause, orderBy, dir, len(args)+1, len(args)+2)

	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []domain.Producto
	for rows.Next() {
		var prod domain.Producto
		var desc, venc, img, catNom, catColor, provNom sql.NullString
		var catID, provID sql.NullInt64

		err := rows.Scan(
			&prod.ID, &prod.Nombre, &prod.Codigo, &desc, &catID, &catNom, &catColor,
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
		if catColor.Valid {
			prod.CategoriaColor = &catColor.String
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

	totalPaginas := int(total) / limit
	if int(total)%limit != 0 {
		totalPaginas++
	}

	return &domain.PagedResult[domain.Producto]{
		Data:         products,
		Total:        total,
		Pagina:       page,
		Limite:       limit,
		TotalPaginas: totalPaginas,
	}, nil
}

func (r *productRepository) FindAllUnpaged(ctx context.Context, filter domain.ProductoFilter) ([]domain.Producto, error) {
	whereClause, args := r.buildFilterQuery(filter)

	query := fmt.Sprintf(`
		SELECT p.id, p.nombre, p.codigo, p.descripcion, p.categoria_id, c.nombre AS categoria_nombre, c.color AS categoria_color,
		       p.unidad, p.costo_unidad, p.venta_unidad, p.costo, p.precio_venta, p.stock,
		       p.proveedor_id, pr.nombre AS proveedor_nombre, TO_CHAR(p.vencimiento, 'YYYY-MM-DD') AS vencimiento,
		       p.imagen, p.estado, p.creado_en, p.actualizado_en
		FROM productos p
		LEFT JOIN categorias c ON p.categoria_id = c.id
		LEFT JOIN proveedores pr ON p.proveedor_id = pr.id
		%s
		ORDER BY p.nombre ASC
	`, whereClause)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []domain.Producto
	for rows.Next() {
		var prod domain.Producto
		var desc, venc, img, catNom, catColor, provNom sql.NullString
		var catID, provID sql.NullInt64

		err := rows.Scan(
			&prod.ID, &prod.Nombre, &prod.Codigo, &desc, &catID, &catNom, &catColor,
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
		if catColor.Valid {
			prod.CategoriaColor = &catColor.String
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

func (r *productRepository) FindByID(ctx context.Context, id int64) (*domain.Producto, error) {
	query := `
		SELECT p.id, p.nombre, p.codigo, p.descripcion, p.categoria_id, c.nombre AS categoria_nombre, c.color AS categoria_color,
		       p.unidad, p.costo_unidad, p.venta_unidad, p.costo, p.precio_venta, p.stock,
		       p.proveedor_id, pr.nombre AS proveedor_nombre, TO_CHAR(p.vencimiento, 'YYYY-MM-DD') AS vencimiento,
		       p.imagen, p.estado, p.creado_en, p.actualizado_en
		FROM productos p
		LEFT JOIN categorias c ON p.categoria_id = c.id
		LEFT JOIN proveedores pr ON p.proveedor_id = pr.id
		WHERE p.id = $1
	`
	var prod domain.Producto
	var desc, venc, img, catNom, catColor, provNom sql.NullString
	var catID, provID sql.NullInt64

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&prod.ID, &prod.Nombre, &prod.Codigo, &desc, &catID, &catNom, &catColor,
		&prod.Unidad, &prod.CostoUnidad, &prod.VentaUnidad, &prod.Costo, &prod.PrecioVenta, &prod.Stock,
		&provID, &provNom, &venc, &img, &prod.Estado, &prod.CreadoEn, &prod.ActualizadoEn,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
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
	if catColor.Valid {
		prod.CategoriaColor = &catColor.String
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
	return &prod, nil
}

func (r *productRepository) FindByCode(ctx context.Context, code string) (*domain.Producto, error) {
	query := `
		SELECT p.id, p.nombre, p.codigo, p.descripcion, p.categoria_id, c.nombre AS categoria_nombre,
		       p.unidad, p.costo_unidad, p.venta_unidad, p.costo, p.precio_venta, p.stock,
		       p.proveedor_id, pr.nombre AS proveedor_nombre, TO_CHAR(p.vencimiento, 'YYYY-MM-DD') AS vencimiento,
		       p.imagen, p.estado, p.creado_en, p.actualizado_en
		FROM productos p
		LEFT JOIN categorias c ON p.categoria_id = c.id
		LEFT JOIN proveedores pr ON p.proveedor_id = pr.id
		WHERE p.codigo = $1
	`
	var prod domain.Producto
	var desc, venc, img, catNom, provNom sql.NullString
	var catID, provID sql.NullInt64

	err := r.db.QueryRowContext(ctx, query, code).Scan(
		&prod.ID, &prod.Nombre, &prod.Codigo, &desc, &catID, &catNom,
		&prod.Unidad, &prod.CostoUnidad, &prod.VentaUnidad, &prod.Costo, &prod.PrecioVenta, &prod.Stock,
		&provID, &provNom, &venc, &img, &prod.Estado, &prod.CreadoEn, &prod.ActualizadoEn,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
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
	return &prod, nil
}

func (r *productRepository) Update(ctx context.Context, id int64, p *domain.Producto) (*domain.Producto, error) {
	return r.updateInternal(ctx, r.db, id, p)
}

func (r *productRepository) UpdateTx(ctx context.Context, tx *sql.Tx, id int64, p *domain.Producto) (*domain.Producto, error) {
	return r.updateInternal(ctx, tx, id, p)
}

func (r *productRepository) updateInternal(ctx context.Context, exec queryExecutor, id int64, p *domain.Producto) (*domain.Producto, error) {
	query := `
		UPDATE productos SET
			nombre = $1,
			codigo = $2,
			descripcion = $3,
			categoria_id = $4,
			unidad = $5,
			costo_unidad = $6,
			venta_unidad = $7,
			costo = $8,
			precio_venta = $9,
			stock = $10,
			proveedor_id = $11,
			vencimiento = $12,
			imagen = $13,
			estado = $14,
			actualizado_en = CURRENT_TIMESTAMP
		WHERE id = $15
		RETURNING creado_en, actualizado_en
	`
	err := exec.QueryRowContext(
		ctx, query,
		p.Nombre, p.Codigo, p.Descripcion, p.CategoriaID, p.Unidad, p.CostoUnidad, p.VentaUnidad,
		p.Costo, p.PrecioVenta, p.Stock, p.ProveedorID, p.Vencimiento, p.Imagen, p.Estado, id,
	).Scan(&p.CreadoEn, &p.ActualizadoEn)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			if pqErr.Code == "23505" {
				return nil, domain.ErrDuplicateKey
			}
			if pqErr.Code == "23503" {
				return nil, domain.ErrForeignKeyViolated
			}
			if pqErr.Code == "23514" {
				return nil, domain.ErrInvalidInput
			}
		}
		return nil, err
	}
	p.ID = id
	return p, nil
}

func (r *productRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM productos WHERE id = $1`
	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			if pqErr.Code == "23503" || pqErr.Code == "23001" {
				return domain.ErrConflict
			}
		}
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

func (r *productRepository) GetProductsForPriceUpdate(ctx context.Context, alcance domain.PreciosAlcance, id *int64) ([]domain.Producto, error) {
	var query string
	var args []interface{}

	switch alcance {
	case domain.AlcanceTodos:
		query = `SELECT id, codigo, nombre, precio_venta FROM productos ORDER BY id ASC`
	case domain.AlcanceCategoria:
		query = `SELECT id, codigo, nombre, precio_venta FROM productos WHERE categoria_id = $1 ORDER BY id ASC`
		args = append(args, *id)
	case domain.AlcanceProveedor:
		query = `SELECT id, codigo, nombre, precio_venta FROM productos WHERE proveedor_id = $1 ORDER BY id ASC`
		args = append(args, *id)
	default:
		return nil, domain.ErrInvalidInput
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []domain.Producto
	for rows.Next() {
		var p domain.Producto
		if err := rows.Scan(&p.ID, &p.Codigo, &p.Nombre, &p.PrecioVenta); err != nil {
			return nil, err
		}
		products = append(products, p)
	}
	return products, nil
}

func (r *productRepository) BulkUpdatePrices(ctx context.Context, operacion domain.PreciosOperacion, porcentaje float64, alcance domain.PreciosAlcance, id *int64) (int64, error) {
	factor := 1.0 + (porcentaje / 100.0)
	if operacion == domain.OperacionDisminucion {
		factor = 1.0 - (porcentaje / 100.0)
	}

	var query string
	var args []interface{}

	switch alcance {
	case domain.AlcanceTodos:
		query = `
			UPDATE productos
			SET precio_venta = ROUND(GREATEST(0, precio_venta * $1), 2),
			    actualizado_en = CURRENT_TIMESTAMP
		`
		args = append(args, factor)
	case domain.AlcanceCategoria:
		query = `
			UPDATE productos
			SET precio_venta = ROUND(GREATEST(0, precio_venta * $1), 2),
			    actualizado_en = CURRENT_TIMESTAMP
			WHERE categoria_id = $2
		`
		args = append(args, factor, *id)
	case domain.AlcanceProveedor:
		query = `
			UPDATE productos
			SET precio_venta = ROUND(GREATEST(0, precio_venta * $1), 2),
			    actualizado_en = CURRENT_TIMESTAMP
			WHERE proveedor_id = $2
		`
		args = append(args, factor, *id)
	default:
		return 0, domain.ErrInvalidInput
	}

	res, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}
