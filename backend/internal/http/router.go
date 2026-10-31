// Package http arma el router de Gin: middlewares globales, /health y el grupo
// /api/v1 donde cada módulo registra sus propias rutas.
package http

import (
	"log/slog"

	"github.com/gin-gonic/gin"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/config"
	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/http/handler"
	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/http/middleware"
)

// RouteRegistrar registra las rutas de un módulo dentro del grupo /api/v1.
// Cada módulo define el suyo en su propio archivo (routes_auth.go,
// routes_postulacion.go, ...) para que nadie edite el mismo archivo de rutas.
type RouteRegistrar func(api *gin.RouterGroup)

// Deps son las dependencias que necesita el router. Las construye main.go.
type Deps struct {
	Config  config.Config
	Logger  *slog.Logger
	Health  *handler.HealthHandler
	Modules []RouteRegistrar
}

// NewRouter arma el motor de Gin con los middlewares globales y las rutas.
//
// Orden de los middlewares: el identificador de petición y el registro van primero
// para cubrir todo; la recuperación de pánicos y el traductor de errores envuelven a
// los handlers; CORS responde las preflight antes de llegar a ellos.
func NewRouter(d Deps) *gin.Engine {
	r := gin.New()
	// Sin proxies de confianza: la IP del cliente es la de la conexión directa.
	_ = r.SetTrustedProxies(nil)

	r.Use(
		middleware.RequestID(),
		middleware.RequestLogger(d.Logger),
		middleware.Recovery(d.Logger),
		middleware.ErrorHandler(d.Logger),
		middleware.CORS(d.Config.CORSAllowedOrigins),
		middleware.Timeout(d.Config.RequestTimeout),
	)
	r.NoRoute(middleware.NoRoute())

	r.GET("/health", d.Health.Get)

	api := r.Group("/api/v1")
	for _, register := range d.Modules {
		register(api)
	}
	return r
}
