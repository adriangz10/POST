package domain_test

import (
	"testing"

	"post/internal/core/domain"
)

func TestCategoriaValidation(t *testing.T) {
	// Valida nombre obligatorio
	c1 := domain.CreateCategoriaDTO{Nombre: "", Color: "#FFFFFF"}
	if err := c1.Validate(); err == nil {
		t.Errorf("Se esperaba error por nombre vacío")
	}

	// Valida color formato hexadecimal inválido
	c2 := domain.CreateCategoriaDTO{Nombre: "Bebidas", Color: "rojo"}
	if err := c2.Validate(); err == nil {
		t.Errorf("Se esperaba error por formato de color inválido")
	}

	// Color válido por defecto
	c3 := domain.CreateCategoriaDTO{Nombre: "Bebidas", Color: ""}
	if err := c3.Validate(); err != nil || c3.Color != "#808080" {
		t.Errorf("Se esperaba color por defecto #808080, obtenido: %s", c3.Color)
	}

	// Color válido personalizado
	c4 := domain.CreateCategoriaDTO{Nombre: "Snacks", Color: "#FF5733"}
	if err := c4.Validate(); err != nil {
		t.Errorf("Error inesperado en categoría válida: %v", err)
	}
}

func TestProductoValidationAndConversions(t *testing.T) {
	p := domain.CreateProductoDTO{
		Nombre:      "Café Molido",
		Codigo:      "CAF-001",
		Unidad:      "gramo",
		CostoUnidad: "kilogramo",
		VentaUnidad: "gramo",
		Costo:       5000.0, // 5000 por kg
		PrecioVenta: 10.0,
		Stock:       2500.0, // 2500 gramos
		Estado:      "activo",
	}

	if err := p.Validate(); err != nil {
		t.Fatalf("Error de validación inesperado: %v", err)
	}

	prod := domain.Producto{
		Unidad:      p.Unidad,
		CostoUnidad: p.CostoUnidad,
		Costo:       p.Costo,
	}

	// 5000 por kg significa 5.0 por gramo
	costPerGram := prod.CostPerBaseUnit()
	if costPerGram != 5.0 {
		t.Errorf("Costo por gramo esperado 5.0, obtenido %.2f", costPerGram)
	}

	// Valida precios y stocks negativos
	pNeg := domain.CreateProductoDTO{
		Nombre:      "Inválido",
		Codigo:      "INV-1",
		Costo:       -10,
		PrecioVenta: 20,
	}
	if err := pNeg.Validate(); err == nil {
		t.Errorf("Se esperaba error por costo negativo")
	}
}

func TestComboCostCalculation(t *testing.T) {
	// Producto 1: Costo $1000 por kg, base gramo -> $1 por gramo
	p1 := domain.Producto{
		ID:          1,
		Nombre:      "Harina",
		Unidad:      "gramo",
		CostoUnidad: "kilogramo",
		Costo:       1000.0,
	}

	// Producto 2: Costo $50 por unidad, base unitario -> $50 por unidad
	p2 := domain.Producto{
		ID:          2,
		Nombre:      "Huevo",
		Unidad:      "unitario",
		CostoUnidad: "unitario",
		Costo:       50.0,
	}

	components := []domain.ComboProducto{
		{
			ProductoID: 1,
			Cantidad:   500, // 500g * $1 = $500
			Producto:   &p1,
		},
		{
			ProductoID: 2,
			Cantidad:   3, // 3 * $50 = $150
			Producto:   &p2,
		},
	}

	total := domain.CalculateTotalCost(components)
	expected := 650.0 // 500 + 150
	if total != expected {
		t.Errorf("Costo total esperado %.2f, obtenido %.2f", expected, total)
	}

	// Validación de componentes duplicados en DTO
	comboDTO := domain.CreateComboDTO{
		Nombre: "Combo Desayuno",
		Codigo: "CMB-DES",
		Componentes: []domain.CreateComboComponentDTO{
			{ProductoID: 1, Cantidad: 100},
			{ProductoID: 1, Cantidad: 200}, // Repetido
		},
	}
	if err := comboDTO.Validate(); err == nil {
		t.Errorf("Se esperaba error por productos duplicados dentro del combo")
	}
}

func TestPreciosCalculationAndRounding(t *testing.T) {
	// Aumento del 3% sobre $100.33
	// 100.33 * 1.03 = 103.3400 -> 103.34
	nuevo := domain.CalculateNewPrice(100.33, domain.OperacionAumento, 3.0)
	if nuevo != 103.34 {
		t.Errorf("Aumento esperado 103.34, obtenido %.2f", nuevo)
	}

	// Disminución del 1% sobre $150.25
	// 150.25 * 0.99 = 148.7475 -> 148.75
	nuevoDisminucion := domain.CalculateNewPrice(150.25, domain.OperacionDisminucion, 1.0)
	if nuevoDisminucion != 148.75 {
		t.Errorf("Disminución esperada 148.75, obtenida %.2f", nuevoDisminucion)
	}

	// Validaciones DTO de Precios
	dtoInvalido := domain.ActualizacionPreciosDTO{
		Operacion:  "multiplicacion", // Inválido
		Porcentaje: -5,
		Alcance:    "todos",
	}
	if err := dtoInvalido.Validate(); err == nil {
		t.Errorf("Se esperaba error por operación y porcentaje inválidos")
	}

	// Límite máximo superado
	dtoMax := domain.ActualizacionPreciosDTO{
		Operacion:  domain.OperacionAumento,
		Porcentaje: 600, // Máximo es 500
		Alcance:    domain.AlcanceTodos,
	}
	if err := dtoMax.Validate(); err == nil {
		t.Errorf("Se esperaba error por superar el límite máximo de porcentaje")
	}
}
