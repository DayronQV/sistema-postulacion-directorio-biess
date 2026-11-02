package postgres_test

import (
	"context"
	"testing"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/domain"
	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/integration/registrocivil"
	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/repository/postgres"
)

func respuestasElegibles() map[int]bool {
	r := map[int]bool{}
	for n := 1; n <= domain.TotalPreguntasInhabilidades; n++ {
		r[n] = n == 1
	}
	return r
}

func TestInhabilidades_SeGuardanYCalculanLaElegibilidad(t *testing.T) {
	pool := newPool(t)
	postulanteID := crearPostulante(t, pool, registrocivil.CedulaFlujoFeliz)
	svc := newService(pool, postgres.NewInhabilidadesRepo(pool))
	ctx := context.Background()
	if _, err := svc.Obtener(ctx, postulanteID); err != nil {
		t.Fatal(err)
	}

	// Las 13 respondidas y elegible.
	guardada, err := svc.Guardar(ctx, postulanteID, domain.Formulario{
		Inhabilidades: domain.Inhabilidades{Respuestas: respuestasElegibles()},
	})
	if err != nil {
		t.Fatalf("guardando: %v", err)
	}
	f := guardada.Formulario
	if len(f.Inhabilidades.Respuestas) != 13 || !f.Elegible {
		t.Fatalf("respuestas = %d, elegible = %v", len(f.Inhabilidades.Respuestas), f.Elegible)
	}

	// La 7 en Sí con su detalle deja de ser elegible y conserva el detalle en el JSON.
	resp := respuestasElegibles()
	resp[domain.PreguntaRemovido] = true
	detalle := domain.RemovidoDetalle{Institucion: "SRI", Cargo: "Analista", FechaFin: "2024-01-31", Motivo: "Reestructuración"}
	guardada, err = svc.Guardar(ctx, postulanteID, domain.Formulario{
		Inhabilidades: domain.Inhabilidades{Respuestas: resp, Removido: detalle},
	})
	if err != nil {
		t.Fatalf("guardando con removido: %v", err)
	}
	f = guardada.Formulario
	if f.Elegible || f.Inhabilidades.Removido != detalle || !f.Inhabilidades.Respuestas[domain.PreguntaRemovido] {
		t.Fatalf("elegible = %v, detalle = %+v", f.Elegible, f.Inhabilidades.Removido)
	}

	// Un borrador a medias se guarda, reemplaza lo anterior y no es elegible.
	guardada, err = svc.Guardar(ctx, postulanteID, domain.Formulario{
		Inhabilidades: domain.Inhabilidades{Respuestas: map[int]bool{1: true, 2: false}},
	})
	if err != nil {
		t.Fatalf("guardando a medias: %v", err)
	}
	f = guardada.Formulario
	if len(f.Inhabilidades.Respuestas) != 2 || f.Elegible || f.Inhabilidades.Removido != (domain.RemovidoDetalle{}) {
		t.Fatalf("respuestas = %v, elegible = %v", f.Inhabilidades.Respuestas, f.Elegible)
	}
}
