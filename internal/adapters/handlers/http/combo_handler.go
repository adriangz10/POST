package http

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"post/internal/core/domain"
	"post/internal/core/ports"
)

type ComboHandler struct {
	service ports.ComboService
}

func NewComboHandler(service ports.ComboService) *ComboHandler {
	return &ComboHandler{service: service}
}

func (h *ComboHandler) parseFilter(r *http.Request) domain.ComboFilter {
	q := r.URL.Query()
	var filter domain.ComboFilter

	if cidStr := q.Get("categoria_id"); cidStr != "" {
		if cid, err := strconv.ParseInt(cidStr, 10, 64); err == nil {
			filter.CategoriaID = &cid
		}
	}

	if estado := strings.TrimSpace(q.Get("estado")); estado != "" {
		filter.Estado = &estado
	}

	if buscar := strings.TrimSpace(q.Get("buscar")); buscar != "" {
		filter.Buscar = &buscar
	}

	if pageStr := q.Get("pagina"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil {
			filter.Pagina = p
		}
	}
	if limitStr := q.Get("limite"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			filter.Limite = l
		}
	}

	return filter
}

func (h *ComboHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	filter := h.parseFilter(r)
	result, err := h.service.GetAll(r.Context(), filter)
	if err != nil {
		HandleError(w, err)
		return
	}
	JSON(w, http.StatusOK, result)
}

func (h *ComboHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		JSON(w, http.StatusBadRequest, ErrorResponse{Error: "ID de combo inválido"})
		return
	}

	combo, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		HandleError(w, err)
		return
	}
	JSON(w, http.StatusOK, combo)
}

func (h *ComboHandler) Create(w http.ResponseWriter, r *http.Request) {
	var dto domain.CreateComboDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		JSON(w, http.StatusBadRequest, ErrorResponse{Error: "Cuerpo de solicitud malformado o JSON inválido"})
		return
	}

	combo, err := h.service.Create(r.Context(), dto)
	if err != nil {
		HandleError(w, err)
		return
	}
	JSON(w, http.StatusCreated, combo)
}

func (h *ComboHandler) Update(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		JSON(w, http.StatusBadRequest, ErrorResponse{Error: "ID de combo inválido"})
		return
	}

	var dto domain.UpdateComboDTO
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

func (h *ComboHandler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		JSON(w, http.StatusBadRequest, ErrorResponse{Error: "ID de combo inválido"})
		return
	}

	if err := h.service.Delete(r.Context(), id); err != nil {
		HandleError(w, err)
		return
	}
	NoContent(w)
}

func (h *ComboHandler) AddComponent(w http.ResponseWriter, r *http.Request) {
	comboIDStr := r.PathValue("id")
	comboID, err := strconv.ParseInt(comboIDStr, 10, 64)
	if err != nil {
		JSON(w, http.StatusBadRequest, ErrorResponse{Error: "ID de combo inválido"})
		return
	}

	var dto domain.ComboComponentInputDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		JSON(w, http.StatusBadRequest, ErrorResponse{Error: "Cuerpo de solicitud malformado o JSON inválido"})
		return
	}

	updated, err := h.service.AddComponent(r.Context(), comboID, dto)
	if err != nil {
		HandleError(w, err)
		return
	}
	JSON(w, http.StatusOK, updated)
}

func (h *ComboHandler) UpdateComponent(w http.ResponseWriter, r *http.Request) {
	comboIDStr := r.PathValue("id")
	comboID, err := strconv.ParseInt(comboIDStr, 10, 64)
	if err != nil {
		JSON(w, http.StatusBadRequest, ErrorResponse{Error: "ID de combo inválido"})
		return
	}

	prodIDStr := r.PathValue("producto_id")
	prodID, err := strconv.ParseInt(prodIDStr, 10, 64)
	if err != nil {
		JSON(w, http.StatusBadRequest, ErrorResponse{Error: "ID de producto inválido"})
		return
	}

	var dto domain.ComboComponentUpdateDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		JSON(w, http.StatusBadRequest, ErrorResponse{Error: "Cuerpo de solicitud malformado o JSON inválido"})
		return
	}

	updated, err := h.service.UpdateComponent(r.Context(), comboID, prodID, dto)
	if err != nil {
		HandleError(w, err)
		return
	}
	JSON(w, http.StatusOK, updated)
}

func (h *ComboHandler) RemoveComponent(w http.ResponseWriter, r *http.Request) {
	comboIDStr := r.PathValue("id")
	comboID, err := strconv.ParseInt(comboIDStr, 10, 64)
	if err != nil {
		JSON(w, http.StatusBadRequest, ErrorResponse{Error: "ID de combo inválido"})
		return
	}

	prodIDStr := r.PathValue("producto_id")
	prodID, err := strconv.ParseInt(prodIDStr, 10, 64)
	if err != nil {
		JSON(w, http.StatusBadRequest, ErrorResponse{Error: "ID de producto inválido"})
		return
	}

	updated, err := h.service.RemoveComponent(r.Context(), comboID, prodID)
	if err != nil {
		HandleError(w, err)
		return
	}
	JSON(w, http.StatusOK, updated)
}
