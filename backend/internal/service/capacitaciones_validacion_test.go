package service

import (
	"testing"
	"time"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/domain"
)

func capacitacionValida() domain.Capacitacion {
	return domain.Capacitacion{Tipo: "curso", Nombre: "Finanzas", Institucion: "Universidad", FechaInicio: "2024-01-01", FechaFin: "2024-01-05", Horas: 40}
}

func TestValidarCapacitacionesAceptaFilaValidaSinPDF(t *testing.T) {
	campos := map[string]string{}
	validarCapacitaciones([]domain.Capacitacion{capacitacionValida()}, time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC), campos)
	if len(campos) != 0 {
		t.Fatalf("errores = %v", campos)
	}
}

func TestValidarCapacitacionesRechazaFechasYHoras(t *testing.T) {
	fila := capacitacionValida()
	fila.FechaInicio = "2026-01-01"
	fila.FechaFin = "2023-01-01"
	fila.Horas = 0
	campos := map[string]string{}
	validarCapacitaciones([]domain.Capacitacion{fila}, time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC), campos)
	for _, campo := range []string{"capacitaciones[0].fecha_inicio", "capacitaciones[0].fecha_fin", "capacitaciones[0].horas"} {
		if campos[campo] == "" {
			t.Errorf("faltó %s: %v", campo, campos)
		}
	}
}
