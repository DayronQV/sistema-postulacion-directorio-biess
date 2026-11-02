package handler_test

import (
	"encoding/json"
	"testing"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/domain"
)

// Criterio de la issue #38: la marca del administrador queda reflejada para el postulante.
func TestPostulacionObtener_EstadoVisibleConElMerito(t *testing.T) {
	casos := []struct {
		estado string
		supero bool
		want   string
	}{
		{domain.EstadoBorrador, false, "Borrador"},
		{domain.EstadoEnviada, false, "Enviada"},
		{domain.EstadoEnviada, true, "Evaluación Técnica"},
	}
	for _, c := range casos {
		p := postulacionDeEjemplo()
		p.Estado, p.SuperoMerito = c.estado, c.supero
		w := send(newPostulacionRouter(&fakePostulacion{resultado: p}), "GET", "postulante", "")
		var got struct {
			SuperoMerito  bool   `json:"supero_merito"`
			EstadoVisible string `json:"estado_visible"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
			t.Fatal(err)
		}
		if got.EstadoVisible != c.want || got.SuperoMerito != c.supero {
			t.Fatalf("%s/%v: cuerpo = %s", c.estado, c.supero, w.Body.String())
		}
	}
}
