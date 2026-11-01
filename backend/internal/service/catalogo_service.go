package service

import (
	"context"
	"log/slog"
	"strconv"
	"time"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/domain"
)

// TTLCatalogos es el tiempo que una lista de catálogo permanece en la caché.
const TTLCatalogos = time.Hour

// CatalogoService sirve los catálogos geográficos con el patrón cache-aside: primero
// la caché, y si la lista no está, la base de datos y luego se guarda en la caché.
// La caché es solo una optimización: si falla, se responde desde la base.
type CatalogoService struct {
	catalogos domain.Catalogos
	cache     domain.CacheCatalogos
	logger    *slog.Logger
}

// NewCatalogoService crea el servicio con sus dependencias inyectadas.
func NewCatalogoService(catalogos domain.Catalogos, cache domain.CacheCatalogos, logger *slog.Logger) *CatalogoService {
	return &CatalogoService{catalogos: catalogos, cache: cache, logger: logger}
}

// Paises lista los países, con Ecuador primero.
func (s *CatalogoService) Paises(ctx context.Context) ([]domain.ItemCatalogo, error) {
	return s.listar(ctx, "paises", s.catalogos.Paises)
}

// Provincias lista las provincias de un país. Devuelve domain.NotFound si el país no
// existe.
func (s *CatalogoService) Provincias(ctx context.Context, paisID int) ([]domain.ItemCatalogo, error) {
	return s.listar(ctx, "provincias:"+strconv.Itoa(paisID), func(ctx context.Context) ([]domain.ItemCatalogo, error) {
		return s.catalogos.Provincias(ctx, paisID)
	})
}

// Cantones lista los cantones de una provincia. Devuelve domain.NotFound si la
// provincia no existe.
func (s *CatalogoService) Cantones(ctx context.Context, provinciaID int) ([]domain.ItemCatalogo, error) {
	return s.listar(ctx, "cantones:"+strconv.Itoa(provinciaID), func(ctx context.Context) ([]domain.ItemCatalogo, error) {
		return s.catalogos.Cantones(ctx, provinciaID)
	})
}

// listar aplica cache-aside a una lista. Un fallo de la caché se registra y se sigue
// con la base de datos; si falló la lectura no se intenta guardar, para no sumar otra
// espera contra una caché caída. Lo que no existe (NotFound) nunca se guarda.
func (s *CatalogoService) listar(
	ctx context.Context,
	clave string,
	cargar func(context.Context) ([]domain.ItemCatalogo, error),
) ([]domain.ItemCatalogo, error) {
	items, hit, err := s.cache.Obtener(ctx, clave)
	cacheDisponible := err == nil
	switch {
	case err != nil:
		// Si la petición ya está cancelada, el fallo es del cliente, no de la caché.
		if ctx.Err() == nil {
			s.logger.WarnContext(ctx, "caché de catálogos no disponible; se consulta PostgreSQL",
				"clave", clave, "error", err)
		}
	case hit:
		return items, nil
	}

	items, err = cargar(ctx)
	if err != nil {
		return nil, err
	}
	if items == nil {
		items = []domain.ItemCatalogo{}
	}

	if cacheDisponible {
		if err := s.cache.Guardar(ctx, clave, items, TTLCatalogos); err != nil {
			s.logger.WarnContext(ctx, "no se pudo guardar el catálogo en la caché",
				"clave", clave, "error", err)
		}
	}
	return items, nil
}
