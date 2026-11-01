package main

import (
	apihttp "github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/http"
	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/http/handler"
	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/integration/registrocivil"
	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/repository/postgres"
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
	)
	return apihttp.PostulacionRoutes(handler.NewPostulacionHandler(svc), c.tokens)
}
