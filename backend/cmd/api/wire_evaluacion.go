package main

import (
	"time"

	apihttp "github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/http"
	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/http/handler"
	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/repository/postgres"
	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/service"
)

// evaluacionRoutes cablea la evaluación teórica del postulante (EF-14): el repositorio hace
// de banco de preguntas y de intentos, el sorteo usa crypto/rand y el reloj es time.Now.
func evaluacionRoutes(c *container) apihttp.RouteRegistrar {
	repo := postgres.NewEvaluacionRepo(c.db)
	evaluacion := service.NewEvaluacionService(repo, repo, service.BarajadorCrypto{}, time.Now)
	return apihttp.EvaluacionRoutes(handler.NewEvaluacionHandler(evaluacion), c.tokens)
}
