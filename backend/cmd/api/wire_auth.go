package main

import (
	apihttp "github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/http"
	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/http/handler"
	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/integration/registrocivil"
	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/repository/postgres"
	redisrepo "github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/repository/redis"
	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/service"
)

// authRoutes cablea el módulo de autenticación: repositorios → servicios → handler.
// Registro y login usan contadores de intentos separados (auth:*:registro:* y
// auth:*:login:*).
func authRoutes(c *container) apihttp.RouteRegistrar {
	postulantes := postgres.NewPostulanteRepo(c.db)
	registro := service.NewRegistroService(
		registrocivil.NewMock(),
		redisrepo.NewIntentos(c.redis, "registro"),
		redisrepo.NewTicketsRegistro(c.redis),
		postulantes,
	)
	login := service.NewAuthService(
		postulantes,
		redisrepo.NewIntentos(c.redis, "login"),
		c.tokens,
	)
	return apihttp.AuthRoutes(handler.NewAuthHandler(registro, login))
}
