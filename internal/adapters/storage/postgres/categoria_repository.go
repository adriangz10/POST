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

type categoryRepository struct {
	db *sql.DB
}

func NewCategoryRepository(db *sql.DB) ports.CategoryRepository {
	return &categoryRepository{db: db}
}

func (r *categoryRepository) Create(ctx context.Context, cat *domain.Categoria) (*domain.Categoria, error) {
	query := `INSERT INTO categorias (nombre, color) VALUES ($1, $2) RETURNING id`
	err := r.db.QueryRowContext(ctx, query, cat.Nombre, cat.Color).Scan(&cat.ID)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			if pqErr.Code == "23505" {
				return nil, domain.ErrDuplicateKey
			}
			if pqErr.Code == "23514" {
				return nil, domain.ErrInvalidInput
			}
		}
		return nil, err
	}
	return cat, nil
}

func (r *categoryRepository) FindAll(ctx context.Context, search string) ([]domain.Categoria, error) {
	var query string
	var args []interface{}

	search = strings.TrimSpace(search)
	if search != "" {
		query = `SELECT id, nombre, color FROM categorias WHERE LOWER(nombre) LIKE LOWER($1) ORDER BY nombre ASC`
		args = append(args, "%"+search+"%")
	} else {
		query = `SELECT id, nombre, color FROM categorias ORDER BY nombre ASC`
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []domain.Categoria
	for rows.Next() {
		var c domain.Categoria
		if err := rows.Scan(&c.ID, &c.Nombre, &c.Color); err != nil {
			return nil, err
		}
		categories = append(categories, c)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return categories, nil
}

func (r *categoryRepository) FindByID(ctx context.Context, id int64) (*domain.Categoria, error) {
	query := `SELECT id, nombre, color FROM categorias WHERE id = $1`
	var c domain.Categoria
	err := r.db.QueryRowContext(ctx, query, id).Scan(&c.ID, &c.Nombre, &c.Color)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &c, nil
}

func (r *categoryRepository) FindByName(ctx context.Context, name string) (*domain.Categoria, error) {
	query := `SELECT id, nombre, color FROM categorias WHERE LOWER(nombre) = LOWER($1)`
	var c domain.Categoria
	err := r.db.QueryRowContext(ctx, query, name).Scan(&c.ID, &c.Nombre, &c.Color)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &c, nil
}

func (r *categoryRepository) Update(ctx context.Context, id int64, cat *domain.Categoria) (*domain.Categoria, error) {
	query := `UPDATE categorias SET nombre = $1, color = $2 WHERE id = $3 RETURNING id, nombre, color`
	var updated domain.Categoria
	err := r.db.QueryRowContext(ctx, query, cat.Nombre, cat.Color, id).Scan(&updated.ID, &updated.Nombre, &updated.Color)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			if pqErr.Code == "23505" {
				return nil, domain.ErrDuplicateKey
			}
			if pqErr.Code == "23514" {
				return nil, domain.ErrInvalidInput
			}
		}
		return nil, err
	}
	return &updated, nil
}

func (r *categoryRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM categorias WHERE id = $1`
	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			if pqErr.Code == "23503" {
				return domain.ErrForeignKeyViolated
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

func (r *categoryRepository) CountProducts(ctx context.Context, id int64) (int64, error) {
	query := `SELECT COUNT(*) FROM productos WHERE categoria_id = $1`
	var count int64
	err := r.db.QueryRowContext(ctx, query, id).Scan(&count)
	return count, err
}

func (r *categoryRepository) CountCombos(ctx context.Context, id int64) (int64, error) {
	query := `SELECT COUNT(*) FROM combos WHERE categoria_id = $1`
	var count int64
	err := r.db.QueryRowContext(ctx, query, id).Scan(&count)
	return count, err
}
