package services

import (
	"context"
	"fmt"
	"math"
	"post/internal/core/domain"
	"post/internal/core/ports"
)

const ComboPolicyStatement = "Los precios de venta de los combos se definen de manera independiente para permitir estrategias comerciales y promocionales. La actualización masiva de precios de productos solo afecta el precio_venta de los productos, por lo que el precio_venta y el costo de los combos se mantienen inalterados (los costos de los combos solo se recalculan cuando cambia el costo base de adquisición de sus productos componentes)."

type priceService struct {
	productRepo  ports.ProductRepository
	categoryRepo ports.CategoryRepository
	providerRepo ports.ProviderRepository
}

func NewPriceService(
	productRepo ports.ProductRepository,
	categoryRepo ports.CategoryRepository,
	providerRepo ports.ProviderRepository,
) ports.PriceService {
	return &priceService{
		productRepo:  productRepo,
		categoryRepo: categoryRepo,
		providerRepo: providerRepo,
	}
}

func (s *priceService) validateScopeExistence(ctx context.Context, dto domain.ActualizacionPreciosDTO) error {
	if dto.Alcance == domain.AlcanceCategoria {
		if _, err := s.categoryRepo.FindByID(ctx, *dto.CategoriaID); err != nil {
			return &domain.ValidationError{Field: "categoria_id", Message: "la categoría especificada no existe"}
		}
	}
	if dto.Alcance == domain.AlcanceProveedor {
		if _, err := s.providerRepo.FindByID(ctx, *dto.ProveedorID); err != nil {
			return &domain.ValidationError{Field: "proveedor_id", Message: "el proveedor especificado no existe"}
		}
	}
	return nil
}

func (s *priceService) Preview(ctx context.Context, dto domain.ActualizacionPreciosDTO) (*domain.PreciosPreviewResult, error) {
	if err := dto.Validate(); err != nil {
		return nil, err
	}

	if err := s.validateScopeExistence(ctx, dto); err != nil {
		return nil, err
	}

	var targetID *int64
	if dto.Alcance == domain.AlcanceCategoria {
		targetID = dto.CategoriaID
	} else if dto.Alcance == domain.AlcanceProveedor {
		targetID = dto.ProveedorID
	}

	products, err := s.productRepo.GetProductsForPriceUpdate(ctx, dto.Alcance, targetID)
	if err != nil {
		return nil, err
	}

	previewItems := make([]domain.PreciosProductoPreview, 0, len(products))
	for _, p := range products {
		nuevoPrecio := domain.CalculateNewPrice(p.PrecioVenta, dto.Operacion, dto.Porcentaje)
		diferencia := math.Round((nuevoPrecio-p.PrecioVenta)*100) / 100
		previewItems = append(previewItems, domain.PreciosProductoPreview{
			ID:           p.ID,
			Codigo:       p.Codigo,
			Nombre:       p.Nombre,
			PrecioActual: p.PrecioVenta,
			NuevoPrecio:  nuevoPrecio,
			Diferencia:   diferencia,
		})
	}

	return &domain.PreciosPreviewResult{
		Operacion:      dto.Operacion,
		Porcentaje:     dto.Porcentaje,
		Alcance:        dto.Alcance,
		TotalAfectados: len(previewItems),
		Productos:      previewItems,
		PoliticaCombos: ComboPolicyStatement,
	}, nil
}

func (s *priceService) Update(ctx context.Context, dto domain.ActualizacionPreciosDTO) (*domain.PreciosUpdateResult, error) {
	if err := dto.Validate(); err != nil {
		return nil, err
	}

	if err := s.validateScopeExistence(ctx, dto); err != nil {
		return nil, err
	}

	var targetID *int64
	if dto.Alcance == domain.AlcanceCategoria {
		targetID = dto.CategoriaID
	} else if dto.Alcance == domain.AlcanceProveedor {
		targetID = dto.ProveedorID
	}

	affected, err := s.productRepo.BulkUpdatePrices(ctx, dto.Operacion, dto.Porcentaje, dto.Alcance, targetID)
	if err != nil {
		return nil, err
	}

	msg := fmt.Sprintf("Se actualizaron exitosamente los precios de venta de %d productos mediante %s del %.2f%% con alcance '%s'",
		affected, dto.Operacion, dto.Porcentaje, dto.Alcance)

	return &domain.PreciosUpdateResult{
		Operacion:      dto.Operacion,
		Porcentaje:     dto.Porcentaje,
		Alcance:        dto.Alcance,
		Actualizados:   affected,
		Mensaje:        msg,
		PoliticaCombos: ComboPolicyStatement,
	}, nil
}
