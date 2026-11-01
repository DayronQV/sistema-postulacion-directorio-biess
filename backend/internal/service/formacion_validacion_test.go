package service

import (
	"errors"
	"testing"
	"time"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/domain"
)

func formacionValida() domain.FormacionAcademica {
	return domain.FormacionAcademica{
		Nivel: domain.NivelMaestria, Institucion: "Universidad de prueba",
		Tipo: domain.InstitucionNacional, Titulo: "Magíster en Finanzas",
		Area: domain.AreaFinanzas, SenescytNumero: "REG-123",
		SenescytFecha: "2025-10-15", DocumentoID: "doc-1",
	}
}

func TestValidarFormacion_AceptaFilaValidaYLimpiaTextos(t *testing.T) {
	fila := formacionValida()
	fila.Institucion = "  Universidad de prueba  "
	limpias := limpiarFormacion([]domain.FormacionAcademica{fila})
	if limpias[0].Institucion != "Universidad de prueba" {
		t.Fatalf("institución = %q", limpias[0].Institucion)
	}
	if err := validarFormularioEn(domain.Formulario{Formacion: limpias}, time.Date(2026, 11, 1, 12, 0, 0, 0, time.Local)); err != nil {
		t.Fatalf("formación válida: %v", err)
	}
}

func TestValidarFormacion_RechazaListasFechaYDocumentoInvalidos(t *testing.T) {
	fila := formacionValida()
	fila.Nivel = "bachiller"
	fila.Tipo = "publica"
	fila.Area = "medicina"
	fila.SenescytFecha = "2026-11-02"
	fila.DocumentoID = ""
	err := validarFormularioEn(domain.Formulario{Formacion: []domain.FormacionAcademica{fila}}, time.Date(2026, 11, 1, 23, 0, 0, 0, time.Local))
	var de *domain.Error
	if !errors.As(err, &de) {
		t.Fatalf("error = %v", err)
	}
	for _, campo := range []string{"nivel", "tipo", "area", "senescyt_fecha", "documento_id"} {
		clave := "formacion[0]." + campo
		if de.Fields[clave] == "" {
			t.Errorf("falta error para %s: %v", clave, de.Fields)
		}
	}
}

func TestValidarFormacion_RechazaFechaInexistente(t *testing.T) {
	fila := formacionValida()
	fila.SenescytFecha = "2026-02-30"
	err := validarFormularioEn(domain.Formulario{Formacion: []domain.FormacionAcademica{fila}}, time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC))
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("error = %v", err)
	}
}
