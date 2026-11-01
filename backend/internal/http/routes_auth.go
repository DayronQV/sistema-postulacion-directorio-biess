package http

import (
	"github.com/gin-gonic/gin"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/http/handler"
)

// AuthRoutes registra las rutas públicas de autenticación bajo /api/v1/auth.
func AuthRoutes(h *handler.AuthHandler) RouteRegistrar {
	return func(api *gin.RouterGroup) {
		auth := api.Group("/auth")
		auth.POST("/identidad", h.Identidad)
		auth.POST("/registro", h.Registro)
	}
}
