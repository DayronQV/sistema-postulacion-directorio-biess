package http

import (
	"github.com/gin-gonic/gin"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/http/handler"
)

// EvaluacionAdminModulo registra el banco de preguntas y la configuración de la evaluación
// (EF-14) bajo /api/v1/admin/evaluacion.
//
// Todas las rutas van en el grupo de escritura del panel, es decir, solo para el
// administrador, también las de lectura: el listado trae la respuesta correcta de cada
// pregunta y un analista no debe verla.
func EvaluacionAdminModulo(h *handler.EvaluacionAdminHandler) ModuloAdmin {
	return func(_, escritura *gin.RouterGroup) {
		evaluacion := escritura.Group("/evaluacion")
		evaluacion.POST("/preguntas", h.CrearPreguntas)
		evaluacion.GET("/preguntas", h.ListarPreguntas)
		evaluacion.PUT("/preguntas/:id", h.ActualizarPregunta)
		evaluacion.DELETE("/preguntas/:id", h.EliminarPregunta)
		evaluacion.GET("/config", h.ObtenerConfig)
		evaluacion.PUT("/config", h.GuardarConfig)
	}
}
