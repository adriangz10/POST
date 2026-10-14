package domain

import "errors"

var (
	ErrNotFound           = errors.New("recurso no encontrado")
	ErrDuplicateKey       = errors.New("el recurso ya existe o viola una restricción de unicidad")
	ErrInvalidInput       = errors.New("datos de entrada inválidos")
	ErrForeignKeyViolated = errors.New("violación de integridad referencial")
	ErrConflict           = errors.New("conflicto con el estado actual del recurso")
)

// ValidationError representa un error de validación específico de un campo
type ValidationError struct {
	Field   string `json:"campo"`
	Message string `json:"mensaje"`
}

func (v ValidationError) Error() string {
	return v.Field + ": " + v.Message
}
