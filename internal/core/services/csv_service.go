package services

import (
	"bufio"
	"bytes"
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"post/internal/core/domain"
	"post/internal/core/ports"
)

type csvService struct {
	productRepo  ports.ProductRepository
	categoryRepo ports.CategoryRepository
	providerRepo ports.ProviderRepository
}

func NewCSVService(
	productRepo ports.ProductRepository,
	categoryRepo ports.CategoryRepository,
	providerRepo ports.ProviderRepository,
) ports.CSVService {
	return &csvService{
		productRepo:  productRepo,
		categoryRepo: categoryRepo,
		providerRepo: providerRepo,
	}
}

// detectDelimiter determina si el CSV usa coma o punto y coma
func detectDelimiter(firstLine string) rune {
	semicolons := strings.Count(firstLine, ";")
	commas := strings.Count(firstLine, ",")
	if semicolons > commas {
		return ';'
	}
	return ','
}

func (s *csvService) Import(ctx context.Context, r io.Reader, modo domain.CSVImportMode, previsualizar bool) (*domain.CSVImportResult, error) {
	if modo == "" {
		modo = domain.CSVModeCrear
	}
	if modo != domain.CSVModeCrear && modo != domain.CSVModeActualizar {
		return nil, &domain.ValidationError{Field: "modo", Message: "modo de importación debe ser 'crear' o 'actualizar'"}
	}

	// Leer todo el contenido para soportar UTF-8 BOM y detección de delimitador
	buf := new(bytes.Buffer)
	if _, err := io.Copy(buf, r); err != nil {
		return nil, fmt.Errorf("error al leer archivo CSV: %w", err)
	}

	contentBytes := buf.Bytes()
	// Remover BOM UTF-8 si existe
	if len(contentBytes) >= 3 && contentBytes[0] == 0xEF && contentBytes[1] == 0xBB && contentBytes[2] == 0xBF {
		contentBytes = contentBytes[3:]
	}

	firstLineReader := bufio.NewReader(bytes.NewReader(contentBytes))
	firstLine, err := firstLineReader.ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return nil, errors.New("archivo CSV vacío o ilegible")
	}

	delimiter := detectDelimiter(firstLine)

	csvReader := csv.NewReader(bytes.NewReader(contentBytes))
	csvReader.Comma = delimiter
	csvReader.TrimLeadingSpace = true
	csvReader.LazyQuotes = true

	records, err := csvReader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("formato CSV malformado: %w", err)
	}

	if len(records) < 1 {
		return nil, errors.New("el archivo CSV no contiene registros")
	}

	// Validar encabezados
	headerRow := records[0]
	headerMap := make(map[string]int)
	for i, h := range headerRow {
		cleaned := strings.ToLower(strings.TrimSpace(h))
		headerMap[cleaned] = i
	}

	// Comprobar columnas obligatorias mínimas
	if _, hasCode := headerMap["codigo"]; !hasCode {
		return nil, errors.New("el archivo CSV debe incluir al menos la columna 'codigo'")
	}
	if _, hasName := headerMap["nombre"]; !hasName {
		return nil, errors.New("el archivo CSV debe incluir al menos la columna 'nombre'")
	}

	// Cargar categorías y proveedores existentes para resolución rápida
	categories, _ := s.categoryRepo.FindAll(ctx, "")
	catIDMap := make(map[int64]bool)
	catNameMap := make(map[string]int64)
	for _, c := range categories {
		catIDMap[c.ID] = true
		catNameMap[strings.ToLower(c.Nombre)] = c.ID
	}

	providers, _ := s.providerRepo.FindAll(ctx, "")
	provIDMap := make(map[int64]bool)
	provNameMap := make(map[string]int64)
	for _, p := range providers {
		provIDMap[p.ID] = true
		provNameMap[strings.ToLower(p.Nombre)] = p.ID
	}

	result := &domain.CSVImportResult{
		Previsualizacion: previsualizar,
		Modo:             modo,
		TotalFilas:       len(records) - 1,
		Errores:          make([]domain.CSVFilaError, 0),
	}

	seenCodesInFile := make(map[string]int) // código -> número de fila
	type validProductItem struct {
		product    domain.Producto
		isUpdate   bool
		existingID int64
	}
	var validProducts []validProductItem

	for rowIdx := 1; rowIdx < len(records); rowIdx++ {
		rowNum := rowIdx + 1 // 1-based con fila 1 como cabecera
		row := records[rowIdx]

		getCol := func(colName string) string {
			if idx, ok := headerMap[colName]; ok && idx < len(row) {
				return strings.TrimSpace(row[idx])
			}
			return ""
		}

		codigo := getCol("codigo")
		nombre := getCol("nombre")

		if codigo == "" {
			result.Errores = append(result.Errores, domain.CSVFilaError{
				Fila:    rowNum,
				Campo:   "codigo",
				Mensaje: "el código del producto es obligatorio",
			})
			continue
		}

		if nombre == "" {
			result.Errores = append(result.Errores, domain.CSVFilaError{
				Fila:    rowNum,
				Codigo:  codigo,
				Campo:   "nombre",
				Mensaje: "el nombre del producto es obligatorio",
			})
			continue
		}

		// Chequear duplicado en el mismo archivo
		if prevRow, dup := seenCodesInFile[strings.ToLower(codigo)]; dup {
			result.Errores = append(result.Errores, domain.CSVFilaError{
				Fila:    rowNum,
				Codigo:  codigo,
				Campo:   "codigo",
				Mensaje: fmt.Sprintf("código duplicado dentro del archivo CSV (ya visto en la fila %d)", prevRow),
			})
			continue
		}
		seenCodesInFile[strings.ToLower(codigo)] = rowNum

		// Validar unidades
		unidad := getCol("unidad")
		if unidad == "" {
			unidad = "unitario"
		}
		if !domain.ValidUnits[unidad] {
			result.Errores = append(result.Errores, domain.CSVFilaError{
				Fila:    rowNum,
				Codigo:  codigo,
				Campo:   "unidad",
				Mensaje: "unidad inválida, debe ser: unitario, kilogramo, gramo, litro o mililitro",
			})
			continue
		}

		costoUnidad := getCol("costo_unidad")
		if costoUnidad == "" {
			costoUnidad = unidad
		}
		if !domain.ValidUnits[costoUnidad] {
			result.Errores = append(result.Errores, domain.CSVFilaError{
				Fila:    rowNum,
				Codigo:  codigo,
				Campo:   "costo_unidad",
				Mensaje: "unidad de costo inválida",
			})
			continue
		}

		ventaUnidad := getCol("venta_unidad")
		if ventaUnidad == "" {
			ventaUnidad = unidad
		}
		if !domain.ValidUnits[ventaUnidad] {
			result.Errores = append(result.Errores, domain.CSVFilaError{
				Fila:    rowNum,
				Codigo:  codigo,
				Campo:   "venta_unidad",
				Mensaje: "unidad de venta inválida",
			})
			continue
		}

		// Validar costo
		costoStr := getCol("costo")
		var costo float64
		if costoStr != "" {
			costoVal, err := strconv.ParseFloat(strings.ReplaceAll(costoStr, ",", "."), 64)
			if err != nil || costoVal < 0 {
				result.Errores = append(result.Errores, domain.CSVFilaError{
					Fila:    rowNum,
					Codigo:  codigo,
					Campo:   "costo",
					Mensaje: "el costo debe ser un número no negativo",
				})
				continue
			}
			costo = costoVal
		}

		// Validar precio_venta
		precioStr := getCol("precio_venta")
		var precioVenta float64
		if precioStr != "" {
			precioVal, err := strconv.ParseFloat(strings.ReplaceAll(precioStr, ",", "."), 64)
			if err != nil || precioVal < 0 {
				result.Errores = append(result.Errores, domain.CSVFilaError{
					Fila:    rowNum,
					Codigo:  codigo,
					Campo:   "precio_venta",
					Mensaje: "el precio de venta debe ser un número no negativo",
				})
				continue
			}
			precioVenta = precioVal
		}

		// Validar stock
		stockStr := getCol("stock")
		var stock float64
		if stockStr != "" {
			stockVal, err := strconv.ParseFloat(strings.ReplaceAll(stockStr, ",", "."), 64)
			if err != nil || stockVal < 0 {
				result.Errores = append(result.Errores, domain.CSVFilaError{
					Fila:    rowNum,
					Codigo:  codigo,
					Campo:   "stock",
					Mensaje: "el stock debe ser un número no negativo",
				})
				continue
			}
			stock = stockVal
		}

		// Validar estado
		estado := getCol("estado")
		if estado == "" {
			estado = "activo"
		} else {
			estado = strings.ToLower(estado)
		}
		if !domain.ValidStates[estado] {
			result.Errores = append(result.Errores, domain.CSVFilaError{
				Fila:    rowNum,
				Codigo:  codigo,
				Campo:   "estado",
				Mensaje: "estado inválido, debe ser 'activo' o 'inactivo'",
			})
			continue
		}

		// Resolver categoría
		var categoriaID *int64
		catIDStr := getCol("categoria_id")
		catNombre := getCol("categoria")
		if catIDStr != "" {
			cid, err := strconv.ParseInt(catIDStr, 10, 64)
			if err != nil || !catIDMap[cid] {
				result.Errores = append(result.Errores, domain.CSVFilaError{
					Fila:    rowNum,
					Codigo:  codigo,
					Campo:   "categoria_id",
					Mensaje: fmt.Sprintf("la categoría con ID %s no existe", catIDStr),
				})
				continue
			}
			categoriaID = &cid
		} else if catNombre != "" {
			if cid, exists := catNameMap[strings.ToLower(catNombre)]; exists {
				categoriaID = &cid
			} else {
				result.Errores = append(result.Errores, domain.CSVFilaError{
					Fila:    rowNum,
					Codigo:  codigo,
					Campo:   "categoria",
					Mensaje: fmt.Sprintf("la categoría '%s' no existe", catNombre),
				})
				continue
			}
		}

		// Resolver proveedor
		var proveedorID *int64
		provIDStr := getCol("proveedor_id")
		provNombre := getCol("proveedor")
		if provIDStr != "" {
			pid, err := strconv.ParseInt(provIDStr, 10, 64)
			if err != nil || !provIDMap[pid] {
				result.Errores = append(result.Errores, domain.CSVFilaError{
					Fila:    rowNum,
					Codigo:  codigo,
					Campo:   "proveedor_id",
					Mensaje: fmt.Sprintf("el proveedor con ID %s no existe", provIDStr),
				})
				continue
			}
			proveedorID = &pid
		} else if provNombre != "" {
			if pid, exists := provNameMap[strings.ToLower(provNombre)]; exists {
				proveedorID = &pid
			} else {
				result.Errores = append(result.Errores, domain.CSVFilaError{
					Fila:    rowNum,
					Codigo:  codigo,
					Campo:   "proveedor",
					Mensaje: fmt.Sprintf("el proveedor '%s' no existe", provNombre),
				})
				continue
			}
		}

		// Validar vencimiento
		vencStr := getCol("vencimiento")
		var vencimiento *string
		if vencStr != "" {
			if _, err := time.Parse("2006-01-02", vencStr); err != nil {
				result.Errores = append(result.Errores, domain.CSVFilaError{
					Fila:    rowNum,
					Codigo:  codigo,
					Campo:   "vencimiento",
					Mensaje: "formato de vencimiento inválido, debe ser AAAA-MM-DD",
				})
				continue
			}
			vencimiento = &vencStr
		}

		descStr := getCol("descripcion")
		var descripcion *string
		if descStr != "" {
			descripcion = &descStr
		}

		imgStr := getCol("imagen")
		var imagen *string
		if imgStr != "" {
			imagen = &imgStr
		}

		// Chequear existencia en base de datos
		existing, err := s.productRepo.FindByCode(ctx, codigo)
		isUpdate := false
		var existingID int64
		if err == nil && existing != nil {
			if modo == domain.CSVModeCrear {
				result.Errores = append(result.Errores, domain.CSVFilaError{
					Fila:    rowNum,
					Codigo:  codigo,
					Campo:   "codigo",
					Mensaje: fmt.Sprintf("el código '%s' ya existe en la base de datos (utilice modo 'actualizar' para modificarlo)", codigo),
				})
				continue
			}
			isUpdate = true
			existingID = existing.ID
		}

		prod := domain.Producto{
			Nombre:      nombre,
			Codigo:      codigo,
			Descripcion: descripcion,
			CategoriaID: categoriaID,
			Unidad:      unidad,
			CostoUnidad: costoUnidad,
			VentaUnidad: ventaUnidad,
			Costo:       costo,
			PrecioVenta: precioVenta,
			Stock:       stock,
			ProveedorID: proveedorID,
			Vencimiento: vencimiento,
			Imagen:      imagen,
			Estado:      estado,
		}

		validProducts = append(validProducts, validProductItem{
			product:    prod,
			isUpdate:   isUpdate,
			existingID: existingID,
		})
	}

	result.Rechazados = len(result.Errores)

	// Si hay errores, no persistir para evitar importación parcial inconsistente
	if len(result.Errores) > 0 {
		return result, nil
	}

	// Si previsualizar es true, solo totalizamos y retornamos
	if previsualizar {
		for _, item := range validProducts {
			if item.isUpdate {
				result.Actualizados++
			} else {
				result.Creados++
			}
		}
		return result, nil
	}

	// Ejecución efectiva dentro de transacción SQL
	tx, err := s.productRepo.BeginTx(ctx)
	if err != nil {
		return nil, fmt.Errorf("error al iniciar transacción: %w", err)
	}
	defer tx.Rollback()

	for _, item := range validProducts {
		if item.isUpdate {
			p := item.product
			p.ID = item.existingID
			if _, err := s.productRepo.UpdateTx(ctx, tx, item.existingID, &p); err != nil {
				return nil, fmt.Errorf("error al actualizar producto %s: %w", p.Codigo, err)
			}
			result.Actualizados++
		} else {
			p := item.product
			if _, err := s.productRepo.CreateTx(ctx, tx, &p); err != nil {
				return nil, fmt.Errorf("error al crear producto %s: %w", p.Codigo, err)
			}
			result.Creados++
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("error al confirmar transacción de importación: %w", err)
	}

	return result, nil
}

func (s *csvService) Export(ctx context.Context, w io.Writer, filter domain.ProductoFilter) error {
	products, err := s.productRepo.FindAllUnpaged(ctx, filter)
	if err != nil {
		return err
	}

	// Escribir BOM UTF-8 para garantizar apertura transparente en Excel
	if _, err := w.Write([]byte{0xEF, 0xBB, 0xBF}); err != nil {
		return err
	}

	csvWriter := csv.NewWriter(w)
	csvWriter.Comma = ','

	// Encabezados descriptivos y legibles
	headers := []string{
		"ID",
		"Código",
		"Nombre",
		"Descripción",
		"Categoría",
		"ID Categoría",
		"Unidad",
		"Unidad Costo",
		"Unidad Venta",
		"Costo",
		"Precio Venta",
		"Stock",
		"Proveedor",
		"ID Proveedor",
		"Vencimiento",
		"Estado",
		"Creado En",
	}

	if err := csvWriter.Write(headers); err != nil {
		return err
	}

	for _, p := range products {
		desc := ""
		if p.Descripcion != nil {
			desc = *p.Descripcion
		}
		catNom := ""
		if p.CategoriaNombre != nil {
			catNom = *p.CategoriaNombre
		}
		catID := ""
		if p.CategoriaID != nil {
			catID = strconv.FormatInt(*p.CategoriaID, 10)
		}
		provNom := ""
		if p.ProveedorNombre != nil {
			provNom = *p.ProveedorNombre
		}
		provID := ""
		if p.ProveedorID != nil {
			provID = strconv.FormatInt(*p.ProveedorID, 10)
		}
		venc := ""
		if p.Vencimiento != nil {
			venc = *p.Vencimiento
		}

		row := []string{
			strconv.FormatInt(p.ID, 10),
			p.Codigo,
			p.Nombre,
			desc,
			catNom,
			catID,
			p.Unidad,
			p.CostoUnidad,
			p.VentaUnidad,
			fmt.Sprintf("%.2f", p.Costo),
			fmt.Sprintf("%.2f", p.PrecioVenta),
			fmt.Sprintf("%.3f", p.Stock),
			provNom,
			provID,
			venc,
			p.Estado,
			p.CreadoEn.Format("2006-01-02 15:04:05"),
		}

		if err := csvWriter.Write(row); err != nil {
			return err
		}
	}

	csvWriter.Flush()
	return csvWriter.Error()
}
