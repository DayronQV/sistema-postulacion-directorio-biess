package main

import (
	apihttp "github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/http"
	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/http/handler"
	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/repository/postgres"
	redisrepo "github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/repository/redis"
	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/service"
)

// adminRoutes cablea el panel de administración (EF-13): login con su propio contador de
// intentos (auth:*:admin:{usuario}) y las rutas protegidas por rol.
//
// Para sumar un módulo del panel: construir su handler aquí y agregar a AdminRoutes una
// función apihttp.ModuloAdmin que registre sus rutas de lectura y de escritura.
func adminRoutes(c *container) apihttp.RouteRegistrar {
	login := service.NewAdminAuthService(
		postgres.NewUsuariosAdminRepo(c.db),
		redisrepo.NewIntentos(c.redis, "admin"),
		c.tokens,
	)
	postulaciones := postgres.NewPostulacionRepo(c.db)
	postulacionAdmin := service.NewAdminPostulacionService(postulaciones).ConConsultas(
		postulaciones,
		postgres.NewDocumentoRepo(c.db),
		seccionesPostulacion(c)...,
	)
	meritos := handler.NewAdminPostulacionHandler(postulacionAdmin).ConConsultas(postulacionAdmin)
	fases := handler.NewAdminFasesHandler(service.NewFaseService(postgres.NewFasesRepo(c.db)))
	dashboard := handler.NewAdminDashboardHandler(
		service.NewDashboardService(postgres.NewPostulacionRepo(c.db)),
	)
	evaluacion := handler.NewEvaluacionAdminHandler(
		service.NewEvaluacionAdminService(postgres.NewEvaluacionRepo(c.db)),
	)
	return apihttp.AdminRoutes(handler.NewAdminAuthHandler(login), c.tokens,
		dashboard.Rutas,
		meritos.Rutas,
		fases.Rutas,
		apihttp.EvaluacionAdminModulo(evaluacion),
		apihttp.EvaluacionResultadosModulo(handler.NewEvaluacionResultadosHandler(nuevaEvaluacion(c))),
	)
}
