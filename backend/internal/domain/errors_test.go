package domain_test

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/domain"
)

func TestErrorsWrapTheirBase(t *testing.T) {
	cases := []struct {
		name string
		err  error
		base error
	}{
		{"validation", domain.Validation("x", nil), domain.ErrValidation},
		{"unprocessable", domain.Unprocessable("x", nil), domain.ErrUnprocessable},
		{"unauthorized", domain.Unauthorized("x"), domain.ErrUnauthorized},
		{"forbidden", domain.Forbidden("x"), domain.ErrForbidden},
		{"not found", domain.NotFound("x"), domain.ErrNotFound},
		{"conflict", domain.Conflict("x", nil), domain.ErrConflict},
		{"locked", domain.Locked("x", time.Minute), domain.ErrLocked},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if !errors.Is(c.err, c.base) {
				t.Fatalf("errors.Is(%v, %v) = false", c.err, c.base)
			}
			// También debe funcionar si otra capa lo envuelve con %w.
			if !errors.Is(fmt.Errorf("capa: %w", c.err), c.base) {
				t.Fatal("el error base se pierde al envolverlo con %w")
			}
		})
	}
}

func TestErrorsDoNotMatchOtherBases(t *testing.T) {
	if errors.Is(domain.NotFound("x"), domain.ErrConflict) {
		t.Fatal("NotFound no debe coincidir con ErrConflict")
	}
}

func TestErrorCarriesDetails(t *testing.T) {
	err := domain.Validation("Datos inválidos", map[string]string{"cedula": "Cédula no válida"})
	var de *domain.Error
	if !errors.As(err, &de) {
		t.Fatal("errors.As debe obtener *domain.Error")
	}
	if de.Message != "Datos inválidos" || de.Fields["cedula"] != "Cédula no válida" {
		t.Fatalf("detalles incorrectos: %+v", de)
	}

	locked := domain.Locked("bloqueado", 10*time.Minute)
	errors.As(locked, &de)
	if de.RetryAfter != 10*time.Minute {
		t.Fatalf("RetryAfter = %v", de.RetryAfter)
	}
}
