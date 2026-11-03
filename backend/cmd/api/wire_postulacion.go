package main

import (
	"github.com/gin-gonic/gin"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/domain"
	apihttp "github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/http"
	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/http/handler"
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
		c.registroCivil,
		postgres.NewTxManager(c.db),
		seccionesPostulacion(c)...,
	).ConCatalogos(service.NewCatalogoService(
		postgres.NewCatalogoRepo(c.db),
		redisrepo.NewCacheCatalogos(c.redis),
		c.logger,
	)).ConHistorialLaboral(c.historial).ConMailer(c.mailer).
		ConFases(postgres.NewFasesRepo(c.db))
	documentos := service.NewDocumentoService(postgres.NewDocumentoRepo(c.db)).
		ConFases(postgres.NewFasesRepo(c.db))
	postulacion := apihttp.PostulacionRoutes(handler.NewPostulacionHandler(svc), c.tokens)
	cargo := apihttp.CargoRoutes(handler.NewCargoHandler(svc), c.tokens)
	archivos := apihttp.DocumentoRoutes(handler.NewDocumentoHandler(documentos), c.tokens)
	return func(api *gin.RouterGroup) {
		postulacion(api)
		cargo(api)
		archivos(api)
	}
}

// seccionesPostulacion conserva una sola lista de repositorios para la vista del
// postulante y el detalle de solo lectura del panel administrativo.
func seccionesPostulacion(c *container) []domain.SeccionPostulacion {
	return []domain.SeccionPostulacion{
		postgres.NewFamiliaresRepo(c.db),
		postgres.NewReferenciasRepo(c.db),
		postgres.NewFormacionRepo(c.db),
		postgres.NewExperienciaRepo(c.db),
		postgres.NewPublicacionesRepo(c.db),
		postgres.NewCapacitacionesRepo(c.db),
		postgres.NewInhabilidadesRepo(c.db),
	}
}
