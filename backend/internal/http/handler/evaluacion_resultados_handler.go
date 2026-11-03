package handler

import (
	"context"
	"math"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/domain"
	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/http/middleware"
	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/service"
)

// ResultadosEvaluacionUseCase es lo que el handler necesita para mostrar los resultados de la
// evaluación al administrador.
type ResultadosEvaluacionUseCase interface {
	Resultados(ctx context.Context) ([]domain.ResultadoEvaluacion, error)
	Reporte(ctx context.Context, intentoID string) (service.ReportePDF, error)
}

// EvaluacionResultadosHandler atiende los resultados de la evaluación en el panel.
type EvaluacionResultadosHandler struct {
	resultados ResultadosEvaluacionUseCase
}

// NewEvaluacionResultadosHandler crea el handler.
func NewEvaluacionResultadosHandler(resultados ResultadosEvaluacionUseCase) *EvaluacionResultadosHandler {
	return &EvaluacionResultadosHandler{resultados: resultados}
}

// Estados de un examen en el listado de resultados.
const (
	estadoEnCurso    = "en_curso"
	estadoFinalizado = "finalizado"
)

type resultadoResponse struct {
	IntentoID      string     `json:"intento_id"`
	Codigo         string     `json:"codigo"`
	Cedula         string     `json:"cedula"`
	Nombres        string     `json:"nombres"`
	Apellidos      string     `json:"apellidos"`
	Cargo          string     `json:"cargo"`
	Estado         string     `json:"estado"`
	Inicio         time.Time  `json:"inicio"`
	Cierre         time.Time  `json:"cierre"`
	FinalizadoEn   *time.Time `json:"finalizado_at"`
	Nota           *float64   `json:"nota"`
	Aciertos       *int       `json:"aciertos"`
	TotalPreguntas int        `json:"total_preguntas"`
}

func respuestaResultado(r domain.ResultadoEvaluacion) resultadoResponse {
	out := resultadoResponse{
		IntentoID: r.IntentoID, Codigo: r.Codigo, Cedula: r.Cedula, Nombres: r.Nombres, Apellidos: r.Apellidos,
		Cargo: r.Cargo, Estado: estadoEnCurso, Inicio: r.Inicio.UTC(), Cierre: r.Cierre.UTC(),
		TotalPreguntas: r.TotalPreguntas,
	}
	if r.FinalizadoEn != nil && r.Nota != nil {
		fin := r.FinalizadoEn.UTC()
		nota := *r.Nota
		// La nota es el porcentaje de aciertos, así que los aciertos salen de ella.
		aciertos := int(math.Round(nota * float64(r.TotalPreguntas) / 100))
		out.Estado, out.FinalizadoEn, out.Nota, out.Aciertos = estadoFinalizado, &fin, &nota, &aciertos
	}
	return out
}

// Listar atiende GET /admin/evaluacion/resultados.
func (h *EvaluacionResultadosHandler) Listar(c *gin.Context) {
	resultados, err := h.resultados.Resultados(c.Request.Context())
	if err != nil {
		middleware.Fail(c, err)
		return
	}
	items := make([]resultadoResponse, len(resultados))
	for i, r := range resultados {
		items[i] = respuestaResultado(r)
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

// Reporte atiende GET /admin/evaluacion/resultados/{id}/reporte: descarga el PDF.
func (h *EvaluacionResultadosHandler) Reporte(c *gin.Context) {
	r, err := h.resultados.Reporte(c.Request.Context(), c.Param("id"))
	if err != nil {
		middleware.Fail(c, err)
		return
	}
	c.Header("Content-Disposition", disposicionAdjunto(r.Nombre))
	c.Data(http.StatusOK, "application/pdf", r.Contenido)
}
