package main

import (
	apihttp "github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/http"
	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/http/handler"
	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/repository/postgres"
	redisrepo "github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/repository/redis"
	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/service"
)

// catalogoRoutes cablea el módulo de catálogos: repositorios → servicio → handler.
func catalogoRoutes(c *container) apihttp.RouteRegistrar {
	catalogos := service.NewCatalogoService(
		postgres.NewCatalogoRepo(c.db),
		redisrepo.NewCacheCatalogos(c.redis),
		c.logger,
	)
	return apihttp.CatalogoRoutes(handler.NewCatalogoHandler(catalogos))
}
