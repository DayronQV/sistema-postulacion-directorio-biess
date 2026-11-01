package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/domain"
	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/service"
)

type documentosFake struct {
	guardado domain.Documento
	limite   int64
	err      error
}

func (f *documentosFake) Guardar(_ context.Context, _ string, d domain.Documento, limite int64) (domain.Documento, error) {
	f.guardado, f.limite = d, limite
	d.ID = "doc-1"
	return d, f.err
}
func (f *documentosFake) ObtenerPropio(context.Context, string, string) (domain.Documento, error) {
	return domain.Documento{}, f.err
}
func (f *documentosFake) EliminarPropioNoReferenciado(context.Context, string, string) error {
	return f.err
}

func TestDocumentoService_LimitesPorTipo(t *testing.T) {
	contenido := make([]byte, 3*1024*1024)
	copy(contenido, "%PDF-1.7")

	t.Run("tres MB pasan para organizacion", func(t *testing.T) {
		repo := &documentosFake{}
		d, err := service.NewDocumentoService(repo).Subir(context.Background(), "postulante", "organizacion", "respaldo.pdf", contenido)
		if err != nil || d.ID != "doc-1" || repo.limite != 50*1024*1024 {
			t.Fatalf("subida = %+v, %v; límite total = %d", d, err, repo.limite)
		}
	})

	t.Run("tres MB fallan para formacion", func(t *testing.T) {
		_, err := service.NewDocumentoService(&documentosFake{}).Subir(context.Background(), "postulante", "formacion", "titulo.pdf", contenido)
		if !errors.Is(err, domain.ErrUnprocessable) {
			t.Fatalf("error = %v; se esperaba 422", err)
		}
	})
}

func TestDocumentoService_RechazaArchivoRenombrado(t *testing.T) {
	_, err := service.NewDocumentoService(&documentosFake{}).Subir(
		context.Background(), "postulante", "formacion", "imagen.pdf", []byte("contenido de una imagen"),
	)
	if !errors.Is(err, domain.ErrUnprocessable) {
		t.Fatalf("error = %v; se esperaba 422", err)
	}
}

func TestDocumentoService_RechazaTipoDesconocidoYArchivoVacio(t *testing.T) {
	svc := service.NewDocumentoService(&documentosFake{})
	if _, err := svc.Subir(context.Background(), "p", "otro", "x.pdf", []byte("%PDF")); !errors.Is(err, domain.ErrUnprocessable) {
		t.Fatalf("tipo desconocido = %v", err)
	}
	if _, err := svc.Subir(context.Background(), "p", "formacion", "x.pdf", nil); !errors.Is(err, domain.ErrUnprocessable) {
		t.Fatalf("archivo vacío = %v", err)
	}
}
