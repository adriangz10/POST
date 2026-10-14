package domain

import (
	"math"
	"strings"
)

const MaxAllowedPercentage = 500.0 // Límite máximo configurable para evitar cambios accidentales

type PreciosOperacion string

const (
	OperacionAumento     PreciosOperacion = "aumento"
	OperacionDisminucion PreciosOperacion = "disminucion"
)

type PreciosAlcance string

const (
	AlcanceTodos     PreciosAlcance = "todos"
	AlcanceCategoria PreciosAlcance = "categoria"
	AlcanceProveedor PreciosAlcance = "proveedor"
)

type ActualizacionPreciosDTO struct {
	Operacion   PreciosOperacion `json:"operacion"`    // "aumento" o "disminucion"
	Porcentaje  float64          `json:"porcentaje"`   // Ej: 3.5 para 3.5%
	Alcance     PreciosAlcance   `json:"alcance"`      // "todos", "categoria", "proveedor"
	CategoriaID *int64           `json:"categoria_id"` // Requerido si alcance es "categoria"
	ProveedorID *int64           `json:"proveedor_id"` // Requerido si alcance es "proveedor"
}

type PreciosProductoPreview struct {
	ID           int64   `json:"id"`
	Codigo       string  `json:"codigo"`
	Nombre       string  `json:"nombre"`
	PrecioActual float64 `json:"precio_actual"`
	NuevoPrecio  float64 `json:"nuevo_precio"`
	Diferencia   float64 `json:"diferencia"`
}

type PreciosPreviewResult struct {
	Operacion      PreciosOperacion         `json:"operacion"`
	Porcentaje     float64                  `json:"porcentaje"`
	Alcance        PreciosAlcance           `json:"alcance"`
	TotalAfectados int                      `json:"total_afectados"`
	Productos      []PreciosProductoPreview `json:"productos"`
	PoliticaCombos string                   `json:"politica_combos"`
}

type PreciosUpdateResult struct {
	Operacion      PreciosOperacion `json:"operacion"`
	Porcentaje     float64          `json:"porcentaje"`
	Alcance        PreciosAlcance   `json:"alcance"`
	Actualizados   int64            `json:"actualizados"`
	Mensaje        string           `json:"mensaje"`
	PoliticaCombos string           `json:"politica_combos"`
}

func (dto *ActualizacionPreciosDTO) Validate() error {
	dto.Operacion = PreciosOperacion(strings.ToLower(strings.TrimSpace(string(dto.Operacion))))
	if dto.Operacion == "descuento" {
		dto.Operacion = OperacionDisminucion
	}
	if dto.Operacion != OperacionAumento && dto.Operacion != OperacionDisminucion {
		return &ValidationError{Field: "operacion", Message: "la operación debe ser 'aumento' o 'disminucion'"}
	}

	if dto.Porcentaje <= 0 {
		return &ValidationError{Field: "porcentaje", Message: "el porcentaje debe ser un número mayor a cero"}
	}
	if dto.Porcentaje > MaxAllowedPercentage {
		return &ValidationError{Field: "porcentaje", Message: "el porcentaje no puede exceder el límite máximo de 500%"}
	}

	dto.Alcance = PreciosAlcance(strings.ToLower(strings.TrimSpace(string(dto.Alcance))))
	if dto.Alcance != AlcanceTodos && dto.Alcance != AlcanceCategoria && dto.Alcance != AlcanceProveedor {
		return &ValidationError{Field: "alcance", Message: "el alcance debe ser 'todos', 'categoria' o 'proveedor'"}
	}

	if dto.Alcance == AlcanceCategoria {
		if dto.CategoriaID == nil || *dto.CategoriaID <= 0 {
			return &ValidationError{Field: "categoria_id", Message: "debe especificar un id de categoría válido para el alcance 'categoria'"}
		}
	}

	if dto.Alcance == AlcanceProveedor {
		if dto.ProveedorID == nil || *dto.ProveedorID <= 0 {
			return &ValidationError{Field: "proveedor_id", Message: "debe especificar un id de proveedor válido para el alcance 'proveedor'"}
		}
	}

	return nil
}

// CalculateNewPrice calcula el nuevo precio con redondeo a 2 decimales
func CalculateNewPrice(precioActual float64, operacion PreciosOperacion, porcentaje float64) float64 {
	var nuevo float64
	if operacion == OperacionAumento {
		nuevo = precioActual * (1.0 + (porcentaje / 100.0))
	} else {
		nuevo = precioActual * (1.0 - (porcentaje / 100.0))
	}
	if nuevo < 0 {
		nuevo = 0
	}
	return math.Round(nuevo*100) / 100
}
