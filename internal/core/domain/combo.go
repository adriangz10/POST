package domain

import (
	"math"
	"strings"
	"time"
)

type ComboProducto struct {
	ID         int64     `json:"id"`
	ComboID    int64     `json:"combo_id"`
	ProductoID int64     `json:"producto_id"`
	Cantidad   float64   `json:"cantidad"`
	Producto   *Producto `json:"producto,omitempty"`
	Subtotal   float64   `json:"subtotal"`
}

type Combo struct {
	ID              int64           `json:"id"`
	Nombre          string          `json:"nombre"`
	Codigo          string          `json:"codigo"`
	CategoriaID     *int64          `json:"categoria_id,omitempty"`
	CategoriaNombre *string         `json:"categoria_nombre,omitempty"`
	Imagen          *string         `json:"imagen,omitempty"`
	Estado          string          `json:"estado"`
	Costo           float64         `json:"costo"`
	PrecioVenta     float64         `json:"precio_venta"`
	CreadoEn        time.Time       `json:"creado_en"`
	ActualizadoEn   time.Time       `json:"actualizado_en"`
	Componentes     []ComboProducto `json:"componentes,omitempty"`
}

type CreateComboDTO struct {
	Nombre      string                    `json:"nombre"`
	Codigo      string                    `json:"codigo"`
	CategoriaID *int64                    `json:"categoria_id,omitempty"`
	Imagen      *string                   `json:"imagen,omitempty"`
	Estado      string                    `json:"estado"`
	PrecioVenta float64                   `json:"precio_venta"`
	Componentes []CreateComboComponentDTO `json:"componentes,omitempty"`
}

type CreateComboComponentDTO struct {
	ProductoID int64   `json:"producto_id"`
	Cantidad   float64 `json:"cantidad"`
}

type UpdateComboDTO struct {
	Nombre      *string  `json:"nombre,omitempty"`
	Codigo      *string  `json:"codigo,omitempty"`
	CategoriaID *int64   `json:"categoria_id,omitempty"`
	Imagen      *string  `json:"imagen,omitempty"`
	Estado      *string  `json:"estado,omitempty"`
	PrecioVenta *float64 `json:"precio_venta,omitempty"`
}

type ComboComponentInputDTO struct {
	ProductoID int64   `json:"producto_id"`
	Cantidad   float64 `json:"cantidad"`
}

type ComboComponentUpdateDTO struct {
	Cantidad float64 `json:"cantidad"`
}

type ComboFilter struct {
	CategoriaID *int64  `json:"categoria_id,omitempty"`
	Estado      *string `json:"estado,omitempty"`
	Buscar      *string `json:"buscar,omitempty"`
	Pagina      int     `json:"pagina,omitempty"`
	Limite      int     `json:"limite,omitempty"`
}

func (dto *CreateComboDTO) Validate() error {
	dto.Nombre = strings.TrimSpace(dto.Nombre)
	if dto.Nombre == "" {
		return &ValidationError{Field: "nombre", Message: "el nombre del combo es obligatorio"}
	}
	if len(dto.Nombre) > 150 {
		return &ValidationError{Field: "nombre", Message: "el nombre no puede exceder 150 caracteres"}
	}

	dto.Codigo = strings.TrimSpace(dto.Codigo)
	if dto.Codigo == "" {
		return &ValidationError{Field: "codigo", Message: "el código del combo es obligatorio"}
	}
	if len(dto.Codigo) > 100 {
		return &ValidationError{Field: "codigo", Message: "el código no puede exceder 100 caracteres"}
	}

	if dto.Estado == "" {
		dto.Estado = "activo"
	} else if !ValidStates[dto.Estado] {
		return &ValidationError{Field: "estado", Message: "estado inválido, debe ser: activo o inactivo"}
	}

	if dto.PrecioVenta < 0 {
		return &ValidationError{Field: "precio_venta", Message: "el precio de venta no puede ser negativo"}
	}

	seenProducts := make(map[int64]bool)
	for i, comp := range dto.Componentes {
		if comp.ProductoID <= 0 {
			return &ValidationError{Field: "componentes", Message: "identificador de producto inválido en componente"}
		}
		if comp.Cantidad <= 0 {
			return &ValidationError{Field: "componentes", Message: "la cantidad debe ser mayor a cero"}
		}
		if seenProducts[comp.ProductoID] {
			return &ValidationError{Field: "componentes", Message: "no se puede repetir un producto dentro del mismo combo"}
		}
		seenProducts[comp.ProductoID] = true
		dto.Componentes[i] = comp
	}

	return nil
}

func (dto *UpdateComboDTO) Validate() error {
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

	if dto.Estado != nil && !ValidStates[*dto.Estado] {
		return &ValidationError{Field: "estado", Message: "estado inválido, debe ser: activo o inactivo"}
	}

	if dto.PrecioVenta != nil && *dto.PrecioVenta < 0 {
		return &ValidationError{Field: "precio_venta", Message: "el precio de venta no puede ser negativo"}
	}

	return nil
}

func (dto *ComboComponentInputDTO) Validate() error {
	if dto.ProductoID <= 0 {
		return &ValidationError{Field: "producto_id", Message: "debe especificar un producto válido"}
	}
	if dto.Cantidad <= 0 {
		return &ValidationError{Field: "cantidad", Message: "la cantidad debe ser mayor a cero"}
	}
	return nil
}

func (dto *ComboComponentUpdateDTO) Validate() error {
	if dto.Cantidad <= 0 {
		return &ValidationError{Field: "cantidad", Message: "la cantidad debe ser mayor a cero"}
	}
	return nil
}

// CalculateTotalCost calcula el costo total del combo a partir de sus componentes
func CalculateTotalCost(components []ComboProducto) float64 {
	var total float64
	for _, comp := range components {
		if comp.Producto != nil {
			unitCost := comp.Producto.CostPerBaseUnit()
			subtotal := math.Round(unitCost*comp.Cantidad*100) / 100
			total += subtotal
		}
	}
	return math.Round(total*100) / 100
}
