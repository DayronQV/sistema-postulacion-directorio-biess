// Package middleware reúne los middlewares de Gin: identificador de petición,
// registro, recuperación de pánicos, traducción de errores, CORS y timeout.
package middleware

import (
	"context"
	"errors"
	"log/slog"
	"math"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/domain"
)

// Códigos de error de la API (docs/api.md).
const (
	CodeValidation   = "VALIDATION_ERROR"
	CodeUnauthorized = "UNAUTHORIZED"
	CodeForbidden    = "FORBIDDEN"
	CodeNotFound     = "NOT_FOUND"
	CodeConflict     = "CONFLICT"
	CodeLocked       = "LOCKED"
	CodeInternal     = "INTERNAL_ERROR"
)

// ErrorBody es el JSON de error único de la API:
// { "error": { code, message, fields?, retry_after_seconds? } }.
type ErrorBody struct {
	Error ErrorPayload `json:"error"`
}

// ErrorPayload es el contenido del campo "error".
type ErrorPayload struct {
	Code              string            `json:"code"`
	Message           string            `json:"message"`
	Fields            map[string]string `json:"fields,omitempty"`
	RetryAfterSeconds int               `json:"retry_after_seconds,omitempty"`
}

type kind struct {
	base    error
	status  int
	code    string
	message string // mensaje por defecto cuando el error no trae uno propio
}

var kinds = []kind{
	{domain.ErrValidation, http.StatusBadRequest, CodeValidation, "Datos inválidos"},
	{domain.ErrUnprocessable, http.StatusUnprocessableEntity, CodeValidation, "No se puede procesar la solicitud"},
	{domain.ErrUnauthorized, http.StatusUnauthorized, CodeUnauthorized, "No autorizado"},
	{domain.ErrForbidden, http.StatusForbidden, CodeForbidden, "No tiene permiso para realizar esta acción"},
	{domain.ErrNotFound, http.StatusNotFound, CodeNotFound, "Recurso no encontrado"},
	{domain.ErrConflict, http.StatusConflict, CodeConflict, "El recurso ya existe o su estado no permite la operación"},
	{domain.ErrLocked, http.StatusTooManyRequests, CodeLocked, "Demasiados intentos. Inténtelo nuevamente más tarde"},
}

// Translate convierte cualquier error en el estado HTTP y el cuerpo JSON de la API.
// internal es true cuando el error no es de dominio (se responde 500 con un mensaje
// genérico: nunca se exponen detalles internos al cliente).
func Translate(err error) (status int, body ErrorBody, internal bool) {
	for _, k := range kinds {
		if !errors.Is(err, k.base) {
			continue
		}
		payload := ErrorPayload{Code: k.code, Message: k.message}
		var de *domain.Error
		if errors.As(err, &de) {
			if de.Message != "" {
				payload.Message = de.Message
			}
			payload.Fields = de.Fields
			if de.RetryAfter > 0 {
				payload.RetryAfterSeconds = int(math.Ceil(de.RetryAfter.Seconds()))
			}
		}
		return k.status, ErrorBody{Error: payload}, false
	}

	message := "Error interno del servidor"
	if errors.Is(err, context.DeadlineExceeded) {
		message = "La solicitud tardó demasiado. Inténtelo nuevamente"
	}
	return http.StatusInternalServerError, ErrorBody{Error: ErrorPayload{Code: CodeInternal, Message: message}}, true
}

// Fail registra el error en el contexto y corta la cadena. Los handlers lo usan en
// lugar de escribir la respuesta ellos mismos: ErrorHandler la arma con el formato
// único de la API.
func Fail(c *gin.Context, err error) {
	_ = c.Error(err)
	c.Abort()
}

// ErrorHandler escribe la respuesta de error de la petición. Debe registrarse antes
// que los demás middlewares que puedan llamar a Fail.
func ErrorHandler(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) == 0 || c.Writer.Written() {
			return
		}
		err := c.Errors.Last().Err
		status, body, internal := Translate(err)
		if internal {
			logger.Error("error interno",
				"error", err,
				"request_id", RequestIDFrom(c),
				"method", c.Request.Method,
				"path", c.Request.URL.Path,
			)
		}
		if body.Error.RetryAfterSeconds > 0 {
			c.Header("Retry-After", strconv.Itoa(body.Error.RetryAfterSeconds))
		}
		c.AbortWithStatusJSON(status, body)
	}
}

// NoRoute responde 404 con el formato de error de la API cuando la ruta no existe.
func NoRoute() gin.HandlerFunc {
	return func(c *gin.Context) {
		Fail(c, domain.NotFound("Recurso no encontrado"))
	}
}
