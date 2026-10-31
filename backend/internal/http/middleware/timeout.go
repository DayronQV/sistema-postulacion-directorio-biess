package middleware

import (
	"context"
	"errors"
	"time"

	"github.com/gin-gonic/gin"
)

// Timeout limita el tiempo de cada petición. Pone un plazo en el contexto de la
// petición: los repositorios y clientes que lo usan (pgx, go-redis) cancelan sus
// consultas cuando se agota. Si el plazo vence y nadie respondió, se responde 500.
//
// Importante: el plazo solo se aplica a lo que recibe ese contexto. Los handlers deben
// pasar siempre c.Request.Context() a los servicios, y estos a los repositorios (pgx,
// go-redis); si se usa context.Background() la consulta ignora el timeout y sigue
// corriendo aunque la petición ya haya vencido. El middleware solo puede responder
// cuando el handler vuelve, no interrumpe código que no mira el contexto.
func Timeout(d time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		if d <= 0 {
			c.Next()
			return
		}
		ctx, cancel := context.WithTimeout(c.Request.Context(), d)
		defer cancel()
		c.Request = c.Request.WithContext(ctx)

		c.Next()

		if errors.Is(ctx.Err(), context.DeadlineExceeded) && !c.Writer.Written() && len(c.Errors) == 0 {
			Fail(c, ctx.Err())
		}
	}
}
