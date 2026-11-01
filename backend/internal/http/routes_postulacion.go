package http

import (
	"github.com/gin-gonic/gin"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/domain"
	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/http/handler"
	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/http/middleware"
)

// PostulacionRoutes registra las rutas de la postulación bajo /api/v1/postulacion.
// Exigen un token válido (401) con rol postulante (403).
func PostulacionRoutes(h *handler.PostulacionHandler, tokens domain.TokenIssuer) RouteRegistrar {
	return func(api *gin.RouterGroup) {
		postulacion := api.Group("/postulacion",
			middleware.RequireAuth(tokens),
			middleware.RequireRole(domain.RolPostulante),
		)
		postulacion.GET("", h.Obtener)
		postulacion.PUT("", h.Guardar)
	}
}
