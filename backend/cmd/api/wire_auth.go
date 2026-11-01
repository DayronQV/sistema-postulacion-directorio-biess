package main

import (
	apihttp "github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/http"
	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/http/handler"
	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/integration/registrocivil"
	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/repository/postgres"
	redisrepo "github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/repository/redis"
	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/service"
)

// authRoutes cablea el módulo de autenticación: repositorios → servicio → handler.
func authRoutes(c *container) apihttp.RouteRegistrar {
	registro := service.NewRegistroService(
		registrocivil.NewMock(),
		redisrepo.NewIntentos(c.redis, "registro"),
		redisrepo.NewTicketsRegistro(c.redis),
		postgres.NewPostulanteRepo(c.db),
	)
	return apihttp.AuthRoutes(handler.NewAuthHandler(registro))
}
