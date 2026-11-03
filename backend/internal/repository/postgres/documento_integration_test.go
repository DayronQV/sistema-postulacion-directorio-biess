package postgres_test

import (
	"context"
	"errors"
	"testing"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/domain"
	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/integration/registrocivil"
	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/repository/postgres"
	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/service"
)

func TestDocumento_DescargaAisladaYEliminacion(t *testing.T) {
	pool := newPool(t)
	idA := crearPostulante(t, pool, registrocivil.CedulaFlujoFeliz)
	idB := crearPostulante(t, pool, registrocivil.CedulaUnionDeHecho)
	postulaciones := newService(pool)
	pA, err := postulaciones.Obtener(context.Background(), idA)
	if err != nil {
		t.Fatal(err)
	}
	pB, err := postulaciones.Obtener(context.Background(), idB)
	if err != nil {
		t.Fatal(err)
	}

	documentosRepo := postgres.NewDocumentoRepo(pool)
	documentos := service.NewDocumentoService(documentosRepo)
	d, err := documentos.Subir(context.Background(), idA, domain.DocumentoFormacion, "titulo.pdf", []byte("%PDF-1.7"))
	if err != nil {
		t.Fatalf("subiendo: %v", err)
	}
	if _, err = documentos.Descargar(context.Background(), idB, d.ID); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("otro postulante descargó el documento: %v", err)
	}
	if propio, err := documentos.Descargar(context.Background(), idA, d.ID); err != nil || string(propio.Contenido) != "%PDF-1.7" {
		t.Fatalf("descarga propia = %+v, %v", propio, err)
	}
	if _, err = documentosRepo.ObtenerDePostulacion(context.Background(), pB.ID, d.ID); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("el panel descargó un documento de otra postulación: %v", err)
	}
	if propio, err := documentosRepo.ObtenerDePostulacion(context.Background(), pA.ID, d.ID); err != nil || string(propio.Contenido) != "%PDF-1.7" {
		t.Fatalf("descarga del panel = %+v, %v", propio, err)
	}
	if err = documentos.Eliminar(context.Background(), idA, d.ID); err != nil {
		t.Fatalf("eliminando documento sin referencia: %v", err)
	}
	if _, err = documentos.Descargar(context.Background(), idA, d.ID); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("el documento eliminado sigue disponible: %v", err)
	}
}
