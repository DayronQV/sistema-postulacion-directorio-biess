package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// CORS permite que el frontend (en otro origen) llame a la API. Solo se aceptan los
// orígenes configurados. La autenticación va en la cabecera Authorization, por lo
// que no se habilitan cookies ni credenciales.
func CORS(allowedOrigins []string) gin.HandlerFunc {
	allowed := make(map[string]bool, len(allowedOrigins))
	for _, o := range allowedOrigins {
		allowed[o] = true
	}

	const (
		methods = "GET, POST, PUT, PATCH, DELETE, OPTIONS"
		headers = "Authorization, Content-Type, X-Request-ID"
		expose  = "X-Request-ID, Retry-After, Content-Disposition"
	)

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin == "" {
			c.Next()
			return
		}
		c.Writer.Header().Add("Vary", "Origin")

		if allowed[origin] {
			h := c.Writer.Header()
			h.Set("Access-Control-Allow-Origin", origin)
			h.Set("Access-Control-Expose-Headers", expose)
			if isPreflight(c.Request) {
				h.Set("Access-Control-Allow-Methods", methods)
				h.Set("Access-Control-Allow-Headers", headers)
				h.Set("Access-Control-Max-Age", "600")
			}
		}
		if isPreflight(c.Request) {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

func isPreflight(r *http.Request) bool {
	return r.Method == http.MethodOptions && strings.TrimSpace(r.Header.Get("Access-Control-Request-Method")) != ""
}
