package service

import (
	"strings"
	"testing"
	"time"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/domain"
)

func experienciaValida() domain.ExperienciaProfesional {
	return domain.ExperienciaProfesional{
		Tipo: "general", Cargo: "Analista", Institucion: "Empresa",
		FechaInicio: "2020-01-01", FechaFin: "2022-02-03", Actividades: "finanzas",
		Descripcion: "Análisis financiero", DocumentoID: "doc-1",
	}
}

func TestValidarExperienciaAceptaUnaFilaValida(t *testing.T) {
	campos := map[string]string{}
	validarExperiencia([]domain.ExperienciaProfesional{experienciaValida()}, time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), campos)
	if len(campos) != 0 {
		t.Fatalf("errores inesperados: %v", campos)
	}
}

func TestValidarExperienciaRechazaFechasDescripcionYDocumento(t *testing.T) {
	fila := experienciaValida()
	fila.FechaInicio = "2025-01-01"
	fila.FechaFin = "2019-01-01"
	fila.Descripcion = strings.Repeat("a", 1001)
	fila.DocumentoID = ""
	campos := map[string]string{}
	validarExperiencia([]domain.ExperienciaProfesional{fila}, time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), campos)
	for _, campo := range []string{"experiencia[0].fecha_inicio", "experiencia[0].fecha_fin", "experiencia[0].descripcion", "experiencia[0].documento_id"} {
		if campos[campo] == "" {
			t.Errorf("faltó validar %s: %v", campo, campos)
		}
	}
}

func TestValidarExperienciaRechazaDocumentoRepetido(t *testing.T) {
	campos := map[string]string{}
	validarExperiencia([]domain.ExperienciaProfesional{experienciaValida(), experienciaValida()}, time.Now(), campos)
	if campos["experiencia[1].documento_id"] == "" {
		t.Fatalf("errores = %v", campos)
	}
}
