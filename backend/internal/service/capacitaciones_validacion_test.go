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

func TestValidarCapacitacionesLimitaLasHoras(t *testing.T) {
	fila := capacitacionValida()
	fila.Horas = domain.MaxHorasCapacitacion + 1
	campos := map[string]string{}
	validarCapacitaciones([]domain.Capacitacion{fila}, time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC), campos)
	if campos["capacitaciones[0].horas"] == "" {
		t.Fatalf("se esperaba el tope de horas: %v", campos)
	}

	fila.Horas = domain.MaxHorasCapacitacion
	campos = map[string]string{}
	validarCapacitaciones([]domain.Capacitacion{fila}, time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC), campos)
	if len(campos) != 0 {
		t.Fatalf("el tope exacto debe ser válido: %v", campos)
	}
}

func TestValidarCapacitacionesRechazaDocumentoRepetido(t *testing.T) {
	a, b := capacitacionValida(), capacitacionValida()
	a.DocumentoID, b.DocumentoID = "doc-1", "doc-1"
	campos := map[string]string{}
	validarCapacitaciones([]domain.Capacitacion{a, b}, time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC), campos)
	if campos["capacitaciones[1].documento_id"] == "" || campos["capacitaciones[0].documento_id"] != "" {
		t.Fatalf("se esperaba el error solo en la segunda fila: %v", campos)
	}
}
