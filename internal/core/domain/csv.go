package domain

type CSVImportMode string

const (
	CSVModeCrear      CSVImportMode = "crear"
	CSVModeActualizar CSVImportMode = "actualizar"
)

type CSVFilaError struct {
	Fila    int    `json:"fila"`
	Codigo  string `json:"codigo,omitempty"`
	Campo   string `json:"campo"`
	Mensaje string `json:"mensaje"`
}

type CSVImportResult struct {
	Previsualizacion bool           `json:"previsualizacion"`
	Modo             CSVImportMode  `json:"modo"`
	TotalFilas       int            `json:"total_filas"`
	Creados          int            `json:"creados"`
	Actualizados     int            `json:"actualizados"`
	Omitidos         int            `json:"omitidos"`
	Rechazados       int            `json:"rechazados"`
	Errores          []CSVFilaError `json:"errores"`
}

// Columnas estándar esperadas para importación CSV
var CSVExpectedColumns = []string{
	"codigo",
	"nombre",
	"descripcion",
	"categoria_id",
	"unidad",
	"costo_unidad",
	"venta_unidad",
	"costo",
	"precio_venta",
	"stock",
	"proveedor_id",
	"vencimiento",
	"imagen",
	"estado",
}
