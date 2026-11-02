package postgres_test

import (
	"context"
	"testing"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/domain"
	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/integration/registrocivil"
	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/repository/postgres"
	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/service"
)

func TestExperiencia_GuardaRecargaYCalculaDuracion(t *testing.T) {
	pool := newPool(t)
	ctx := context.Background()
	postulanteID := crearPostulante(t, pool, registrocivil.CedulaFlujoFeliz)
	svc := newService(pool, postgres.NewExperienciaRepo(pool))
	if _, err := svc.Obtener(ctx, postulanteID); err != nil {
		t.Fatal(err)
	}

	documentos := service.NewDocumentoService(postgres.NewDocumentoRepo(pool))
	doc, err := documentos.Subir(ctx, postulanteID, domain.DocumentoExperiencia, "experiencia.pdf", []byte("%PDF-1.7"))
	if err != nil {
		t.Fatal(err)
	}
	fila := domain.ExperienciaProfesional{
		Tipo: domain.TipoExperienciaGeneral, Cargo: "Analista", Institucion: "Empresa",
		FechaInicio: "2018-01-10", FechaFin: "2020-03-15",
		Actividades: domain.ActividadFinanzas, Descripcion: "Análisis financiero", DocumentoID: doc.ID,
	}
	p, err := svc.Guardar(ctx, postulanteID, domain.Formulario{Experiencia: []domain.ExperienciaProfesional{fila}})
	if err != nil {
		t.Fatal(err)
	}
	if len(p.Formulario.Experiencia) != 1 {
		t.Fatalf("experiencias = %+v", p.Formulario.Experiencia)
	}
	duracion := p.Formulario.Experiencia[0].Duracion
	if duracion != (domain.Duracion{Anios: 2, Meses: 2, Dias: 5}) || p.Formulario.DuracionExperiencia != duracion {
		t.Fatalf("duración individual = %+v, total = %+v", duracion, p.Formulario.DuracionExperiencia)
	}

	recargada, err := svc.Obtener(ctx, postulanteID)
	if err != nil {
		t.Fatal(err)
	}
	if len(recargada.Formulario.Experiencia) != 1 || recargada.Formulario.Experiencia[0].DocumentoID != doc.ID {
		t.Fatalf("experiencia recargada = %+v", recargada.Formulario.Experiencia)
	}
}
