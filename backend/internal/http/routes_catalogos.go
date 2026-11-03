package http

import (
	"github.com/gin-gonic/gin"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/http/handler"
)

// CatalogoRoutes registra las rutas públicas de catálogos bajo /api/v1/catalogos.
func CatalogoRoutes(h *handler.CatalogoHandler) RouteRegistrar {
	return func(api *gin.RouterGroup) {
		catalogos := api.Group("/catalogos")
		catalogos.GET("/paises", h.Paises)
		catalogos.GET("/paises/:id/provincias", h.Provincias)
		catalogos.GET("/provincias/:id/cantones", h.Cantones)
	}
}
