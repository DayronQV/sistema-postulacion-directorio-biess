package handler

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/domain"
	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/http/middleware"
)

// DashboardUseCase es lo que el handler necesita para consultar los indicadores.
type DashboardUseCase interface {
	Obtener(ctx context.Context) (domain.DashboardPostulaciones, error)
}

// AdminDashboardHandler atiende el dashboard del panel.
type AdminDashboardHandler struct {
	dashboard DashboardUseCase
}

// NewAdminDashboardHandler crea el handler.
func NewAdminDashboardHandler(dashboard DashboardUseCase) *AdminDashboardHandler {
	return &AdminDashboardHandler{dashboard: dashboard}
}

type dashboardResponse struct {
	PorCargo struct {
		Afiliados int `json:"afiliados"`
		Jubilados int `json:"jubilados"`
		Empleador int `json:"empleador"`
	} `json:"por_cargo"`
	PorEstado struct {
		Borrador int `json:"borrador"`
		Enviada  int `json:"enviada"`
		Cerrada  int `json:"cerrada"`
	} `json:"por_estado"`
	Total int `json:"total"`
}

// Obtener atiende GET /admin/dashboard.
func (h *AdminDashboardHandler) Obtener(c *gin.Context) {
	d, err := h.dashboard.Obtener(c.Request.Context())
	if err != nil {
		middleware.Fail(c, err)
		return
	}
	var response dashboardResponse
	response.PorCargo.Afiliados = d.PorCargo.Afiliados
	response.PorCargo.Jubilados = d.PorCargo.Jubilados
	response.PorCargo.Empleador = d.PorCargo.Empleador
	response.PorEstado.Borrador = d.PorEstado.Borrador
	response.PorEstado.Enviada = d.PorEstado.Enviada
	response.PorEstado.Cerrada = d.PorEstado.Cerrada
	response.Total = d.Total
	c.JSON(http.StatusOK, response)
}

// Rutas registra la ruta de lectura para analistas y administradores.
func (h *AdminDashboardHandler) Rutas(lectura, _ *gin.RouterGroup) {
	lectura.GET("/dashboard", h.Obtener)
}
