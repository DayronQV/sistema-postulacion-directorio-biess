package postgres_test

import (
	"context"
	"testing"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/domain"
	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/integration/registrocivil"
	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/repository/postgres"
	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/service"
)

func TestPublicaciones_GuardaYNoEliminaLasFilas(t *testing.T) {
	pool := newPool(t)
	ctx := context.Background()
	postulanteID := crearPostulante(t, pool, registrocivil.CedulaFlujoFeliz)
	svc := newService(pool, postgres.NewPublicacionesRepo(pool))
	if _, err := svc.Obtener(ctx, postulanteID); err != nil {
		t.Fatal(err)
	}
	documentos := service.NewDocumentoService(postgres.NewDocumentoRepo(pool))
	doc, err := documentos.Subir(ctx, postulanteID, domain.DocumentoPublicacion, "publicacion.pdf", []byte("%PDF-1.7"))
	if err != nil {
		t.Fatal(err)
	}
	fila := domain.PublicacionAcademica{Titulo: "Mercado de valores", Tipo: domain.TipoPublicacion, Medio: "Revista", Fecha: "2024-01-15", Relacionada: true, DocumentoID: doc.ID}
	p, err := svc.Guardar(ctx, postulanteID, domain.Formulario{TienePublicaciones: true, Publicaciones: []domain.PublicacionAcademica{fila}})
	if err != nil {
		t.Fatal(err)
	}
	if !p.Formulario.TienePublicaciones || len(p.Formulario.Publicaciones) != 1 {
		t.Fatalf("formulario = %+v", p.Formulario)
	}

	p, err = svc.Guardar(ctx, postulanteID, domain.Formulario{TienePublicaciones: false})
	if err != nil {
		t.Fatal(err)
	}
	if p.Formulario.TienePublicaciones || len(p.Formulario.Publicaciones) != 0 {
		t.Fatalf("formulario = %+v", p.Formulario)
	}
}
