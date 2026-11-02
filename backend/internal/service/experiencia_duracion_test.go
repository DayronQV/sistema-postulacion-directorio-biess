package service

import (
	"testing"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/domain"
)

func TestCompletarDuraciones_CadaExperienciaYElTotal(t *testing.T) {
	a, b := experienciaValida(), experienciaValida()
	a.FechaInicio, a.FechaFin = "2018-01-10", "2020-03-15" // 2 años, 2 meses, 5 días
	b.FechaInicio, b.FechaFin = "2021-01-01", "2021-07-28" // 6 meses, 27 días
	f := domain.Formulario{Experiencia: []domain.ExperienciaProfesional{a, b}}

	completarDuraciones(&f)

	if got := f.Experiencia[0].Duracion; got != (domain.Duracion{Anios: 2, Meses: 2, Dias: 5}) {
		t.Fatalf("duración 1 = %+v", got)
	}
	if got := f.Experiencia[1].Duracion; got != (domain.Duracion{Meses: 6, Dias: 27}) {
		t.Fatalf("duración 2 = %+v", got)
	}
	// 5 + 27 = 32 días → 1 mes y 2 días; 2 + 6 + 1 = 9 meses.
	if f.DuracionExperiencia != (domain.Duracion{Anios: 2, Meses: 9, Dias: 2}) {
		t.Fatalf("total = %+v", f.DuracionExperiencia)
	}
}

func TestCompletarDuraciones_SinExperienciaElTotalEsCero(t *testing.T) {
	f := domain.Formulario{DuracionExperiencia: domain.Duracion{Anios: 9}}
	completarDuraciones(&f)
	if f.DuracionExperiencia != (domain.Duracion{}) {
		t.Fatalf("total = %+v; se esperaba cero", f.DuracionExperiencia)
	}
}
