package handler

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/http/middleware"
	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/service"
)

// MeritoUseCase es lo que el handler necesita para marcar el mérito de una postulación.
type MeritoUseCase interface {
	MarcarMerito(ctx context.Context, postulacionID string, supero bool) (service.ResultadoMerito, error)
}

// AdminPostulacionHandler atiende las acciones del panel sobre las postulaciones.
type AdminPostulacionHandler struct {
	meritos MeritoUseCase
}

// NewAdminPostulacionHandler crea el handler.
func NewAdminPostulacionHandler(meritos MeritoUseCase) *AdminPostulacionHandler {
	return &AdminPostulacionHandler{meritos: meritos}
}

type meritoRequest struct {
	// Puntero para exigir el campo: un false explícito es válido, ausente no.
	Supero *bool `json:"supero" binding:"required"`
}

type meritoResponse struct {
	ID            string `json:"id"`
	SuperoMerito  bool   `json:"supero_merito"`
	EstadoVisible string `json:"estado_visible"`
}

// MarcarMerito atiende PUT /admin/postulaciones/{id}/merito.
func (h *AdminPostulacionHandler) MarcarMerito(c *gin.Context) {
	var req meritoRequest
	if !middleware.BindJSON(c, &req) {
		return
	}
	r, err := h.meritos.MarcarMerito(c.Request.Context(), c.Param("id"), *req.Supero)
	if err != nil {
		middleware.Fail(c, err)
		return
	}
	c.JSON(http.StatusOK, meritoResponse{ID: r.ID, SuperoMerito: r.SuperoMerito, EstadoVisible: r.EstadoVisible})
}

// Rutas registra las rutas de escritura de este módulo en el panel (solo administrador).
func (h *AdminPostulacionHandler) Rutas(_, escritura *gin.RouterGroup) {
	escritura.PUT("/postulaciones/:id/merito", h.MarcarMerito)
}
