package tests

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	_ "github.com/lib/pq"
	httpAdapter "post/internal/adapters/handlers/http"
	"post/internal/adapters/storage/postgres"
	"post/internal/core/domain"
	"post/internal/core/services"
)

var (
	testDB     *sql.DB
	testServer *httptest.Server
)

func setupTestApp(t *testing.T) {
	if testServer != nil {
		return
	}

	// Cargar configuración de PostgreSQL
	os.Setenv("DB_NAME", "db_post")
	os.Setenv("DB_USER", "postgres")
	os.Setenv("DB_PASS", "adrian123")
	os.Setenv("DB_HOST", "localhost")
	os.Setenv("DB_PORT", "5432")
	os.Setenv("DB_SSLMODE", "disable")

	db, err := postgres.Connect()
	if err != nil {
		t.Fatalf("No se pudo conectar a PostgreSQL para pruebas: %v", err)
	}
	testDB = db

	categoryRepo := postgres.NewCategoryRepository(db)
	providerRepo := postgres.NewProviderRepository(db)
	comboRepo := postgres.NewComboRepository(db)
	productRepo := postgres.NewProductRepository(db)

	categoryService := services.NewCategoryService(categoryRepo)
	providerService := services.NewProviderService(providerRepo)
	comboService := services.NewComboService(comboRepo, productRepo, categoryRepo)
	productService := services.NewProductService(productRepo, categoryRepo, providerRepo, comboRepo)
	priceService := services.NewPriceService(productRepo, categoryRepo, providerRepo)
	csvService := services.NewCSVService(productRepo, categoryRepo, providerRepo)

	categoryHandler := httpAdapter.NewCategoryHandler(categoryService)
	providerHandler := httpAdapter.NewProviderHandler(providerService)
	productHandler := httpAdapter.NewProductHandler(productService)
	comboHandler := httpAdapter.NewComboHandler(comboService)
	priceHandler := httpAdapter.NewPriceHandler(priceService)
	csvHandler := httpAdapter.NewCSVHandler(csvService, productHandler)

	handlers := &httpAdapter.Handlers{
		CategoryHandler: categoryHandler,
		ProviderHandler: providerHandler,
		ProductHandler:  productHandler,
		ComboHandler:    comboHandler,
		PriceHandler:    priceHandler,
		CSVHandler:      csvHandler,
	}

	router := httpAdapter.SetupRouter(handlers)
	testServer = httptest.NewServer(router)
}

func cleanTables(t *testing.T) {
	ctx := context.Background()
	_, err := testDB.ExecContext(ctx, `
		DELETE FROM combo_productos;
		DELETE FROM combos;
		DELETE FROM productos;
		DELETE FROM categorias;
		DELETE FROM proveedores;
	`)
	if err != nil {
		t.Fatalf("Error limpiando tablas de prueba: %v", err)
	}
}

func TestCategoriasCRUD(t *testing.T) {
	setupTestApp(t)
	cleanTables(t)

	// 1. Crear categoría válida
	payload := map[string]string{
		"nombre": "Bebidas",
		"color":  "#00FF00",
	}
	b, _ := json.Marshal(payload)
	resp, err := http.Post(testServer.URL+"/api/categorias", "application/json", bytes.NewReader(b))
	if err != nil || resp.StatusCode != http.StatusCreated {
		t.Fatalf("Se esperaba 201 al crear categoría, obtenido %d", resp.StatusCode)
	}

	var created domain.Categoria
	_ = json.NewDecoder(resp.Body).Decode(&created)
	if created.ID <= 0 || created.Nombre != "Bebidas" {
		t.Fatalf("Respuesta inesperada al crear categoría: %+v", created)
	}

	// 2. Intentar duplicar nombre -> 409
	respDup, _ := http.Post(testServer.URL+"/api/categorias", "application/json", bytes.NewReader(b))
	if respDup.StatusCode != http.StatusConflict {
		t.Errorf("Se esperaba 409 al duplicar categoría, obtenido %d", respDup.StatusCode)
	}

	// 3. Crear con color inválido -> 422
	badColor := map[string]string{"nombre": "Golosinas", "color": "azul"}
	bBad, _ := json.Marshal(badColor)
	respBad, _ := http.Post(testServer.URL+"/api/categorias", "application/json", bytes.NewReader(bBad))
	if respBad.StatusCode != http.StatusUnprocessableEntity {
		t.Errorf("Se esperaba 422 al enviar color inválido, obtenido %d", respBad.StatusCode)
	}

	// 4. Obtener por ID
	respGet, _ := http.Get(fmt.Sprintf("%s/api/categorias/%d", testServer.URL, created.ID))
	if respGet.StatusCode != http.StatusOK {
		t.Errorf("Se esperaba 200 al consultar por ID, obtenido %d", respGet.StatusCode)
	}

	// 5. Actualizar categoría
	updatePayload := map[string]string{"nombre": "Bebidas y Jugos", "color": "#112233"}
	bUp, _ := json.Marshal(updatePayload)
	reqUp, _ := http.NewRequest(http.MethodPatch, fmt.Sprintf("%s/api/categorias/%d", testServer.URL, created.ID), bytes.NewReader(bUp))
	respUp, _ := http.DefaultClient.Do(reqUp)
	if respUp.StatusCode != http.StatusOK {
		t.Errorf("Se esperaba 200 al actualizar categoría, obtenido %d", respUp.StatusCode)
	}

	// 6. Eliminar categoría
	reqDel, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/api/categorias/%d", testServer.URL, created.ID), nil)
	respDel, _ := http.DefaultClient.Do(reqDel)
	if respDel.StatusCode != http.StatusNoContent {
		t.Errorf("Se esperaba 204 al eliminar categoría, obtenido %d", respDel.StatusCode)
	}
}

func TestProveedoresCRUD(t *testing.T) {
	setupTestApp(t)
	cleanTables(t)

	// Crear proveedor
	pPayload := map[string]string{"nombre": "Distribuidora Central"}
	b, _ := json.Marshal(pPayload)
	resp, err := http.Post(testServer.URL+"/api/proveedores", "application/json", bytes.NewReader(b))
	if err != nil || resp.StatusCode != http.StatusCreated {
		t.Fatalf("Se esperaba 201 al crear proveedor, obtenido %d", resp.StatusCode)
	}

	var prov domain.Proveedor
	_ = json.NewDecoder(resp.Body).Decode(&prov)

	// Consultar productos asociados (debe ser lista vacía al inicio)
	respProd, _ := http.Get(fmt.Sprintf("%s/api/proveedores/%d/productos", testServer.URL, prov.ID))
	if respProd.StatusCode != http.StatusOK {
		t.Errorf("Se esperaba 200 al consultar productos de proveedor, obtenido %d", respProd.StatusCode)
	}

	// Eliminar
	reqDel, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/api/proveedores/%d", testServer.URL, prov.ID), nil)
	respDel, _ := http.DefaultClient.Do(reqDel)
	if respDel.StatusCode != http.StatusNoContent {
		t.Errorf("Se esperaba 204 al eliminar proveedor, obtenido %d", respDel.StatusCode)
	}
}

func TestProductosCRUDAndFilters(t *testing.T) {
	setupTestApp(t)
	cleanTables(t)

	// Crear Categoría y Proveedor base
	cPayload, _ := json.Marshal(map[string]string{"nombre": "Lácteos", "color": "#FFFFFF"})
	respC, _ := http.Post(testServer.URL+"/api/categorias", "application/json", bytes.NewReader(cPayload))
	var cat domain.Categoria
	_ = json.NewDecoder(respC.Body).Decode(&cat)

	pPayload, _ := json.Marshal(map[string]string{"nombre": "La Serenísima"})
	respP, _ := http.Post(testServer.URL+"/api/proveedores", "application/json", bytes.NewReader(pPayload))
	var prov domain.Proveedor
	_ = json.NewDecoder(respP.Body).Decode(&prov)

	// 1. Crear Producto
	prodDTO := domain.CreateProductoDTO{
		Nombre:      "Leche Entera 1L",
		Codigo:      "LEC-001",
		CategoriaID: &cat.ID,
		ProveedorID: &prov.ID,
		Unidad:      "litro",
		CostoUnidad: "litro",
		VentaUnidad: "litro",
		Costo:       800.0,
		PrecioVenta: 1200.0,
		Stock:       15.0,
		Estado:      "activo",
	}
	bProd, _ := json.Marshal(prodDTO)
	respProd, err := http.Post(testServer.URL+"/api/productos", "application/json", bytes.NewReader(bProd))
	if err != nil || respProd.StatusCode != http.StatusCreated {
		t.Fatalf("Se esperaba 201 al crear producto, obtenido %d", respProd.StatusCode)
	}

	var prod domain.Producto
	_ = json.NewDecoder(respProd.Body).Decode(&prod)

	// 2. Duplicado de código -> 409
	respDup, _ := http.Post(testServer.URL+"/api/productos", "application/json", bytes.NewReader(bProd))
	if respDup.StatusCode != http.StatusConflict {
		t.Errorf("Se esperaba 409 al duplicar código de producto, obtenido %d", respDup.StatusCode)
	}

	// 3. Crear segundo producto con stock bajo
	prod2DTO := domain.CreateProductoDTO{
		Nombre:      "Yogur Frutilla",
		Codigo:      "YOG-002",
		CategoriaID: &cat.ID,
		ProveedorID: &prov.ID,
		Unidad:      "unitario",
		Costo:       300.0,
		PrecioVenta: 500.0,
		Stock:       2.0, // stock bajo
		Estado:      "activo",
	}
	bProd2, _ := json.Marshal(prod2DTO)
	_, _ = http.Post(testServer.URL+"/api/productos", "application/json", bytes.NewReader(bProd2))

	// 4. Filtrar por stock bajo
	respLowStock, _ := http.Get(testServer.URL + "/api/productos?stock_bajo=true&umbral_stock_bajo=5")
	if respLowStock.StatusCode != http.StatusOK {
		t.Errorf("Se esperaba 200 en filtro stock bajo")
	}
	var paged domain.PagedResult[domain.Producto]
	_ = json.NewDecoder(respLowStock.Body).Decode(&paged)
	if paged.Total != 1 || paged.Data[0].Codigo != "YOG-002" {
		t.Errorf("Filtro stock bajo incorrecto, total esperado 1, obtenido %d", paged.Total)
	}

	// 5. Filtrar por categoría
	respCatFilter, _ := http.Get(fmt.Sprintf("%s/api/productos?categoria_id=%d", testServer.URL, cat.ID))
	var pagedCat domain.PagedResult[domain.Producto]
	_ = json.NewDecoder(respCatFilter.Body).Decode(&pagedCat)
	if pagedCat.Total != 2 {
		t.Errorf("Filtro por categoría esperado 2, obtenido %d", pagedCat.Total)
	}
}

func TestCombosAndCostRecalculation(t *testing.T) {
	setupTestApp(t)
	cleanTables(t)

	// Crear productos para el combo
	// Producto A: Harina $1000/kg (base gramo -> $1/g)
	p1DTO := domain.CreateProductoDTO{
		Nombre:      "Harina 000",
		Codigo:      "HAR-1",
		Unidad:      "gramo",
		CostoUnidad: "kilogramo",
		Costo:       1000.0,
		PrecioVenta: 2.0,
		Stock:       10000.0,
	}
	bP1, _ := json.Marshal(p1DTO)
	respP1, _ := http.Post(testServer.URL+"/api/productos", "application/json", bytes.NewReader(bP1))
	var p1 domain.Producto
	_ = json.NewDecoder(respP1.Body).Decode(&p1)

	// Producto B: Queso $4000/kg (base gramo -> $4/g)
	p2DTO := domain.CreateProductoDTO{
		Nombre:      "Queso Muzzarella",
		Codigo:      "QUE-1",
		Unidad:      "gramo",
		CostoUnidad: "kilogramo",
		Costo:       4000.0,
		PrecioVenta: 7.0,
		Stock:       5000.0,
	}
	bP2, _ := json.Marshal(p2DTO)
	respP2, _ := http.Post(testServer.URL+"/api/productos", "application/json", bytes.NewReader(bP2))
	var p2 domain.Producto
	_ = json.NewDecoder(respP2.Body).Decode(&p2)

	// 1. Crear Combo con componentes
	// Harina 500g ($500) + Queso 200g ($800) = Costo total $1300
	comboDTO := domain.CreateComboDTO{
		Nombre:      "Combo Pizza Casera",
		Codigo:      "CMB-PIZZA",
		PrecioVenta: 2500.0,
		Componentes: []domain.CreateComboComponentDTO{
			{ProductoID: p1.ID, Cantidad: 500},
			{ProductoID: p2.ID, Cantidad: 200},
		},
	}
	bCombo, _ := json.Marshal(comboDTO)
	respCombo, err := http.Post(testServer.URL+"/api/combos", "application/json", bytes.NewReader(bCombo))
	if err != nil || respCombo.StatusCode != http.StatusCreated {
		t.Fatalf("Se esperaba 201 al crear combo, obtenido %d", respCombo.StatusCode)
	}

	var createdCombo domain.Combo
	_ = json.NewDecoder(respCombo.Body).Decode(&createdCombo)
	if createdCombo.Costo != 1300.0 {
		t.Errorf("Costo de combo esperado 1300.0, obtenido %.2f", createdCombo.Costo)
	}

	// 2. Modificar el costo del producto P1 (Harina pasa a $2000/kg -> $2/g)
	// Nuevo costo del combo esperado: 500g * $2 ($1000) + 200g * $4 ($800) = $1800
	newCosto := 2000.0
	upDTO := domain.UpdateProductoDTO{Costo: &newCosto}
	bUp, _ := json.Marshal(upDTO)
	reqUp, _ := http.NewRequest(http.MethodPatch, fmt.Sprintf("%s/api/productos/%d", testServer.URL, p1.ID), bytes.NewReader(bUp))
	_, _ = http.DefaultClient.Do(reqUp)

	// Consultar combo y verificar sincronización de costo
	respCheckCombo, _ := http.Get(fmt.Sprintf("%s/api/combos/%d", testServer.URL, createdCombo.ID))
	var updatedCombo domain.Combo
	_ = json.NewDecoder(respCheckCombo.Body).Decode(&updatedCombo)
	if updatedCombo.Costo != 1800.0 {
		t.Errorf("Costo recalculado de combo esperado 1800.0, obtenido %.2f", updatedCombo.Costo)
	}

	// 3. Probar restricción de borrado de producto en combo (ON DELETE RESTRICT)
	reqDelProd, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/api/productos/%d", testServer.URL, p1.ID), nil)
	respDelProd, _ := http.DefaultClient.Do(reqDelProd)
	if respDelProd.StatusCode != http.StatusConflict {
		t.Errorf("Se esperaba 409 Conflict al intentar borrar producto con combo asociado, obtenido %d", respDelProd.StatusCode)
	}
}

func TestBulkPriceUpdateAndPreview(t *testing.T) {
	setupTestApp(t)
	cleanTables(t)

	// Crear categoría
	cPayload, _ := json.Marshal(map[string]string{"nombre": "Gaseosas", "color": "#FF0000"})
	respC, _ := http.Post(testServer.URL+"/api/categorias", "application/json", bytes.NewReader(cPayload))
	var cat domain.Categoria
	_ = json.NewDecoder(respC.Body).Decode(&cat)

	// Crear 2 productos
	p1DTO := domain.CreateProductoDTO{
		Nombre:      "Cola 500ml",
		Codigo:      "COL-500",
		CategoriaID: &cat.ID,
		PrecioVenta: 100.0,
		Stock:       50,
	}
	bP1, _ := json.Marshal(p1DTO)
	_, _ = http.Post(testServer.URL+"/api/productos", "application/json", bytes.NewReader(bP1))

	p2DTO := domain.CreateProductoDTO{
		Nombre:      "Naranja 500ml",
		Codigo:      "NAR-500",
		CategoriaID: &cat.ID,
		PrecioVenta: 200.0,
		Stock:       50,
	}
	bP2, _ := json.Marshal(p2DTO)
	_, _ = http.Post(testServer.URL+"/api/productos", "application/json", bytes.NewReader(bP2))

	// 1. Previsualizar aumento 10%
	previewDTO := domain.ActualizacionPreciosDTO{
		Operacion:   domain.OperacionAumento,
		Porcentaje:  10.0,
		Alcance:     domain.AlcanceCategoria,
		CategoriaID: &cat.ID,
	}
	bPrev, _ := json.Marshal(previewDTO)
	respPrev, err := http.Post(testServer.URL+"/api/productos/precios/previsualizar", "application/json", bytes.NewReader(bPrev))
	if err != nil || respPrev.StatusCode != http.StatusOK {
		t.Fatalf("Se esperaba 200 en previsualización, obtenido %d", respPrev.StatusCode)
	}

	var prevResult domain.PreciosPreviewResult
	_ = json.NewDecoder(respPrev.Body).Decode(&prevResult)
	if prevResult.TotalAfectados != 2 {
		t.Errorf("Total afectados esperado 2, obtenido %d", prevResult.TotalAfectados)
	}

	// Verificar que los precios en la BD NO cambiaron
	respP1Check, _ := http.Get(testServer.URL + "/api/productos?buscar=COL-500")
	var pagedCheck domain.PagedResult[domain.Producto]
	_ = json.NewDecoder(respP1Check.Body).Decode(&pagedCheck)
	if pagedCheck.Data[0].PrecioVenta != 100.0 {
		t.Errorf("La previsualización NO debe modificar los precios, precio actual: %.2f", pagedCheck.Data[0].PrecioVenta)
	}

	// 2. Ejecutar actualización efectiva 10%
	respExec, err := http.Post(testServer.URL+"/api/productos/precios/actualizar", "application/json", bytes.NewReader(bPrev))
	if err != nil || respExec.StatusCode != http.StatusOK {
		t.Fatalf("Se esperaba 200 en actualización efectiva, obtenido %d", respExec.StatusCode)
	}

	var updateResult domain.PreciosUpdateResult
	_ = json.NewDecoder(respExec.Body).Decode(&updateResult)
	if updateResult.Actualizados != 2 {
		t.Errorf("Actualizados esperado 2, obtenido %d", updateResult.Actualizados)
	}

	// Verificar precios actualizados en BD: 100 -> 110, 200 -> 220
	respP1Updated, _ := http.Get(testServer.URL + "/api/productos?buscar=COL-500")
	var pagedUp domain.PagedResult[domain.Producto]
	_ = json.NewDecoder(respP1Updated.Body).Decode(&pagedUp)
	if pagedUp.Data[0].PrecioVenta != 110.0 {
		t.Errorf("Precio actualizado esperado 110.0, obtenido %.2f", pagedUp.Data[0].PrecioVenta)
	}
}

func TestCSVExportAndImport(t *testing.T) {
	setupTestApp(t)
	cleanTables(t)

	// Crear un producto inicial
	pDTO := domain.CreateProductoDTO{
		Nombre:      "Alfajor Chocolate",
		Codigo:      "ALF-001",
		Unidad:      "unitario",
		Costo:       200.0,
		PrecioVenta: 400.0,
		Stock:       100.0,
	}
	bP, _ := json.Marshal(pDTO)
	_, _ = http.Post(testServer.URL+"/api/productos", "application/json", bytes.NewReader(bP))

	// 1. Exportar CSV
	respExp, err := http.Get(testServer.URL + "/api/productos/exportar-csv")
	if err != nil || respExp.StatusCode != http.StatusOK {
		t.Fatalf("Se esperaba 200 al exportar CSV, obtenido %d", respExp.StatusCode)
	}
	csvContent, _ := io.ReadAll(respExp.Body)
	if !strings.Contains(string(csvContent), "ALF-001") {
		t.Errorf("El CSV exportado no contiene el código del producto")
	}

	// 2. Importar CSV en modo previsualización
	csvToImport := "codigo;nombre;unidad;costo;precio_venta;stock;estado\n" +
		"IMP-001;Galletitas;unitario;150.0;250.0;50;activo\n" +
		"IMP-002;Caramelos;unitario;50.0;100.0;200;activo\n"

	respPrev, err := http.Post(testServer.URL+"/api/productos/importar-csv?previsualizar=true", "text/csv", strings.NewReader(csvToImport))
	if err != nil || respPrev.StatusCode != http.StatusOK {
		t.Fatalf("Se esperaba 200 en importación previa, obtenido %d", respPrev.StatusCode)
	}
	var resPrev domain.CSVImportResult
	_ = json.NewDecoder(respPrev.Body).Decode(&resPrev)
	if resPrev.Creados != 2 || !resPrev.Previsualizacion {
		t.Errorf("Resultado de previsualización inesperado: %+v", resPrev)
	}

	// Verificar que no se hayan insertado aún en BD
	respCheck, _ := http.Get(testServer.URL + "/api/productos?buscar=IMP-001")
	var pagedCheck domain.PagedResult[domain.Producto]
	_ = json.NewDecoder(respCheck.Body).Decode(&pagedCheck)
	if pagedCheck.Total != 0 {
		t.Errorf("La previsualización no debió persistir productos")
	}

	// 3. Importar efectivamente con modo crear
	respImport, err := http.Post(testServer.URL+"/api/productos/importar-csv?modo=crear", "text/csv", strings.NewReader(csvToImport))
	if err != nil || respImport.StatusCode != http.StatusOK {
		t.Fatalf("Se esperaba 200 en importación efectiva, obtenido %d", respImport.StatusCode)
	}
	var resImport domain.CSVImportResult
	_ = json.NewDecoder(respImport.Body).Decode(&resImport)
	if resImport.Creados != 2 {
		t.Errorf("Se esperaba 2 creados, obtenido %d", resImport.Creados)
	}

	// 4. Importar con error de validación (precio negativo en fila 2) -> 422
	badCSV := "codigo;nombre;precio_venta\n" +
		"BAD-001;Producto Malo;-50.0\n"
	respBad, _ := http.Post(testServer.URL+"/api/productos/importar-csv", "text/csv", strings.NewReader(badCSV))
	if respBad.StatusCode != http.StatusUnprocessableEntity {
		t.Errorf("Se esperaba 422 en importación con errores, obtenido %d", respBad.StatusCode)
	}
	var resBad domain.CSVImportResult
	_ = json.NewDecoder(respBad.Body).Decode(&resBad)
	if len(resBad.Errores) == 0 || resBad.Errores[0].Fila != 2 {
		t.Errorf("Error por fila esperado en fila 2: %+v", resBad.Errores)
	}
}

func TestMultipartCSVImport(t *testing.T) {
	setupTestApp(t)
	cleanTables(t)

	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	part, err := w.CreateFormFile("archivo", "productos.csv")
	if err != nil {
		t.Fatalf("Error creando multipart form file: %v", err)
	}

	csvData := "codigo,nombre,unidad,costo,precio_venta,stock,estado\n" +
		"MULTI-01,Producto Multipart,unitario,100,200,10,activo\n"
	_, _ = part.Write([]byte(csvData))
	_ = w.Close()

	req, _ := http.NewRequest(http.MethodPost, testServer.URL+"/api/productos/importar-csv", &b)
	req.Header.Set("Content-Type", w.FormDataContentType())

	resp, err := http.DefaultClient.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		t.Fatalf("Se esperaba 200 en importación multipart, obtenido %d", resp.StatusCode)
	}

	var res domain.CSVImportResult
	_ = json.NewDecoder(resp.Body).Decode(&res)
	if res.Creados != 1 {
		t.Errorf("Se esperaba 1 producto creado, obtenido %d", res.Creados)
	}
}
