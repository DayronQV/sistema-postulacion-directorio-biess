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

func TestLimpiarReferencias_QuitaEspaciosAntesDeValidar(t *testing.T) {
	sucia := []domain.ReferenciaPersonal{{
		NombresApellidos: "  Ana Pérez ", Relacion: " Amiga ", Telefono: " 0999999999 ", Correo: " ana@example.com ",
	}}
	limpias := limpiarReferencias(sucia)
	want := domain.ReferenciaPersonal{NombresApellidos: "Ana Pérez", Relacion: "Amiga", Telefono: "0999999999", Correo: "ana@example.com"}
	if limpias[0] != want {
		t.Fatalf("limpiarReferencias = %+v; se esperaba %+v", limpias[0], want)
	}
	if err := validarFormulario(domain.Formulario{Referencias: limpias}); err != nil {
		t.Fatalf("la referencia limpia debe ser válida: %v", err)
	}
	if sucia[0].NombresApellidos != "  Ana Pérez " {
		t.Fatal("limpiarReferencias no debe modificar la lista original")
	}
}

func TestValidarFormulario_ReferenciaConTextoDemasiadoLargo(t *testing.T) {
	largo := make([]rune, domain.MaxLongitudNombreReferencia+1)
	for i := range largo {
		largo[i] = 'á'
	}
	r := domain.ReferenciaPersonal{NombresApellidos: string(largo), Relacion: "Amiga", Telefono: "0999999999", Correo: "ana@example.com"}
	err := validarFormulario(domain.Formulario{Referencias: []domain.ReferenciaPersonal{r}})
	var de *domain.Error
	if !errors.As(err, &de) || de.Fields["referencias[0].nombres_apellidos"] == "" {
		t.Fatalf("error = %#v; se esperaba el límite de caracteres", err)
	}
}

func TestValidarFormulario_MaximoTresReferencias(t *testing.T) {
	r := domain.ReferenciaPersonal{NombresApellidos: "Ana", Relacion: "Amiga", Telefono: "0999999999", Correo: "ana@example.com"}
	err := validarFormulario(domain.Formulario{Referencias: []domain.ReferenciaPersonal{r, r, r, r}})
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("error = %v; se esperaba validación", err)
	}
}
