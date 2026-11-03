package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/domain"
	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/service"
)

type documentosFake struct {
	guardado    domain.Documento
	limite      int64
	err         error
	guardados   int
	eliminados  int
	descargados int
}

func (f *documentosFake) Guardar(_ context.Context, _ string, d domain.Documento, limite int64) (domain.Documento, error) {
	f.guardado, f.limite = d, limite
	f.guardados++
	d.ID = "doc-1"
	return d, f.err
}
func (f *documentosFake) ObtenerPropio(context.Context, string, string) (domain.Documento, error) {
	f.descargados++
	return domain.Documento{}, f.err
}
func (f *documentosFake) EliminarPropioNoReferenciado(context.Context, string, string) error {
	f.eliminados++
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

// Con el periodo de postulación cerrado no se suben ni se eliminan documentos, igual que no se
// guarda ni se envía el formulario.
func TestDocumentoService_RespetaElPeriodoDePostulacion(t *testing.T) {
	ctx := context.Background()
	ahora := time.Date(2026, 11, 20, 12, 0, 0, 0, time.UTC)
	antes, despues := ahora.Add(-time.Hour), ahora.Add(time.Hour)
	pdf := []byte("%PDF-1.4 contenido")

	cerrados := map[string]func(*domain.Fase){
		"sin publicar":        func(f *domain.Fase) { f.Publicada = false },
		"sin habilitar":       func(f *domain.Fase) { f.HabilitaFormulario = false },
		"todavía no empieza":  func(f *domain.Fase) { f.FechaInicio = &despues },
		"el periodo ya cerró": func(f *domain.Fase) { f.FechaFin = &antes },
	}
	for nombre, cierra := range cerrados {
		t.Run(nombre, func(t *testing.T) {
			fases := nuevasFases()
			cierra(&fases.fases[0])
			docs := &documentosFake{}
			svc := service.NewDocumentoService(docs).ConFases(fases).ConReloj(func() time.Time { return ahora })

			_, err := svc.Subir(ctx, "p", "formacion", "x.pdf", pdf)
			if !errors.Is(err, domain.ErrForbidden) || err.Error() != domain.MsgPeriodoCerrado {
				t.Fatalf("subir: se esperaba 403 con el periodo cerrado, llegó %v", err)
			}
			err = svc.Eliminar(ctx, "p", "doc-1")
			if !errors.Is(err, domain.ErrForbidden) || err.Error() != domain.MsgPeriodoCerrado {
				t.Fatalf("eliminar: se esperaba 403 con el periodo cerrado, llegó %v", err)
			}
			if docs.guardados != 0 || docs.eliminados != 0 {
				t.Fatalf("no debe llegar al repositorio: guardados = %d, eliminados = %d", docs.guardados, docs.eliminados)
			}
			// Leer un documento propio sigue permitido.
			if _, err := svc.Descargar(ctx, "p", "doc-1"); err != nil || docs.descargados != 1 {
				t.Fatalf("descargar debe seguir funcionando: %v", err)
			}
		})
	}

	t.Run("con el periodo abierto sube y elimina", func(t *testing.T) {
		fases := nuevasFases()
		fases.fases[0].FechaInicio, fases.fases[0].FechaFin = &antes, &despues
		docs := &documentosFake{}
		svc := service.NewDocumentoService(docs).ConFases(fases).ConReloj(func() time.Time { return ahora })
		if _, err := svc.Subir(ctx, "p", "formacion", "x.pdf", pdf); err != nil {
			t.Fatalf("subir con el periodo abierto: %v", err)
		}
		if err := svc.Eliminar(ctx, "p", "doc-1"); err != nil {
			t.Fatalf("eliminar con el periodo abierto: %v", err)
		}
		if docs.guardados != 1 || docs.eliminados != 1 {
			t.Fatalf("guardados = %d, eliminados = %d", docs.guardados, docs.eliminados)
		}
	})

	t.Run("sin fases configuradas no restringe", func(t *testing.T) {
		docs := &documentosFake{}
		svc := service.NewDocumentoService(docs)
		if _, err := svc.Subir(ctx, "p", "formacion", "x.pdf", pdf); err != nil {
			t.Fatalf("sin fases no debe restringir: %v", err)
		}
	})

	t.Run("el periodo se revisa antes que el archivo", func(t *testing.T) {
		fases := nuevasFases()
		fases.fases[0].Publicada = false
		svc := service.NewDocumentoService(&documentosFake{}).ConFases(fases)
		// Un archivo inválido con el periodo cerrado da 403, no 422.
		if _, err := svc.Subir(ctx, "p", "otro", "x.pdf", nil); !errors.Is(err, domain.ErrForbidden) {
			t.Fatalf("se esperaba 403, llegó %v", err)
		}
	})
}
