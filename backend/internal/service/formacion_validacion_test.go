package service

import (
	"errors"
	"strings"
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

func TestValidarFormacion_RechazaListasYFechaInvalidas(t *testing.T) {
	fila := formacionValida()
	fila.Nivel = "bachiller"
	fila.Tipo = "publica"
	fila.Area = "medicina"
	fila.SenescytFecha = "2026-11-02"
	err := validarFormularioEn(domain.Formulario{Formacion: []domain.FormacionAcademica{fila}}, time.Date(2026, 11, 1, 23, 0, 0, 0, time.Local))
	var de *domain.Error
	if !errors.As(err, &de) {
		t.Fatalf("error = %v", err)
	}
	for _, campo := range []string{"nivel", "tipo", "area", "senescyt_fecha"} {
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

var hoyFormacion = time.Date(2026, 11, 1, 12, 0, 0, 0, time.Local)

func camposConError(t *testing.T, err error) map[string]string {
	t.Helper()
	var de *domain.Error
	if !errors.As(err, &de) {
		t.Fatalf("error = %v; se esperaba validación", err)
	}
	return de.Fields
}

func TestValidarFormacion_ElRespaldoEsOpcional(t *testing.T) {
	fila := formacionValida()
	fila.DocumentoID = ""
	if err := validarFormularioEn(domain.Formulario{Formacion: []domain.FormacionAcademica{fila}}, hoyFormacion); err != nil {
		t.Fatalf("sin respaldo debe ser válida (documento_id admite NULL en el esquema): %v", err)
	}
}

func TestValidarFormacion_RechazaElMismoDocumentoEnDosFilas(t *testing.T) {
	a, b := formacionValida(), formacionValida()
	b.Titulo = "Economista"
	fields := camposConError(t, validarFormularioEn(domain.Formulario{Formacion: []domain.FormacionAcademica{a, b}}, hoyFormacion))
	if fields["formacion[1].documento_id"] == "" || fields["formacion[0].documento_id"] != "" {
		t.Fatalf("se esperaba el error solo en la segunda fila: %v", fields)
	}
}

func TestValidarFormacion_LimitesDeFilasYDeTexto(t *testing.T) {
	filas := make([]domain.FormacionAcademica, domain.MaxFormaciones+1)
	for i := range filas {
		filas[i] = formacionValida()
		filas[i].DocumentoID = ""
	}
	if fields := camposConError(t, validarFormularioEn(domain.Formulario{Formacion: filas}, hoyFormacion)); fields["formacion"] == "" {
		t.Fatalf("se esperaba el máximo de filas: %v", fields)
	}

	larga := formacionValida()
	larga.Titulo = strings.Repeat("á", domain.MaxLongitudTituloFormacion+1)
	if fields := camposConError(t, validarFormularioEn(domain.Formulario{Formacion: []domain.FormacionAcademica{larga}}, hoyFormacion)); fields["formacion[0].titulo"] == "" {
		t.Fatalf("se esperaba el límite de caracteres: %v", fields)
	}
}
