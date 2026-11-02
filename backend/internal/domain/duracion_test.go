package domain

import (
	"testing"
	"time"
)

func fecha(t *testing.T, valor string) time.Time {
	t.Helper()
	fecha, err := time.Parse(time.DateOnly, valor)
	if err != nil {
		t.Fatal(err)
	}
	return fecha
}

func TestCalcularDuracion(t *testing.T) {
	casos := []struct {
		nombre      string
		inicio, fin string
		anios       int
		meses       int
		dias        int
	}{
		{"mismo día", "2024-06-15", "2024-06-15", 0, 0, 0},
		{"días del mismo mes", "2024-06-01", "2024-06-16", 0, 0, 15},
		{"fin de mes", "2024-01-31", "2024-02-29", 0, 0, 29},
		{"año bisiesto", "2020-02-29", "2021-02-28", 0, 11, 30},
		{"años meses y días", "2018-01-10", "2020-03-15", 2, 2, 5},
		{"cambio de año", "2023-12-20", "2024-01-05", 0, 0, 16},
		{"fin anterior", "2024-06-02", "2024-06-01", 0, 0, 0},
	}
	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			a, m, d := CalcularDuracion(fecha(t, caso.inicio), fecha(t, caso.fin))
			if a != caso.anios || m != caso.meses || d != caso.dias {
				t.Fatalf("resultado = %d/%d/%d; esperado %d/%d/%d", a, m, d, caso.anios, caso.meses, caso.dias)
			}
		})
	}
}

func TestSumarDuracionesNormalizaElTotal(t *testing.T) {
	total := SumarDuraciones([]Duracion{{Anios: 1, Meses: 11, Dias: 20}, {Meses: 2, Dias: 15}})
	if total != (Duracion{Anios: 2, Meses: 2, Dias: 5}) {
		t.Fatalf("total inesperado: %+v", total)
	}
}
