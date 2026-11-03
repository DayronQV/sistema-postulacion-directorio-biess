package middleware_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/domain"
	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/http/middleware"
)

func TestTranslate(t *testing.T) {
	cases := []struct {
		name     string
		err      error
		status   int
		code     string
		message  string
		internal bool
	}{
		{"validation", domain.Validation("Datos inválidos", nil), 400, "VALIDATION_ERROR", "Datos inválidos", false},
		{"unprocessable", domain.Unprocessable("Nacionalidad no permitida", nil), 422, "VALIDATION_ERROR", "Nacionalidad no permitida", false},
		{"unauthorized", domain.Unauthorized("sus credenciales de ingreso son incorrectas"), 401, "UNAUTHORIZED", "sus credenciales de ingreso son incorrectas", false},
		{"forbidden", domain.Forbidden("sin permiso"), 403, "FORBIDDEN", "sin permiso", false},
		{"not found", domain.NotFound("no existe"), 404, "NOT_FOUND", "no existe", false},
		{"conflict", domain.Conflict("ya existe", nil), 409, "CONFLICT", "ya existe", false},
		{"locked", domain.Locked("bloqueado", time.Minute), 429, "LOCKED", "bloqueado", false},
		{"wrapped with %w", fmt.Errorf("servicio: %w", domain.NotFound("no existe")), 404, "NOT_FOUND", "no existe", false},
		{"bare sentinel uses default message", domain.ErrNotFound, 404, "NOT_FOUND", "Recurso no encontrado", false},
		{"client canceled is not an internal error", fmt.Errorf("consulta: %w", context.Canceled), 499, "INTERNAL_ERROR", "La solicitud fue cancelada", false},
		{"unknown error", errors.New("pq: connection refused"), 500, "INTERNAL_ERROR", "Error interno del servidor", true},
		{"deadline exceeded", fmt.Errorf("consulta: %w", context.DeadlineExceeded), 500, "INTERNAL_ERROR", "La solicitud tardó demasiado. Inténtelo nuevamente", true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			status, body, internal := middleware.Translate(c.err)
			if status != c.status || body.Error.Code != c.code || body.Error.Message != c.message || internal != c.internal {
				t.Fatalf("Translate = %d %+v internal=%v", status, body.Error, internal)
			}
		})
	}
}

func TestTranslateRoundsRetryAfterUp(t *testing.T) {
	_, body, _ := middleware.Translate(domain.Locked("x", 1500*time.Millisecond))
	if body.Error.RetryAfterSeconds != 2 {
		t.Fatalf("RetryAfterSeconds = %d, se esperaba 2", body.Error.RetryAfterSeconds)
	}
}
