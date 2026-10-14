package http

import (
	"log"
	"net/http"
	"time"
)

type Handlers struct {
	CategoryHandler *CategoryHandler
	ProviderHandler *ProviderHandler
	ProductHandler  *ProductHandler
	ComboHandler    *ComboHandler
	PriceHandler    *PriceHandler
	CSVHandler      *CSVHandler
}

// loggingMiddleware registra cada solicitud HTTP con método, ruta, estado y duración
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("[%s] %s (%s)", r.Method, r.URL.Path, time.Since(start))
	})
}

// corsMiddleware añade cabeceras para permitir consumo desde navegadores y clientes frontend
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS, PUT")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		w.Header().Set("Access-Control-Expose-Headers", "*")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func SetupRouter(h *Handlers) http.Handler {
	mux := http.NewServeMux()

	// Healthcheck
	mux.HandleFunc("GET /api/health", func(w http.ResponseWriter, r *http.Request) {
		JSON(w, http.StatusOK, map[string]string{
			"estado":  "operativo",
			"sistema": "API REST Gestión de Inventario",
		})
	})

	// Categorías
	mux.HandleFunc("GET /api/categorias", h.CategoryHandler.GetAll)
	mux.HandleFunc("GET /api/categorias/{id}", h.CategoryHandler.GetByID)
	mux.HandleFunc("POST /api/categorias", h.CategoryHandler.Create)
	mux.HandleFunc("PATCH /api/categorias/{id}", h.CategoryHandler.Update)
	mux.HandleFunc("DELETE /api/categorias/{id}", h.CategoryHandler.Delete)

	// Proveedores
	mux.HandleFunc("GET /api/proveedores", h.ProviderHandler.GetAll)
	mux.HandleFunc("GET /api/proveedores/{id}", h.ProviderHandler.GetByID)
	mux.HandleFunc("POST /api/proveedores", h.ProviderHandler.Create)
	mux.HandleFunc("PATCH /api/proveedores/{id}", h.ProviderHandler.Update)
	mux.HandleFunc("DELETE /api/proveedores/{id}", h.ProviderHandler.Delete)
	mux.HandleFunc("GET /api/proveedores/{id}/productos", h.ProviderHandler.GetProducts)

	// Actualización masiva de precios (deben registrarse antes de /{id} para evitar colisiones)
	mux.HandleFunc("POST /api/productos/precios/previsualizar", h.PriceHandler.Preview)
	mux.HandleFunc("POST /api/productos/precios/actualizar", h.PriceHandler.Update)

	// Importación y exportación CSV
	mux.HandleFunc("POST /api/productos/importar-csv", h.CSVHandler.Import)
	mux.HandleFunc("GET /api/productos/exportar-csv", h.CSVHandler.Export)

	// Productos
	mux.HandleFunc("GET /api/productos", h.ProductHandler.GetAll)
	mux.HandleFunc("GET /api/productos/{id}", h.ProductHandler.GetByID)
	mux.HandleFunc("POST /api/productos", h.ProductHandler.Create)
	mux.HandleFunc("PATCH /api/productos/{id}", h.ProductHandler.Update)
	mux.HandleFunc("DELETE /api/productos/{id}", h.ProductHandler.Delete)

	// Componentes de Combos
	mux.HandleFunc("POST /api/combos/{id}/productos", h.ComboHandler.AddComponent)
	mux.HandleFunc("PATCH /api/combos/{id}/productos/{producto_id}", h.ComboHandler.UpdateComponent)
	mux.HandleFunc("DELETE /api/combos/{id}/productos/{producto_id}", h.ComboHandler.RemoveComponent)

	// Combos
	mux.HandleFunc("GET /api/combos", h.ComboHandler.GetAll)
	mux.HandleFunc("GET /api/combos/{id}", h.ComboHandler.GetByID)
	mux.HandleFunc("POST /api/combos", h.ComboHandler.Create)
	mux.HandleFunc("PATCH /api/combos/{id}", h.ComboHandler.Update)
	mux.HandleFunc("DELETE /api/combos/{id}", h.ComboHandler.Delete)

	return corsMiddleware(loggingMiddleware(mux))
}
