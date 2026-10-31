package middleware

import (
	"context"
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

// RequestLogger registra cada petición al terminar: método, ruta, estado, duración
// e identificador. Las consultas a /health se registran en nivel debug para no
// llenar el log con los chequeos periódicos.
func RequestLogger(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		level := slog.LevelInfo
		if c.Request.URL.Path == "/health" {
			level = slog.LevelDebug
		}
		logger.Log(context.Background(), level, "petición",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
			"duration_ms", time.Since(start).Milliseconds(),
			"request_id", RequestIDFrom(c),
		)
	}
}
