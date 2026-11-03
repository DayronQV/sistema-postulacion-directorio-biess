package postgres_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/domain"
	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/repository/postgres"
)

func TestFases_SemillaYActualizacion(t *testing.T) {
	pool := newPool(t)
	repo := postgres.NewFasesRepo(pool)
	ctx := context.Background()

	// db/schema.sql siembra las tres fases; Postulación arranca abierta y sin fechas.
	fases, err := repo.Listar(ctx)
	if err != nil {
		t.Fatalf("listando: %v", err)
	}
	nombres := []string{domain.FasePostulacion, domain.FaseMeritos, domain.FaseEvaluacion}
	if len(fases) != len(nombres) {
		t.Fatalf("fases = %+v", fases)
	}
	for i, f := range fases {
		if f.Nombre != nombres[i] {
			t.Fatalf("orden de las fases = %+v", fases)
		}
	}
	original, err := repo.PorNombre(ctx, domain.FasePostulacion)
	if err != nil {
		t.Fatalf("PorNombre: %v", err)
	}
	if !original.FormularioAbierto(time.Now()) {
		t.Fatalf("la fase Postulación debe arrancar abierta: %+v", original)
	}
	t.Cleanup(func() { _ = repo.Actualizar(ctx, original) })

	// Las fechas y las banderas se guardan y se leen igual.
	inicio := time.Date(2026, 11, 1, 5, 0, 0, 0, time.UTC)
	fin := inicio.AddDate(0, 0, 15)
	cambio := original
	cambio.HabilitaFormulario, cambio.FechaInicio, cambio.FechaFin = false, &inicio, &fin
	if err := repo.Actualizar(ctx, cambio); err != nil {
		t.Fatalf("actualizando: %v", err)
	}
	leida, err := repo.Obtener(ctx, original.ID)
	if err != nil {
		t.Fatalf("Obtener: %v", err)
	}
	if leida.HabilitaFormulario || leida.FechaInicio == nil || !leida.FechaInicio.Equal(inicio) ||
		leida.FechaFin == nil || !leida.FechaFin.Equal(fin) {
		t.Fatalf("fase leída = %+v", leida)
	}

	// La base rechaza un fin anterior al inicio aunque el servicio no lo haya validado.
	antes := inicio.Add(-time.Hour)
	cambio.FechaFin = &antes
	if err := repo.Actualizar(ctx, cambio); err == nil {
		t.Fatal("se esperaba que el CHECK rechazara fecha_fin < fecha_inicio")
	}

	if _, err := repo.Obtener(ctx, 9999); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("fase inexistente: %v", err)
	}
	if err := repo.Actualizar(ctx, domain.Fase{ID: 9999}); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("actualizar inexistente: %v", err)
	}
}
