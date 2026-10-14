package domain

import (
	"fmt"
	"strings"
	"time"
)

var ValidUnits = map[string]bool{
	"unitario":  true,
	"kilogramo": true,
	"gramo":     true,
	"litro":     true,
	"mililitro": true,
}

var ValidStates = map[string]bool{
	"activo":   true,
	"inactivo": true,
}

type Producto struct {
	ID              int64     `json:"id"`
	Nombre          string    `json:"nombre"`
	Codigo          string    `json:"codigo"`
	Descripcion     *string   `json:"descripcion,omitempty"`
	CategoriaID     *int64    `json:"categoria_id,omitempty"`
	CategoriaNombre *string   `json:"categoria_nombre,omitempty"`
	CategoriaColor  *string   `json:"categoria_color,omitempty"`
	Unidad          string    `json:"unidad"`
	CostoUnidad     string    `json:"costo_unidad"`
	VentaUnidad     string    `json:"venta_unidad"`
	Costo           float64   `json:"costo"`
	PrecioVenta     float64   `json:"precio_venta"`
	Stock           float64   `json:"stock"`
	ProveedorID     *int64    `json:"proveedor_id,omitempty"`
	ProveedorNombre *string   `json:"proveedor_nombre,omitempty"`
	Vencimiento     *string   `json:"vencimiento,omitempty"` // YYYY-MM-DD
	Imagen          *string   `json:"imagen,omitempty"`
	Estado          string    `json:"estado"`
	CreadoEn        time.Time `json:"creado_en"`
	ActualizadoEn   time.Time `json:"actualizado_en"`
}

type CreateProductoDTO struct {
	Nombre      string  `json:"nombre"`
	Codigo      string  `json:"codigo"`
	Descripcion *string `json:"descripcion,omitempty"`
	CategoriaID *int64  `json:"categoria_id,omitempty"`
	Unidad      string  `json:"unidad"`
	CostoUnidad string  `json:"costo_unidad"`
	VentaUnidad string  `json:"venta_unidad"`
	Costo       float64 `json:"costo"`
	PrecioVenta float64 `json:"precio_venta"`
	Stock       float64 `json:"stock"`
	ProveedorID *int64  `json:"proveedor_id,omitempty"`
	Vencimiento *string `json:"vencimiento,omitempty"`
	Imagen      *string `json:"imagen,omitempty"`
	Estado      string  `json:"estado"`
}

type UpdateProductoDTO struct {
	Nombre      *string  `json:"nombre,omitempty"`
	Codigo      *string  `json:"codigo,omitempty"`
	Descripcion *string  `json:"descripcion,omitempty"`
	CategoriaID *int64   `json:"categoria_id,omitempty"`
	Unidad      *string  `json:"unidad,omitempty"`
	CostoUnidad *string  `json:"costo_unidad,omitempty"`
	VentaUnidad *string  `json:"venta_unidad,omitempty"`
	Costo       *float64 `json:"costo,omitempty"`
	PrecioVenta *float64 `json:"precio_venta,omitempty"`
	Stock       *float64 `json:"stock,omitempty"`
	ProveedorID *int64   `json:"proveedor_id,omitempty"`
	Vencimiento *string  `json:"vencimiento,omitempty"`
	Imagen      *string  `json:"imagen,omitempty"`
	Estado      *string  `json:"estado,omitempty"`
}

type ProductoFilter struct {
	CategoriaID      *int64   `json:"categoria_id,omitempty"`
	ProveedorID      *int64   `json:"proveedor_id,omitempty"`
	Estado           *string  `json:"estado,omitempty"`
	Buscar           *string  `json:"buscar,omitempty"`
	VencimientoDesde *string  `json:"vencimiento_desde,omitempty"`
	VencimientoHasta *string  `json:"vencimiento_hasta,omitempty"`
	StockMin         *float64 `json:"stock_min,omitempty"`
	StockMax         *float64 `json:"stock_max,omitempty"`
	PrecioMin        *float64 `json:"precio_min,omitempty"`
	PrecioMax        *float64 `json:"precio_max,omitempty"`
	StockBajo        *bool    `json:"stock_bajo,omitempty"`
	UmbralStockBajo  *float64 `json:"umbral_stock_bajo,omitempty"`
	OrdenPor         string   `json:"orden_por,omitempty"`
	Direccion        string   `json:"direccion,omitempty"`
	Pagina           int      `json:"pagina,omitempty"`
	Limite           int      `json:"limite,omitempty"`
}

type PagedResult[T any] struct {
	Data         []T   `json:"datos"`
	Total        int64 `json:"total"`
	Pagina       int   `json:"pagina"`
	Limite       int   `json:"limite"`
	TotalPaginas int   `json:"total_paginas"`
}

func (dto *CreateProductoDTO) Validate() error {
	dto.Nombre = strings.TrimSpace(dto.Nombre)
	if dto.Nombre == "" {
		return &ValidationError{Field: "nombre", Message: "el nombre del producto es obligatorio"}
	}
	if len(dto.Nombre) > 150 {
		return &ValidationError{Field: "nombre", Message: "el nombre no puede exceder 150 caracteres"}
	}

	dto.Codigo = strings.TrimSpace(dto.Codigo)
	if dto.Codigo == "" {
		return &ValidationError{Field: "codigo", Message: "el código del producto es obligatorio"}
	}
	if len(dto.Codigo) > 100 {
		return &ValidationError{Field: "codigo", Message: "el código no puede exceder 100 caracteres"}
	}

	if dto.Unidad == "" {
		dto.Unidad = "unitario"
	} else if !ValidUnits[dto.Unidad] {
		return &ValidationError{Field: "unidad", Message: "unidad inválida, debe ser: unitario, kilogramo, gramo, litro o mililitro"}
	}

	if dto.CostoUnidad == "" {
		dto.CostoUnidad = dto.Unidad
	} else if !ValidUnits[dto.CostoUnidad] {
		return &ValidationError{Field: "costo_unidad", Message: "unidad de costo inválida"}
	}

	if dto.VentaUnidad == "" {
		dto.VentaUnidad = dto.Unidad
	} else if !ValidUnits[dto.VentaUnidad] {
		return &ValidationError{Field: "venta_unidad", Message: "unidad de venta inválida"}
	}

	if dto.Costo < 0 {
		return &ValidationError{Field: "costo", Message: "el costo no puede ser negativo"}
	}
	if dto.PrecioVenta < 0 {
		return &ValidationError{Field: "precio_venta", Message: "el precio de venta no puede ser negativo"}
	}
	if dto.Stock < 0 {
		return &ValidationError{Field: "stock", Message: "el stock no puede ser negativo"}
	}

	if dto.Estado == "" {
		dto.Estado = "activo"
	} else if !ValidStates[dto.Estado] {
		return &ValidationError{Field: "estado", Message: "estado inválido, debe ser: activo o inactivo"}
	}

	if dto.Vencimiento != nil && *dto.Vencimiento != "" {
		*dto.Vencimiento = strings.TrimSpace(*dto.Vencimiento)
		if _, err := time.Parse("2006-01-02", *dto.Vencimiento); err != nil {
			return &ValidationError{Field: "vencimiento", Message: "formato de fecha inválido, debe ser AAAA-MM-DD"}
		}
	}

	return nil
}

func (dto *UpdateProductoDTO) Validate() error {
	if dto.Nombre != nil {
		*dto.Nombre = strings.TrimSpace(*dto.Nombre)
		if *dto.Nombre == "" {
			return &ValidationError{Field: "nombre", Message: "el nombre no puede estar vacío"}
		}
		if len(*dto.Nombre) > 150 {
			return &ValidationError{Field: "nombre", Message: "el nombre no puede exceder 150 caracteres"}
		}
	}

	if dto.Codigo != nil {
		*dto.Codigo = strings.TrimSpace(*dto.Codigo)
		if *dto.Codigo == "" {
			return &ValidationError{Field: "codigo", Message: "el código no puede estar vacío"}
		}
		if len(*dto.Codigo) > 100 {
			return &ValidationError{Field: "codigo", Message: "el código no puede exceder 100 caracteres"}
		}
	}

	if dto.Unidad != nil && !ValidUnits[*dto.Unidad] {
		return &ValidationError{Field: "unidad", Message: "unidad inválida"}
	}
	if dto.CostoUnidad != nil && !ValidUnits[*dto.CostoUnidad] {
		return &ValidationError{Field: "costo_unidad", Message: "unidad de costo inválida"}
	}
	if dto.VentaUnidad != nil && !ValidUnits[*dto.VentaUnidad] {
		return &ValidationError{Field: "venta_unidad", Message: "unidad de venta inválida"}
	}

	if dto.Costo != nil && *dto.Costo < 0 {
		return &ValidationError{Field: "costo", Message: "el costo no puede ser negativo"}
	}
	if dto.PrecioVenta != nil && *dto.PrecioVenta < 0 {
		return &ValidationError{Field: "precio_venta", Message: "el precio de venta no puede ser negativo"}
	}
	if dto.Stock != nil && *dto.Stock < 0 {
		return &ValidationError{Field: "stock", Message: "el stock no puede ser negativo"}
	}

	if dto.Estado != nil && !ValidStates[*dto.Estado] {
		return &ValidationError{Field: "estado", Message: "estado inválido, debe ser: activo o inactivo"}
	}

	if dto.Vencimiento != nil && *dto.Vencimiento != "" {
		*dto.Vencimiento = strings.TrimSpace(*dto.Vencimiento)
		if _, err := time.Parse("2006-01-02", *dto.Vencimiento); err != nil {
			return &ValidationError{Field: "vencimiento", Message: "formato de fecha inválido, debe ser AAAA-MM-DD"}
		}
	}

	return nil
}

// ConvertQuantity convierte una cantidad entre unidades compatibles
func ConvertQuantity(amount float64, fromUnit, toUnit string) (float64, error) {
	if fromUnit == toUnit {
		return amount, nil
	}
	// Peso
	if fromUnit == "kilogramo" && toUnit == "gramo" {
		return amount * 1000.0, nil
	}
	if fromUnit == "gramo" && toUnit == "kilogramo" {
		return amount / 1000.0, nil
	}
	// Volumen
	if fromUnit == "litro" && toUnit == "mililitro" {
		return amount * 1000.0, nil
	}
	if fromUnit == "mililitro" && toUnit == "litro" {
		return amount / 1000.0, nil
	}
	return 0, fmt.Errorf("no se puede convertir de %s a %s: unidades incompatibles", fromUnit, toUnit)
}

// CostPerBaseUnit calcula el costo del producto por unidad base de referencia
func (p *Producto) CostPerBaseUnit() float64 {
	if p.CostoUnidad == p.Unidad {
		return p.Costo
	}
	// Si costo_unidad es kilogramo y base es gramo: 1 gramo cuesta costo / 1000
	if p.CostoUnidad == "kilogramo" && p.Unidad == "gramo" {
		return p.Costo / 1000.0
	}
	// Si costo_unidad es gramo y base es kilogramo: 1 kg cuesta costo * 1000
	if p.CostoUnidad == "gramo" && p.Unidad == "kilogramo" {
		return p.Costo * 1000.0
	}
	// Si costo_unidad es litro y base es mililitro: 1 ml cuesta costo / 1000
	if p.CostoUnidad == "litro" && p.Unidad == "mililitro" {
		return p.Costo / 1000.0
	}
	// Si costo_unidad es mililitro y base es litro: 1 litro cuesta costo * 1000
	if p.CostoUnidad == "mililitro" && p.Unidad == "litro" {
		return p.Costo * 1000.0
	}
	return p.Costo
}
