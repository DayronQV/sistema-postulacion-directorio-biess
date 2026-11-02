package service

import (
	"testing"
	"time"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/domain"
)

func publicacionValida() domain.PublicacionAcademica {
	return domain.PublicacionAcademica{Titulo: "Mercado de valores", Tipo: "publicacion", Medio: "Revista", Fecha: "2024-01-15", Relacionada: true, DocumentoID: "doc-1"}
}

func TestValidarPublicacionesSegunLaRespuesta(t *testing.T) {
	hoy := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	casos := []struct {
		nombre string
		form   domain.Formulario
		campo  string
	}{
		{"no sin filas", domain.Formulario{}, ""},
		{"no con filas", domain.Formulario{Publicaciones: []domain.PublicacionAcademica{publicacionValida()}}, "publicaciones"},
		{"sí sin filas", domain.Formulario{TienePublicaciones: true}, "publicaciones"},
		{"sí con fila", domain.Formulario{TienePublicaciones: true, Publicaciones: []domain.PublicacionAcademica{publicacionValida()}}, ""},
	}
	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			campos := map[string]string{}
			validarPublicaciones(caso.form, hoy, campos)
			if caso.campo == "" && len(campos) > 0 {
				t.Fatalf("errores = %v", campos)
			}
			if caso.campo != "" && campos[caso.campo] == "" {
				t.Fatalf("faltó %s: %v", caso.campo, campos)
			}
		})
	}
}

func TestValidarPublicacionesRechazaFechaFuturaYDocumentoFaltante(t *testing.T) {
	fila := publicacionValida()
	fila.Fecha = "2026-01-01"
	fila.DocumentoID = ""
	campos := map[string]string{}
	validarPublicaciones(domain.Formulario{TienePublicaciones: true, Publicaciones: []domain.PublicacionAcademica{fila}}, time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC), campos)
	for _, campo := range []string{"publicaciones[0].fecha", "publicaciones[0].documento_id"} {
		if campos[campo] == "" {
			t.Errorf("faltó %s: %v", campo, campos)
		}
	}
}
