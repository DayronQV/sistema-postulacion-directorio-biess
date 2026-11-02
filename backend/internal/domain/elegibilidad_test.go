package domain

import "testing"

// respuestas arma las 13 respuestas con la 1 en Sí y el resto en No, y aplica los cambios.
func respuestasBase(cambios map[int]bool) Inhabilidades {
	r := map[int]bool{}
	for n := 1; n <= TotalPreguntasInhabilidades; n++ {
		r[n] = n == 1
	}
	for n, v := range cambios {
		r[n] = v
	}
	return Inhabilidades{Respuestas: r}
}

func TestInhabilidades_Elegible(t *testing.T) {
	casos := []struct {
		nombre string
		inhab  Inhabilidades
		want   bool
	}{
		{"1 en Sí y 2 a 13 en No", respuestasBase(nil), true},
		{"la 1 en No", respuestasBase(map[int]bool{1: false}), false},
		{"la 2 en Sí", respuestasBase(map[int]bool{2: true}), false},
		{"la 7 en Sí", respuestasBase(map[int]bool{7: true}), false},
		{"la 13 en Sí", respuestasBase(map[int]bool{13: true}), false},
		{"todas en Sí", func() Inhabilidades {
			c := map[int]bool{}
			for n := 2; n <= TotalPreguntasInhabilidades; n++ {
				c[n] = true
			}
			return respuestasBase(c)
		}(), false},
		{"sin respuestas", Inhabilidades{}, false},
		{"falta una pregunta", func() Inhabilidades {
			i := respuestasBase(nil)
			delete(i.Respuestas, 13)
			return i
		}(), false},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			if got := c.inhab.Elegible(); got != c.want {
				t.Fatalf("Elegible() = %v; se esperaba %v", got, c.want)
			}
		})
	}
}

func TestInhabilidades_Completa(t *testing.T) {
	if !respuestasBase(nil).Completa() {
		t.Fatal("13 respuestas deben estar completas")
	}
	if (Inhabilidades{Respuestas: map[int]bool{1: true}}).Completa() {
		t.Fatal("con una sola respuesta no está completa")
	}
}
