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

func TestFormacion_SeGuardaRecargaYElimina(t *testing.T) {
	pool := newPool(t)
	postulanteID := crearPostulante(t, pool, registrocivil.CedulaFlujoFeliz)
	svc := newService(pool, postgres.NewFormacionRepo(pool))
	p, err := svc.Obtener(context.Background(), postulanteID)
	if err != nil {
		t.Fatal(err)
	}
	documentos := service.NewDocumentoService(postgres.NewDocumentoRepo(pool))
	doc, err := documentos.Subir(context.Background(), postulanteID, domain.DocumentoFormacion, "titulo.pdf", []byte("%PDF-1.7"))
	if err != nil {
		t.Fatal(err)
	}
	fila := domain.FormacionAcademica{
		Nivel: domain.NivelTercerNivel, Institucion: "UPSE", Tipo: domain.InstitucionNacional,
		Titulo: "Economista", Area: domain.AreaEconomia, SenescytNumero: "1001",
		SenescytFecha: "2020-01-31", DocumentoID: doc.ID,
	}
	guardada, err := svc.Guardar(context.Background(), postulanteID, domain.Formulario{Formacion: []domain.FormacionAcademica{fila}})
	if err != nil {
		t.Fatalf("guardando: %v", err)
	}
	if len(guardada.Formulario.Formacion) != 1 || guardada.Formulario.Formacion[0] != fila {
		t.Fatalf("formación recargada = %+v", guardada.Formulario.Formacion)
	}
	if _, err = svc.Guardar(context.Background(), postulanteID, domain.Formulario{}); err != nil {
		t.Fatalf("eliminando: %v", err)
	}
	var total int
	if err := pool.QueryRow(context.Background(), `SELECT count(*) FROM formacion_academica WHERE postulacion_id = $1`, p.ID).Scan(&total); err != nil {
		t.Fatal(err)
	}
	if total != 0 {
		t.Fatalf("quedaron %d filas después de eliminar", total)
	}
}

func TestFormacion_RechazaDocumentoAjenoODeOtroTipo(t *testing.T) {
	pool := newPool(t)
	idA := crearPostulante(t, pool, registrocivil.CedulaFlujoFeliz)
	idB := crearPostulante(t, pool, registrocivil.CedulaUnionDeHecho)
	svcA := newService(pool, postgres.NewFormacionRepo(pool))
	svcB := newService(pool)
	if _, err := svcA.Obtener(context.Background(), idA); err != nil {
		t.Fatal(err)
	}
	if _, err := svcB.Obtener(context.Background(), idB); err != nil {
		t.Fatal(err)
	}
	docs := service.NewDocumentoService(postgres.NewDocumentoRepo(pool))
	ajeno, err := docs.Subir(context.Background(), idB, domain.DocumentoFormacion, "ajeno.pdf", []byte("%PDF"))
	if err != nil {
		t.Fatal(err)
	}
	propioOtroTipo, err := docs.Subir(context.Background(), idA, domain.DocumentoExperiencia, "otro.pdf", []byte("%PDF"))
	if err != nil {
		t.Fatal(err)
	}
	base := domain.FormacionAcademica{Nivel: domain.NivelTercerNivel, Institucion: "UPSE", Tipo: domain.InstitucionNacional, Titulo: "Economista", Area: domain.AreaEconomia, SenescytNumero: "1", SenescytFecha: "2020-01-01"}
	for _, docID := range []string{ajeno.ID, propioOtroTipo.ID} {
		base.DocumentoID = docID
		_, err := svcA.Guardar(context.Background(), idA, domain.Formulario{Formacion: []domain.FormacionAcademica{base}})
		if !errors.Is(err, domain.ErrValidation) {
			t.Fatalf("documento %s: %v", docID, err)
		}
	}
}
