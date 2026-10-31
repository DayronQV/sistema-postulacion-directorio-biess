package middleware

import (
	"io"
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
)

// Recovery captura los pánicos de los handlers, los registra con su traza y
// responde 500 con el formato de error de la API, sin exponer detalles.
func Recovery(logger *slog.Logger) gin.HandlerFunc {
	return gin.CustomRecoveryWithWriter(io.Discard, func(c *gin.Context, recovered any) {
		logger.Error("pánico recuperado",
			"panic", recovered,
			"stack", string(debug.Stack()),
			"request_id", RequestIDFrom(c),
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
		)
		c.AbortWithStatusJSON(http.StatusInternalServerError, ErrorBody{
			Error: ErrorPayload{Code: CodeInternal, Message: "Error interno del servidor"},
		})
	})
}
