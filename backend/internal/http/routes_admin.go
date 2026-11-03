package http

import (
	"github.com/gin-gonic/gin"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/domain"
	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/http/handler"
	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/http/middleware"
)

// ModuloAdmin registra las rutas de un módulo del panel (fases, dashboard, nómina, ...)
// en dos grupos ya protegidos bajo /api/v1/admin:
//   - lectura: analista y administrador (el analista es de solo lectura);
//   - escritura: solo administrador (un analista recibe 403).
//
// Ejemplo: func(lectura, escritura *gin.RouterGroup) {
// lectura.GET("/fases", h.Listar); escritura.PUT("/fases/:id", h.Actualizar) }
type ModuloAdmin func(lectura, escritura *gin.RouterGroup)

// AdminRoutes registra el login del panel (público) y los módulos de administración.
// Sin token responde 401; con un token de postulante, 403.
func AdminRoutes(h *handler.AdminAuthHandler, tokens domain.TokenIssuer, modulos ...ModuloAdmin) RouteRegistrar {
	return func(api *gin.RouterGroup) {
		admin := api.Group("/admin")
		admin.POST("/auth/login", h.Login)

		lectura := admin.Group("",
			middleware.RequireAuth(tokens),
			middleware.RequireRole(domain.RolAnalista, domain.RolAdministrador),
		)
		escritura := lectura.Group("", middleware.RequireRole(domain.RolAdministrador))

		lectura.GET("/sesion", h.Sesion)
		for _, registrar := range modulos {
			registrar(lectura, escritura)
		}
	}
}
