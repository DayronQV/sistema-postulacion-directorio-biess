package handler

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/domain"
	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/http/middleware"
	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/service"
)

// EvaluacionUseCase es lo que el handler necesita del servicio de la evaluación del postulante.
type EvaluacionUseCase interface {
	Iniciar(ctx context.Context, postulanteID string) (service.SesionEvaluacion, bool, error)
	Pregunta(ctx context.Context, postulanteID string, n int) (service.PreguntaEvaluacion, error)
}

// EvaluacionHandler atiende al postulante en la evaluación teórica.
type EvaluacionHandler struct {
	evaluacion EvaluacionUseCase
}

// NewEvaluacionHandler crea el handler.
func NewEvaluacionHandler(evaluacion EvaluacionUseCase) *EvaluacionHandler {
	return &EvaluacionHandler{evaluacion: evaluacion}
}

// sesionResponse es la respuesta de iniciar la evaluación. Las horas van en UTC y el tiempo
// restante lo calcula el servidor.
type sesionResponse struct {
	TotalPreguntas         int       `json:"total_preguntas"`
	Inicio                 time.Time `json:"inicio"`
	Cierre                 time.Time `json:"cierre"`
	TiempoRestanteSegundos int       `json:"tiempo_restante_segundos"`
}

// preguntaEvaluacionResponse es una pregunta del examen para el postulante. Es un DTO propio:
// no tiene campo de respuesta correcta, así que no hay forma de que salga por descuido.
type preguntaEvaluacionResponse struct {
	Numero                 int       `json:"numero"`
	Total                  int       `json:"total"`
	ID                     int       `json:"id"`
	Enunciado              string    `json:"enunciado"`
	Opciones               []string  `json:"opciones"`
	Cierre                 time.Time `json:"cierre"`
	TiempoRestanteSegundos int       `json:"tiempo_restante_segundos"`
}

// segundosRestantes redondea hacia abajo y nunca devuelve un valor negativo.
func segundosRestantes(d time.Duration) int {
	if d < 0 {
		return 0
	}
	return int(d / time.Second)
}

// Iniciar atiende POST /evaluacion/iniciar: 201 si crea el examen, 200 si devuelve el que ya
// estaba en curso.
func (h *EvaluacionHandler) Iniciar(c *gin.Context) {
	id, ok := usuarioID(c)
	if !ok {
		return
	}
	s, creado, err := h.evaluacion.Iniciar(c.Request.Context(), id)
	if err != nil {
		middleware.Fail(c, err)
		return
	}
	estado := http.StatusOK
	if creado {
		estado = http.StatusCreated
	}
	c.JSON(estado, sesionResponse{
		TotalPreguntas:         s.TotalPreguntas,
		Inicio:                 s.Inicio.UTC(),
		Cierre:                 s.Cierre.UTC(),
		TiempoRestanteSegundos: segundosRestantes(s.TiempoRestante),
	})
}

// Pregunta atiende GET /evaluacion/pregunta/{n}.
func (h *EvaluacionHandler) Pregunta(c *gin.Context) {
	id, ok := usuarioID(c)
	if !ok {
		return
	}
	n, err := strconv.Atoi(c.Param("n"))
	if err != nil {
		middleware.Fail(c, domain.NotFound(service.MsgPreguntaFueraDeRango))
		return
	}
	p, err := h.evaluacion.Pregunta(c.Request.Context(), id, n)
	if err != nil {
		middleware.Fail(c, err)
		return
	}
	c.JSON(http.StatusOK, preguntaEvaluacionResponse{
		Numero:                 p.Numero,
		Total:                  p.Total,
		ID:                     p.Pregunta.ID,
		Enunciado:              p.Pregunta.Enunciado,
		Opciones:               p.Pregunta.Opciones,
		Cierre:                 p.Cierre.UTC(),
		TiempoRestanteSegundos: segundosRestantes(p.TiempoRestante),
	})
}
