package service

import (
	"errors"
	"testing"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/domain"
)

func TestValidarFormulario_Referencias(t *testing.T) {
	valida := domain.ReferenciaPersonal{
		NombresApellidos: "Ana Pérez", Relacion: "Amiga", Telefono: "+593999999999", Correo: "ana@example.com",
	}
	if err := validarFormulario(domain.Formulario{Referencias: []domain.ReferenciaPersonal{valida}}); err != nil {
		t.Fatalf("referencia válida: %v", err)
	}

	casos := []struct {
		nombre     string
		referencia domain.ReferenciaPersonal
		campo      string
	}{
		{"nombre vacío", domain.ReferenciaPersonal{Relacion: "Amiga", Telefono: "0999999999", Correo: "ana@example.com"}, "referencias[0].nombres_apellidos"},
		{"relación vacía", domain.ReferenciaPersonal{NombresApellidos: "Ana", Telefono: "0999999999", Correo: "ana@example.com"}, "referencias[0].relacion"},
		{"teléfono inválido", domain.ReferenciaPersonal{NombresApellidos: "Ana", Relacion: "Amiga", Telefono: "abc", Correo: "ana@example.com"}, "referencias[0].telefono"},
		{"correo inválido", domain.ReferenciaPersonal{NombresApellidos: "Ana", Relacion: "Amiga", Telefono: "0999999999", Correo: "sin-arroba"}, "referencias[0].correo"},
	}
	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			err := validarFormulario(domain.Formulario{Referencias: []domain.ReferenciaPersonal{caso.referencia}})
			var de *domain.Error
			if !errors.As(err, &de) || de.Fields[caso.campo] == "" {
				t.Fatalf("error = %#v; falta %s", err, caso.campo)
			}
		})
	}
}

func TestValidarFormulario_MaximoTresReferencias(t *testing.T) {
	r := domain.ReferenciaPersonal{NombresApellidos: "Ana", Relacion: "Amiga", Telefono: "0999999999", Correo: "ana@example.com"}
	err := validarFormulario(domain.Formulario{Referencias: []domain.ReferenciaPersonal{r, r, r, r}})
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("error = %v; se esperaba validación", err)
	}
}
