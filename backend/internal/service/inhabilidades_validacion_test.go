package service

import (
	"errors"
	"strings"
	"testing"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/domain"
)

func conRemovido(d domain.RemovidoDetalle) domain.Inhabilidades {
	return domain.Inhabilidades{Respuestas: map[int]bool{domain.PreguntaRemovido: true}, Removido: d}
}

func TestValidarFormulario_InhabilidadesSinRemovidoNoExigeDetalle(t *testing.T) {
	f := domain.Formulario{Inhabilidades: domain.Inhabilidades{Respuestas: map[int]bool{1: true, 7: false}}}
	if err := validarFormulario(f); err != nil {
		t.Fatalf("borrador a medias: %v", err)
	}
}

func TestValidarFormulario_RemovidoExigeInstitucionYMotivo(t *testing.T) {
	casos := []struct {
		nombre  string
		detalle domain.RemovidoDetalle
		campo   string
	}{
		{"sin institución", domain.RemovidoDetalle{Motivo: "Reestructuración"}, "inhabilidades.removido_detalle.institucion"},
		{"sin motivo", domain.RemovidoDetalle{Institucion: "SRI"}, "inhabilidades.removido_detalle.motivo"},
		{"fecha mal escrita", domain.RemovidoDetalle{Institucion: "SRI", Motivo: "x", FechaFin: "31/01/2024"}, "inhabilidades.removido_detalle.fecha_fin"},
		{"motivo demasiado largo", domain.RemovidoDetalle{Institucion: "SRI", Motivo: strings.Repeat("a", domain.MaxLongitudMotivoRemovido+1)}, "inhabilidades.removido_detalle.motivo"},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			err := validarFormulario(domain.Formulario{Inhabilidades: conRemovido(c.detalle)})
			var de *domain.Error
			if !errors.As(err, &de) || de.Fields[c.campo] == "" {
				t.Fatalf("error = %#v; falta %s", err, c.campo)
			}
		})
	}
}

func TestValidarFormulario_RemovidoCompletoEsValido(t *testing.T) {
	d := domain.RemovidoDetalle{Institucion: "SRI", Cargo: "Analista", FechaFin: "2024-01-31", Motivo: "Reestructuración"}
	if err := validarFormulario(domain.Formulario{Inhabilidades: conRemovido(d)}); err != nil {
		t.Fatalf("detalle completo: %v", err)
	}
}

func TestLimpiarInhabilidades(t *testing.T) {
	sucio := conRemovido(domain.RemovidoDetalle{Institucion: "  SRI ", Cargo: " Analista ", FechaFin: " 2024-01-31 ", Motivo: " Fin "})
	limpio := limpiarInhabilidades(sucio)
	want := domain.RemovidoDetalle{Institucion: "SRI", Cargo: "Analista", FechaFin: "2024-01-31", Motivo: "Fin"}
	if limpio.Removido != want {
		t.Fatalf("detalle = %+v; se esperaba %+v", limpio.Removido, want)
	}

	suelto := limpiarInhabilidades(domain.Inhabilidades{
		Respuestas: map[int]bool{domain.PreguntaRemovido: false},
		Removido:   domain.RemovidoDetalle{Institucion: "SRI", Motivo: "x"},
	})
	if suelto.Removido != (domain.RemovidoDetalle{}) {
		t.Fatalf("el detalle de una pregunta 7 en No debe descartarse: %+v", suelto.Removido)
	}
}
