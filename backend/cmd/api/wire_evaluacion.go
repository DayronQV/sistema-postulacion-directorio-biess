package main

import (
	"context"
	"time"

	apihttp "github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/http"
	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/http/handler"
	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/repository/postgres"
	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/service"
)

// evaluacionRoutes cablea la evaluación teórica del postulante (EF-14): el repositorio hace
// de banco de preguntas y de intentos, las fases dicen cuándo se puede rendir, el sorteo usa
// crypto/rand y el reloj es time.Now.
func evaluacionRoutes(c *container) apihttp.RouteRegistrar {
	return apihttp.EvaluacionRoutes(handler.NewEvaluacionHandler(nuevaEvaluacion(c)), c.tokens)
}

// nuevaEvaluacion construye el servicio de la evaluación: el repositorio hace de banco, de
// intentos y de respuestas, y el reporte se envía por el correo configurado.
func nuevaEvaluacion(c *container) *service.EvaluacionService {
	repo := postgres.NewEvaluacionRepo(c.db)
	return service.NewEvaluacionService(repo, repo, repo, postgres.NewFasesRepo(c.db), postgres.NewTxManager(c.db), service.BarajadorCrypto{}, time.Now).
		ConMailer(c.mailer, c.logger)
}

// cierreAutomaticoCada es cada cuánto se finalizan los exámenes vencidos que nadie finalizó.
const cierreAutomaticoCada = time.Minute

// iniciarCierreAutomatico finaliza en segundo plano, cada minuto, los exámenes cuya hora de
// cierre ya pasó (con las respuestas guardadas) y envía su reporte. Así un postulante que
// cierra el navegador también recibe su nota. Se detiene cuando se apaga el servidor.
func iniciarCierreAutomatico(ctx context.Context, c *container) {
	evaluacion := nuevaEvaluacion(c)
	go func() {
		ticker := time.NewTicker(cierreAutomaticoCada)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				vuelta, cancel := context.WithTimeout(ctx, cierreAutomaticoCada)
				n, err := evaluacion.FinalizarVencidos(vuelta)
				cancel()
				if err != nil {
					c.logger.Error("cierre automático de exámenes", "error", err)
				} else if n > 0 {
					c.logger.Info("exámenes vencidos finalizados", "cantidad", n)
				}
			}
		}
	}()
}
