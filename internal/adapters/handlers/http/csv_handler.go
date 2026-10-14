package http

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"post/internal/core/domain"
	"post/internal/core/ports"
)

const MaxCSVFileSize = 10 * 1024 * 1024 // 10 MB

type CSVHandler struct {
	service        ports.CSVService
	productHandler *ProductHandler
}

func NewCSVHandler(service ports.CSVService, productHandler *ProductHandler) *CSVHandler {
	return &CSVHandler{
		service:        service,
		productHandler: productHandler,
	}
}

func (h *CSVHandler) Import(w http.ResponseWriter, r *http.Request) {
	// Limitar el tamaño máximo de lectura para evitar ataques por denegación de servicio (413)
	r.Body = http.MaxBytesReader(w, r.Body, MaxCSVFileSize)

	q := r.URL.Query()
	modo := domain.CSVImportMode(q.Get("modo"))
	if modo == "" {
		modo = domain.CSVModeCrear
	}
	previsualizar := q.Get("previsualizar") == "true" || q.Get("previsualizar") == "1"

	contentType := r.Header.Get("Content-Type")

	var reader io.Reader

	if strings.HasPrefix(contentType, "multipart/form-data") {
		err := r.ParseMultipartForm(MaxCSVFileSize)
		if err != nil {
			if strings.Contains(err.Error(), "too large") {
				JSON(w, http.StatusRequestEntityTooLarge, ErrorResponse{Error: "El archivo excede el tamaño máximo permitido de 10MB"})
				return
			}
			JSON(w, http.StatusBadRequest, ErrorResponse{Error: "Error al procesar formulario multipart: " + err.Error()})
			return
		}

		file, header, err := r.FormFile("archivo")
		if err != nil {
			// Intentar con nombre alternativo "file"
			file, header, err = r.FormFile("file")
			if err != nil {
				JSON(w, http.StatusBadRequest, ErrorResponse{Error: "Debe incluir un archivo con el campo 'archivo' o 'file'"})
				return
			}
		}
		defer file.Close()

		filename := strings.ToLower(header.Filename)
		if !strings.HasSuffix(filename, ".csv") && !strings.HasSuffix(filename, ".txt") {
			JSON(w, http.StatusUnsupportedMediaType, ErrorResponse{Error: "El archivo debe tener extensión .csv"})
			return
		}

		reader = file
	} else if strings.HasPrefix(contentType, "text/csv") || strings.HasPrefix(contentType, "text/plain") || contentType == "application/octet-stream" {
		reader = r.Body
	} else if contentType == "" {
		// Asumir cuerpo plano
		reader = r.Body
	} else {
		JSON(w, http.StatusUnsupportedMediaType, ErrorResponse{Error: "Formato no admitido. Utilice multipart/form-data o Content-Type: text/csv"})
		return
	}

	result, err := h.service.Import(r.Context(), reader, modo, previsualizar)
	if err != nil {
		if strings.Contains(err.Error(), "too large") {
			JSON(w, http.StatusRequestEntityTooLarge, ErrorResponse{Error: "El archivo excede el tamaño máximo permitido de 10MB"})
			return
		}
		HandleError(w, err)
		return
	}

	// Si hay errores de validación por fila, responder 422 con el detalle
	if len(result.Errores) > 0 {
		JSON(w, http.StatusUnprocessableEntity, result)
		return
	}

	JSON(w, http.StatusOK, result)
}

func (h *CSVHandler) Export(w http.ResponseWriter, r *http.Request) {
	filter := h.productHandler.parseFilter(r)

	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", "attachment; filename=\"productos.csv\"")

	if err := h.service.Export(r.Context(), w, filter); err != nil {
		// Como los encabezados ya pudieron haberse enviado, registramos el error
		fmt.Printf("Error al exportar CSV: %v\n", err)
	}
}
