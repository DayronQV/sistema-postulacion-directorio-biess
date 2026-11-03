package domain

import (
	"testing"
	"time"
)

func TestFase_FormularioAbierto(t *testing.T) {
	ahora := time.Date(2026, 11, 2, 12, 0, 0, 0, time.UTC)
	antes, despues := ahora.Add(-time.Hour), ahora.Add(time.Hour)
	casos := []struct {
		nombre string
		fase   Fase
		want   bool
	}{
		{"publicada y habilitada sin fechas", Fase{Publicada: true, HabilitaFormulario: true}, true},
		{"sin publicar", Fase{HabilitaFormulario: true}, false},
		{"sin habilitar el formulario", Fase{Publicada: true}, false},
		{"dentro de las fechas", Fase{Publicada: true, HabilitaFormulario: true, FechaInicio: &antes, FechaFin: &despues}, true},
		{"antes del inicio", Fase{Publicada: true, HabilitaFormulario: true, FechaInicio: &despues}, false},
		{"después del fin", Fase{Publicada: true, HabilitaFormulario: true, FechaFin: &antes}, false},
		// Ambos extremos se incluyen.
		{"justo en el inicio", Fase{Publicada: true, HabilitaFormulario: true, FechaInicio: &ahora}, true},
		{"justo en el fin", Fase{Publicada: true, HabilitaFormulario: true, FechaFin: &ahora}, true},
	}
	for _, c := range casos {
		if got := c.fase.FormularioAbierto(ahora); got != c.want {
			t.Errorf("%s: FormularioAbierto = %v, se esperaba %v", c.nombre, got, c.want)
		}
	}
}
