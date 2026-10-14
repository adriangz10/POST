package http

import (
	"encoding/json"
	"net/http"
	"strconv"

	"post/internal/core/domain"
	"post/internal/core/ports"
)

type CategoryHandler struct {
	service ports.CategoryService
}

func NewCategoryHandler(service ports.CategoryService) *CategoryHandler {
	return &CategoryHandler{service: service}
}

func (h *CategoryHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	search := r.URL.Query().Get("buscar")
	cats, err := h.service.GetAll(r.Context(), search)
	if err != nil {
		HandleError(w, err)
		return
	}
	JSON(w, http.StatusOK, cats)
}

func (h *CategoryHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		JSON(w, http.StatusBadRequest, ErrorResponse{Error: "ID de categoría inválido"})
		return
	}

	cat, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		HandleError(w, err)
		return
	}
	JSON(w, http.StatusOK, cat)
}

func (h *CategoryHandler) Create(w http.ResponseWriter, r *http.Request) {
	var dto domain.CreateCategoriaDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		JSON(w, http.StatusBadRequest, ErrorResponse{Error: "Cuerpo de solicitud malformado o JSON inválido"})
		return
	}

	cat, err := h.service.Create(r.Context(), dto)
	if err != nil {
		HandleError(w, err)
		return
	}
	JSON(w, http.StatusCreated, cat)
}

func (h *CategoryHandler) Update(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		JSON(w, http.StatusBadRequest, ErrorResponse{Error: "ID de categoría inválido"})
		return
	}

	var dto domain.UpdateCategoriaDTO
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

func (h *CategoryHandler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		JSON(w, http.StatusBadRequest, ErrorResponse{Error: "ID de categoría inválido"})
		return
	}

	if err := h.service.Delete(r.Context(), id); err != nil {
		HandleError(w, err)
		return
	}
	NoContent(w)
}
