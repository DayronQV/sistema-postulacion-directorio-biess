package main

import (
	"github.com/gin-gonic/gin"

	apihttp "github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/http"
	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/http/handler"
	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/integration/registrocivil"
	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/repository/postgres"
	redisrepo "github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/repository/redis"
	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/service"
)

// postulacionRoutes cablea el módulo de postulación: repositorios → servicio → handler.
//
// Para conectar una sección nueva (formación, experiencia, ...): agregar su repositorio
// a la lista de secciones de abajo. La transacción no cambia.
func postulacionRoutes(c *container) apihttp.RouteRegistrar {
	svc := service.NewPostulacionService(
		postgres.NewPostulacionRepo(c.db),
		registrocivil.NewMock(),
		postgres.NewTxManager(c.db),
		postgres.NewFamiliaresRepo(c.db),
	).ConCatalogos(service.NewCatalogoService(
		postgres.NewCatalogoRepo(c.db),
		redisrepo.NewCacheCatalogos(c.redis),
		c.logger,
	))
	documentos := service.NewDocumentoService(postgres.NewDocumentoRepo(c.db))
	postulacion := apihttp.PostulacionRoutes(handler.NewPostulacionHandler(svc), c.tokens)
	archivos := apihttp.DocumentoRoutes(handler.NewDocumentoHandler(documentos), c.tokens)
	return func(api *gin.RouterGroup) {
		postulacion(api)
		archivos(api)
	}
}
