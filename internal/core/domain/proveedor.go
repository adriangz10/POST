package domain

import "strings"

type Proveedor struct {
	ID     int64  `json:"id"`
	Nombre string `json:"nombre"`
}

type CreateProveedorDTO struct {
	Nombre string `json:"nombre"`
}

type UpdateProveedorDTO struct {
	Nombre *string `json:"nombre"`
}

func (dto *CreateProveedorDTO) Validate() error {
	dto.Nombre = strings.TrimSpace(dto.Nombre)
	if dto.Nombre == "" {
		return &ValidationError{Field: "nombre", Message: "el nombre del proveedor es obligatorio"}
	}
	if len(dto.Nombre) > 150 {
		return &ValidationError{Field: "nombre", Message: "el nombre no puede exceder 150 caracteres"}
	}
	return nil
}

func (dto *UpdateProveedorDTO) Validate() error {
	if dto.Nombre != nil {
		*dto.Nombre = strings.TrimSpace(*dto.Nombre)
		if *dto.Nombre == "" {
			return &ValidationError{Field: "nombre", Message: "el nombre no puede ser vacío"}
		}
		if len(*dto.Nombre) > 150 {
			return &ValidationError{Field: "nombre", Message: "el nombre no puede exceder 150 caracteres"}
		}
	}
	return nil
}
