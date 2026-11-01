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
// Registro, login y recuperación usan contadores de intentos separados
// (auth:*:registro:*, auth:*:login:* y auth:*:recuperar:*).
func authRoutes(c *container) apihttp.RouteRegistrar {
	registroCivil := registrocivil.NewMock()
	postulantes := postgres.NewPostulanteRepo(c.db)
	intentosLogin := redisrepo.NewIntentos(c.redis, "login")

	registro := service.NewRegistroService(
		registroCivil,
		redisrepo.NewIntentos(c.redis, "registro"),
		redisrepo.NewTicketsRegistro(c.redis),
		postulantes,
	)
	login := service.NewAuthService(postulantes, intentosLogin, c.tokens)
	recuperacion := service.NewRecuperacionService(
		registroCivil,
		postulantes,
		redisrepo.NewIntentos(c.redis, "recuperar"),
		intentosLogin,
		redisrepo.NewTokensRestablecer(c.redis),
		c.mailer,
		urlFrontend(c)+"/restablecer",
	)
	return apihttp.AuthRoutes(handler.NewAuthHandler(registro, login, recuperacion))
}

// urlFrontend es el origen del frontend, para armar los enlaces de los correos. Se
// toma el primer origen permitido por CORS (CORS_ALLOWED_ORIGIN), que es la URL
// del frontend; config ya garantiza que haya al menos uno.
func urlFrontend(c *container) string {
	return c.cfg.CORSAllowedOrigins[0]
}
