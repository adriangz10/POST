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

type comboRepository struct {
	db *sql.DB
}

func NewComboRepository(db *sql.DB) ports.ComboRepository {
	return &comboRepository{db: db}
}

func (r *comboRepository) BeginTx(ctx context.Context) (*sql.Tx, error) {
	return r.db.BeginTx(ctx, nil)
}

func (r *comboRepository) Create(ctx context.Context, combo *domain.Combo) (*domain.Combo, error) {
	query := `
		INSERT INTO combos (
			nombre, codigo, categoria_id, imagen, estado, costo, precio_venta
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7
		) RETURNING id, creado_en, actualizado_en
	`
	err := r.db.QueryRowContext(
		ctx, query,
		combo.Nombre, combo.Codigo, combo.CategoriaID, combo.Imagen, combo.Estado, combo.Costo, combo.PrecioVenta,
	).Scan(&combo.ID, &combo.CreadoEn, &combo.ActualizadoEn)

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
	return combo, nil
}

func (r *comboRepository) FindAll(ctx context.Context, filter domain.ComboFilter) (*domain.PagedResult[domain.Combo], error) {
	var conditions []string
	var args []interface{}
	idx := 1

	if filter.CategoriaID != nil {
		conditions = append(conditions, "cb.categoria_id = $"+strconv.Itoa(idx))
		args = append(args, *filter.CategoriaID)
		idx++
	}

	if filter.Estado != nil && *filter.Estado != "" {
		conditions = append(conditions, "cb.estado = $"+strconv.Itoa(idx))
		args = append(args, *filter.Estado)
		idx++
	}

	if filter.Buscar != nil && strings.TrimSpace(*filter.Buscar) != "" {
		searchTerm := "%" + strings.TrimSpace(*filter.Buscar) + "%"
		conditions = append(conditions, "(LOWER(cb.nombre) LIKE LOWER($"+strconv.Itoa(idx)+") OR LOWER(cb.codigo) LIKE LOWER($"+strconv.Itoa(idx)+"))")
		args = append(args, searchTerm)
		idx++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM combos cb %s", whereClause)
	var total int64
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, err
	}

	limit := filter.Limite
	if limit <= 0 {
		limit = 20
	}
	if limit > 200 {
		limit = 200
	}

	page := filter.Pagina
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * limit

	query := fmt.Sprintf(`
		SELECT cb.id, cb.nombre, cb.codigo, cb.categoria_id, c.nombre AS categoria_nombre,
		       cb.imagen, cb.estado, cb.costo, cb.precio_venta, cb.creado_en, cb.actualizado_en
		FROM combos cb
		LEFT JOIN categorias c ON cb.categoria_id = c.id
		%s
		ORDER BY cb.nombre ASC
		LIMIT $%d OFFSET $%d
	`, whereClause, len(args)+1, len(args)+2)

	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var combos []domain.Combo
	for rows.Next() {
		var cb domain.Combo
		var catID sql.NullInt64
		var catNom, img sql.NullString

		err := rows.Scan(
			&cb.ID, &cb.Nombre, &cb.Codigo, &catID, &catNom,
			&img, &cb.Estado, &cb.Costo, &cb.PrecioVenta, &cb.CreadoEn, &cb.ActualizadoEn,
		)
		if err != nil {
			return nil, err
		}
		if catID.Valid {
			cb.CategoriaID = &catID.Int64
		}
		if catNom.Valid {
			cb.CategoriaNombre = &catNom.String
		}
		if img.Valid {
			cb.Imagen = &img.String
		}
		combos = append(combos, cb)
	}

	totalPaginas := int(total) / limit
	if int(total)%limit != 0 {
		totalPaginas++
	}

	return &domain.PagedResult[domain.Combo]{
		Data:         combos,
		Total:        total,
		Pagina:       page,
		Limite:       limit,
		TotalPaginas: totalPaginas,
	}, nil
}

func (r *comboRepository) FindByID(ctx context.Context, id int64) (*domain.Combo, error) {
	query := `
		SELECT cb.id, cb.nombre, cb.codigo, cb.categoria_id, c.nombre AS categoria_nombre,
		       cb.imagen, cb.estado, cb.costo, cb.precio_venta, cb.creado_en, cb.actualizado_en
		FROM combos cb
		LEFT JOIN categorias c ON cb.categoria_id = c.id
		WHERE cb.id = $1
	`
	var cb domain.Combo
	var catID sql.NullInt64
	var catNom, img sql.NullString

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&cb.ID, &cb.Nombre, &cb.Codigo, &catID, &catNom,
		&img, &cb.Estado, &cb.Costo, &cb.PrecioVenta, &cb.CreadoEn, &cb.ActualizadoEn,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	if catID.Valid {
		cb.CategoriaID = &catID.Int64
	}
	if catNom.Valid {
		cb.CategoriaNombre = &catNom.String
	}
	if img.Valid {
		cb.Imagen = &img.String
	}

	components, err := r.GetComponents(ctx, id)
	if err != nil {
		return nil, err
	}
	cb.Componentes = components

	return &cb, nil
}

func (r *comboRepository) FindByCode(ctx context.Context, code string) (*domain.Combo, error) {
	query := `
		SELECT cb.id, cb.nombre, cb.codigo, cb.categoria_id, c.nombre AS categoria_nombre,
		       cb.imagen, cb.estado, cb.costo, cb.precio_venta, cb.creado_en, cb.actualizado_en
		FROM combos cb
		LEFT JOIN categorias c ON cb.categoria_id = c.id
		WHERE cb.codigo = $1
	`
	var cb domain.Combo
	var catID sql.NullInt64
	var catNom, img sql.NullString

	err := r.db.QueryRowContext(ctx, query, code).Scan(
		&cb.ID, &cb.Nombre, &cb.Codigo, &catID, &catNom,
		&img, &cb.Estado, &cb.Costo, &cb.PrecioVenta, &cb.CreadoEn, &cb.ActualizadoEn,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	if catID.Valid {
		cb.CategoriaID = &catID.Int64
	}
	if catNom.Valid {
		cb.CategoriaNombre = &catNom.String
	}
	if img.Valid {
		cb.Imagen = &img.String
	}
	return &cb, nil
}

func (r *comboRepository) Update(ctx context.Context, id int64, combo *domain.Combo) (*domain.Combo, error) {
	query := `
		UPDATE combos SET
			nombre = $1,
			codigo = $2,
			categoria_id = $3,
			imagen = $4,
			estado = $5,
			precio_venta = $6,
			actualizado_en = CURRENT_TIMESTAMP
		WHERE id = $7
		RETURNING costo, creado_en, actualizado_en
	`
	err := r.db.QueryRowContext(
		ctx, query,
		combo.Nombre, combo.Codigo, combo.CategoriaID, combo.Imagen, combo.Estado, combo.PrecioVenta, id,
	).Scan(&combo.Costo, &combo.CreadoEn, &combo.ActualizadoEn)

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
	combo.ID = id
	return combo, nil
}

func (r *comboRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM combos WHERE id = $1`
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

func (r *comboRepository) GetComponents(ctx context.Context, comboID int64) ([]domain.ComboProducto, error) {
	query := `
		SELECT cp.id, cp.combo_id, cp.producto_id, cp.cantidad,
		       p.id, p.nombre, p.codigo, p.descripcion, p.categoria_id, c.nombre AS categoria_nombre,
		       p.unidad, p.costo_unidad, p.venta_unidad, p.costo, p.precio_venta, p.stock,
		       p.proveedor_id, pr.nombre AS proveedor_nombre, TO_CHAR(p.vencimiento, 'YYYY-MM-DD') AS vencimiento,
		       p.imagen, p.estado, p.creado_en, p.actualizado_en
		FROM combo_productos cp
		INNER JOIN productos p ON cp.producto_id = p.id
		LEFT JOIN categorias c ON p.categoria_id = c.id
		LEFT JOIN proveedores pr ON p.proveedor_id = pr.id
		WHERE cp.combo_id = $1
		ORDER BY cp.id ASC
	`
	rows, err := r.db.QueryContext(ctx, query, comboID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var components []domain.ComboProducto
	for rows.Next() {
		var cp domain.ComboProducto
		var prod domain.Producto
		var desc, venc, img, catNom, provNom sql.NullString
		var catID, provID sql.NullInt64

		err := rows.Scan(
			&cp.ID, &cp.ComboID, &cp.ProductoID, &cp.Cantidad,
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

		cp.Producto = &prod
		unitCost := prod.CostPerBaseUnit()
		cp.Subtotal = unitCost * cp.Cantidad
		components = append(components, cp)
	}
	return components, nil
}

func (r *comboRepository) AddComponent(ctx context.Context, comboID int64, prodID int64, cantidad float64) error {
	query := `INSERT INTO combo_productos (combo_id, producto_id, cantidad) VALUES ($1, $2, $3)`
	_, err := r.db.ExecContext(ctx, query, comboID, prodID, cantidad)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			if pqErr.Code == "23505" {
				return domain.ErrDuplicateKey
			}
			if pqErr.Code == "23503" {
				return domain.ErrForeignKeyViolated
			}
			if pqErr.Code == "23514" {
				return domain.ErrInvalidInput
			}
		}
		return err
	}
	return nil
}

func (r *comboRepository) UpdateComponent(ctx context.Context, comboID int64, prodID int64, cantidad float64) error {
	query := `UPDATE combo_productos SET cantidad = $1 WHERE combo_id = $2 AND producto_id = $3`
	res, err := r.db.ExecContext(ctx, query, cantidad, comboID, prodID)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			if pqErr.Code == "23514" {
				return domain.ErrInvalidInput
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

func (r *comboRepository) RemoveComponent(ctx context.Context, comboID int64, prodID int64) error {
	query := `DELETE FROM combo_productos WHERE combo_id = $1 AND producto_id = $2`
	res, err := r.db.ExecContext(ctx, query, comboID, prodID)
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

func (r *comboRepository) UpdateStoredCost(ctx context.Context, comboID int64, cost float64) error {
	query := `UPDATE combos SET costo = $1, actualizado_en = CURRENT_TIMESTAMP WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, cost, comboID)
	return err
}

func (r *comboRepository) FindCombosByProductID(ctx context.Context, prodID int64) ([]int64, error) {
	query := `SELECT combo_id FROM combo_productos WHERE producto_id = $1`
	rows, err := r.db.QueryContext(ctx, query, prodID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comboIDs []int64
	for rows.Next() {
		var cid int64
		if err := rows.Scan(&cid); err != nil {
			return nil, err
		}
		comboIDs = append(comboIDs, cid)
	}
	return comboIDs, nil
}
