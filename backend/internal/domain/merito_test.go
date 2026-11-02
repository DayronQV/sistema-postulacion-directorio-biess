package domain

import "testing"

func TestEstadoVisible(t *testing.T) {
	casos := []struct {
		estado string
		supero bool
		want   string
	}{
		{EstadoBorrador, false, "Borrador"},
		{EstadoEnviada, false, "Enviada"},
		{EstadoEnviada, true, "Evaluación Técnica"},
		{EstadoCerrada, false, "Cerrada"},
		// La marca solo cuenta en una postulación enviada.
		{EstadoBorrador, true, "Borrador"},
		{EstadoCerrada, true, "Cerrada"},
	}
	for _, c := range casos {
		if got := EstadoVisible(c.estado, c.supero); got != c.want {
			t.Errorf("EstadoVisible(%q, %v) = %q; se esperaba %q", c.estado, c.supero, got, c.want)
		}
	}
}
