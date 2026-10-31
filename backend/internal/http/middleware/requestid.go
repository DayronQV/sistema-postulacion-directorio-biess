package middleware

import (
	"crypto/rand"
	"encoding/hex"

	"github.com/gin-gonic/gin"
)

// RequestIDHeader es la cabecera con el identificador de la petición.
const RequestIDHeader = "X-Request-ID"

const requestIDKey = "request_id"

// RequestID asigna un identificador a cada petición (el que envía el cliente si es
// seguro, o uno nuevo) y lo devuelve en la cabecera de la respuesta. Sirve para
// relacionar un error del cliente con el registro del servidor.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader(RequestIDHeader)
		if !validRequestID(id) {
			id = newRequestID()
		}
		c.Set(requestIDKey, id)
		c.Header(RequestIDHeader, id)
		c.Next()
	}
}

// RequestIDFrom devuelve el identificador de la petición actual.
func RequestIDFrom(c *gin.Context) string {
	return c.GetString(requestIDKey)
}

// validRequestID acepta solo caracteres seguros para registros y cabeceras.
func validRequestID(id string) bool {
	if id == "" || len(id) > 64 {
		return false
	}
	for _, r := range id {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '-', r == '_', r == '.':
		default:
			return false
		}
	}
	return true
}

func newRequestID() string {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return "sin-id"
	}
	return hex.EncodeToString(b)
}
