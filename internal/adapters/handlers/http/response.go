package http

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"post/internal/core/domain"
)

type ErrorResponse struct {
	Error   string `json:"error"`
	Detalle any    `json:"detalle,omitempty"`
}

func JSON(w http.ResponseWriter, statusCode int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)
	if data != nil {
		if err := json.NewEncoder(w).Encode(data); err != nil {
			log.Printf("Error encoding response: %v", err)
		}
	}
}

func NoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

func HandleError(w http.ResponseWriter, err error) {
	if err == nil {
		return
	}

	var valErr *domain.ValidationError
	if errors.As(err, &valErr) {
		JSON(w, http.StatusUnprocessableEntity, ErrorResponse{
			Error:   "Error de validación",
			Detalle: valErr,
		})
		return
	}

	if errors.Is(err, domain.ErrNotFound) {
		JSON(w, http.StatusNotFound, ErrorResponse{
			Error: "Recurso no encontrado",
		})
		return
	}

	if errors.Is(err, domain.ErrDuplicateKey) {
		JSON(w, http.StatusConflict, ErrorResponse{
			Error: "Conflicto de unicidad: el recurso o uno de sus campos únicos ya existe",
		})
		return
	}

	if errors.Is(err, domain.ErrConflict) {
		JSON(w, http.StatusConflict, ErrorResponse{
			Error: "Conflicto: la operación no puede realizarse debido a relaciones activas",
		})
		return
	}

	if errors.Is(err, domain.ErrForeignKeyViolated) {
		JSON(w, http.StatusUnprocessableEntity, ErrorResponse{
			Error: "Violación de integridad referencial: una entidad relacionada no existe o no permite la operación",
		})
		return
	}

	if errors.Is(err, domain.ErrInvalidInput) {
		JSON(w, http.StatusBadRequest, ErrorResponse{
			Error: "Datos de entrada inválidos",
		})
		return
	}

	// Error no esperado del servidor
	log.Printf("[ERROR INTERNO] %v", err)
	JSON(w, http.StatusInternalServerError, ErrorResponse{
		Error: "Ha ocurrido un error interno en el servidor",
	})
}
