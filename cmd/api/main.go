package main

import (
	"log"
	"net/http"
	"os"

	httpAdapter "post/internal/adapters/handlers/http"
	"post/internal/adapters/storage/postgres"
	"post/internal/core/services"

	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load(".env")

	// Conexión a la base de datos PostgreSQL
	db, err := postgres.Connect()
	if err != nil {
		log.Fatalf("Error al conectar a PostgreSQL: %v", err)
	}
	defer db.Close()

	// 1. Instanciación de Repositorios (Adaptadores de Infraestructura)
	categoryRepo := postgres.NewCategoryRepository(db)
	providerRepo := postgres.NewProviderRepository(db)
	comboRepo := postgres.NewComboRepository(db)
	productRepo := postgres.NewProductRepository(db)

	// 2. Instanciación de Servicios de Aplicación (Casos de Uso)
	categoryService := services.NewCategoryService(categoryRepo)
	providerService := services.NewProviderService(providerRepo)
	comboService := services.NewComboService(comboRepo, productRepo, categoryRepo)
	productService := services.NewProductService(productRepo, categoryRepo, providerRepo, comboRepo)
	priceService := services.NewPriceService(productRepo, categoryRepo, providerRepo)
	csvService := services.NewCSVService(productRepo, categoryRepo, providerRepo)

	// 3. Instanciación de Controladores/Handlers HTTP (Adaptadores de Entrada)
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

	// 4. Configuración del Router con prefijo centralizado /api
	router := httpAdapter.SetupRouter(handlers)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	serverAddr := ":" + port
	log.Printf("Iniciando servidor API REST de Inventario en http://localhost%s/api", serverAddr)
	if err := http.ListenAndServe(serverAddr, router); err != nil {
		log.Fatalf("Error en el servidor HTTP: %v", err)
	}
}
