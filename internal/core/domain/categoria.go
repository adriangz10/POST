package domain

import (
	"regexp"
	"strings"
)

var colorRegex = regexp.MustCompile(`^#[0-9A-Fa-f]{6}$`)

type Categoria struct {
	ID     int64  `json:"id"`
	Nombre string `json:"nombre"`
	Color  string `json:"color"`
}

type CreateCategoriaDTO struct {
	Nombre string `json:"nombre"`
	Color  string `json:"color"`
}

type UpdateCategoriaDTO struct {
	Nombre *string `json:"nombre"`
	Color  *string `json:"color"`
}

func (dto *CreateCategoriaDTO) Validate() error {
	dto.Nombre = strings.TrimSpace(dto.Nombre)
	if dto.Nombre == "" {
		return &ValidationError{Field: "nombre", Message: "el nombre de la categoría es obligatorio"}
	}
	if len(dto.Nombre) > 100 {
		return &ValidationError{Field: "nombre", Message: "el nombre no puede exceder 100 caracteres"}
	}

	dto.Color = strings.TrimSpace(dto.Color)
	if dto.Color == "" {
		dto.Color = "#808080"
	} else if !colorRegex.MatchString(dto.Color) {
		return &ValidationError{Field: "color", Message: "el color debe tener formato hexadecimal válido (ej. #808080)"}
	}
	return nil
}

func (dto *UpdateCategoriaDTO) Validate() error {
	if dto.Nombre != nil {
		*dto.Nombre = strings.TrimSpace(*dto.Nombre)
		if *dto.Nombre == "" {
			return &ValidationError{Field: "nombre", Message: "el nombre no puede ser vacío"}
		}
		if len(*dto.Nombre) > 100 {
			return &ValidationError{Field: "nombre", Message: "el nombre no puede exceder 100 caracteres"}
		}
	}

	if dto.Color != nil {
		*dto.Color = strings.TrimSpace(*dto.Color)
		if !colorRegex.MatchString(*dto.Color) {
			return &ValidationError{Field: "color", Message: "el color debe tener formato hexadecimal válido (ej. #808080)"}
		}
	}
	return nil
}
