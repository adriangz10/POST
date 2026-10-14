package http

import (
	"encoding/json"
	"net/http"
	"strconv"

	"post/internal/core/domain"
	"post/internal/core/ports"
)

type ProviderHandler struct {
	service ports.ProviderService
}

func NewProviderHandler(service ports.ProviderService) *ProviderHandler {
	return &ProviderHandler{service: service}
}

func (h *ProviderHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	search := r.URL.Query().Get("buscar")
	provs, err := h.service.GetAll(r.Context(), search)
	if err != nil {
		HandleError(w, err)
		return
	}
	JSON(w, http.StatusOK, provs)
}

func (h *ProviderHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		JSON(w, http.StatusBadRequest, ErrorResponse{Error: "ID de proveedor inválido"})
		return
	}

	prov, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		HandleError(w, err)
		return
	}
	JSON(w, http.StatusOK, prov)
}

func (h *ProviderHandler) Create(w http.ResponseWriter, r *http.Request) {
	var dto domain.CreateProveedorDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		JSON(w, http.StatusBadRequest, ErrorResponse{Error: "Cuerpo de solicitud malformado o JSON inválido"})
		return
	}

	prov, err := h.service.Create(r.Context(), dto)
	if err != nil {
		HandleError(w, err)
		return
	}
	JSON(w, http.StatusCreated, prov)
}

func (h *ProviderHandler) Update(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		JSON(w, http.StatusBadRequest, ErrorResponse{Error: "ID de proveedor inválido"})
		return
	}

	var dto domain.UpdateProveedorDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		JSON(w, http.StatusBadRequest, ErrorResponse{Error: "Cuerpo de solicitud malformado o JSON inválido"})
		return
	}

	updated, err := h.service.Update(r.Context(), id, dto)
	if err != nil {
		HandleError(w, err)
		return
	}
	JSON(w, http.StatusOK, updated)
}

func (h *ProviderHandler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		JSON(w, http.StatusBadRequest, ErrorResponse{Error: "ID de proveedor inválido"})
		return
	}

	if err := h.service.Delete(r.Context(), id); err != nil {
		HandleError(w, err)
		return
	}
	NoContent(w)
}

func (h *ProviderHandler) GetProducts(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		JSON(w, http.StatusBadRequest, ErrorResponse{Error: "ID de proveedor inválido"})
		return
	}

	products, err := h.service.GetProducts(r.Context(), id)
	if err != nil {
		HandleError(w, err)
		return
	}
	JSON(w, http.StatusOK, products)
}
