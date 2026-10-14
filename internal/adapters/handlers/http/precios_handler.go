package http

import (
	"encoding/json"
	"net/http"

	"post/internal/core/domain"
	"post/internal/core/ports"
)

type PriceHandler struct {
	service ports.PriceService
}

func NewPriceHandler(service ports.PriceService) *PriceHandler {
	return &PriceHandler{service: service}
}

func (h *PriceHandler) Preview(w http.ResponseWriter, r *http.Request) {
	var dto domain.ActualizacionPreciosDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		JSON(w, http.StatusBadRequest, ErrorResponse{Error: "Cuerpo de solicitud malformado o JSON inválido"})
		return
	}

	result, err := h.service.Preview(r.Context(), dto)
	if err != nil {
		HandleError(w, err)
		return
	}
	JSON(w, http.StatusOK, result)
}

func (h *PriceHandler) Update(w http.ResponseWriter, r *http.Request) {
	var dto domain.ActualizacionPreciosDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		JSON(w, http.StatusBadRequest, ErrorResponse{Error: "Cuerpo de solicitud malformado o JSON inválido"})
		return
	}

	result, err := h.service.Update(r.Context(), dto)
	if err != nil {
		HandleError(w, err)
		return
	}
	JSON(w, http.StatusOK, result)
}
