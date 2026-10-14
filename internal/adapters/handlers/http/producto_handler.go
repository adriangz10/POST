package http

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"post/internal/core/domain"
	"post/internal/core/ports"
)

type ProductHandler struct {
	service ports.ProductService
}

func NewProductHandler(service ports.ProductService) *ProductHandler {
	return &ProductHandler{service: service}
}

func (h *ProductHandler) parseFilter(r *http.Request) domain.ProductoFilter {
	q := r.URL.Query()
	var filter domain.ProductoFilter

	if cidStr := q.Get("categoria_id"); cidStr != "" {
		if cid, err := strconv.ParseInt(cidStr, 10, 64); err == nil {
			filter.CategoriaID = &cid
		}
	}

	if pidStr := q.Get("proveedor_id"); pidStr != "" {
		if pid, err := strconv.ParseInt(pidStr, 10, 64); err == nil {
			filter.ProveedorID = &pid
		}
	}

	if estado := strings.TrimSpace(q.Get("estado")); estado != "" {
		filter.Estado = &estado
	}

	if buscar := strings.TrimSpace(q.Get("buscar")); buscar != "" {
		filter.Buscar = &buscar
	}

	if vd := strings.TrimSpace(q.Get("vencimiento_desde")); vd != "" {
		filter.VencimientoDesde = &vd
	}

	if vh := strings.TrimSpace(q.Get("vencimiento_hasta")); vh != "" {
		filter.VencimientoHasta = &vh
	}

	if sminStr := q.Get("stock_min"); sminStr != "" {
		if smin, err := strconv.ParseFloat(sminStr, 64); err == nil {
			filter.StockMin = &smin
		}
	}

	if smaxStr := q.Get("stock_max"); smaxStr != "" {
		if smax, err := strconv.ParseFloat(smaxStr, 64); err == nil {
			filter.StockMax = &smax
		}
	}

	if pminStr := q.Get("precio_min"); pminStr != "" {
		if pmin, err := strconv.ParseFloat(pminStr, 64); err == nil {
			filter.PrecioMin = &pmin
		}
	}

	if pmaxStr := q.Get("precio_max"); pmaxStr != "" {
		if pmax, err := strconv.ParseFloat(pmaxStr, 64); err == nil {
			filter.PrecioMax = &pmax
		}
	}

	if sbStr := q.Get("stock_bajo"); sbStr != "" {
		sb := strings.ToLower(sbStr) == "true" || sbStr == "1"
		filter.StockBajo = &sb
	}

	if umbralStr := q.Get("umbral_stock_bajo"); umbralStr != "" {
		if umbral, err := strconv.ParseFloat(umbralStr, 64); err == nil {
			filter.UmbralStockBajo = &umbral
		}
	}

	filter.OrdenPor = q.Get("orden_por")
	filter.Direccion = q.Get("direccion")

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

func (h *ProductHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	filter := h.parseFilter(r)
	result, err := h.service.GetAll(r.Context(), filter)
	if err != nil {
		HandleError(w, err)
		return
	}
	JSON(w, http.StatusOK, result)
}

func (h *ProductHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		JSON(w, http.StatusBadRequest, ErrorResponse{Error: "ID de producto inválido"})
		return
	}

	prod, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		HandleError(w, err)
		return
	}
	JSON(w, http.StatusOK, prod)
}

func (h *ProductHandler) Create(w http.ResponseWriter, r *http.Request) {
	var dto domain.CreateProductoDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		JSON(w, http.StatusBadRequest, ErrorResponse{Error: "Cuerpo de solicitud malformado o JSON inválido"})
		return
	}

	prod, err := h.service.Create(r.Context(), dto)
	if err != nil {
		HandleError(w, err)
		return
	}
	JSON(w, http.StatusCreated, prod)
}

func (h *ProductHandler) Update(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		JSON(w, http.StatusBadRequest, ErrorResponse{Error: "ID de producto inválido"})
		return
	}

	var dto domain.UpdateProductoDTO
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

func (h *ProductHandler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		JSON(w, http.StatusBadRequest, ErrorResponse{Error: "ID de producto inválido"})
		return
	}

	if err := h.service.Delete(r.Context(), id); err != nil {
		HandleError(w, err)
		return
	}
	NoContent(w)
}
