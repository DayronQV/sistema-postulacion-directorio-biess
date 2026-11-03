package postgres_test

import (
	"context"
	"testing"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/domain"
	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/repository/postgres"
)

func TestPostulacionesAdmin_FiltrosPaginacionYDetalle(t *testing.T) {
	pool := newPool(t)
	repo := postgres.NewPostulacionRepo(pool)
	ctx := context.Background()

	idA := postulacionConEstado(t, pool, "1799999910", domain.EstadoEnviada)
	idB := postulacionConEstado(t, pool, "1799999911", domain.EstadoEnviada)
	idC := postulacionConEstado(t, pool, "1799999912", domain.EstadoBorrador)
	if _, err := pool.Exec(ctx,
		`UPDATE postulaciones SET cargo = 'afiliados', elegible = true WHERE id IN ($1, $2)`, idA, idB,
	); err != nil {
		t.Fatalf("preparando filtros: %v", err)
	}
	if _, err := pool.Exec(ctx,
		`UPDATE postulaciones SET cargo = 'jubilados', elegible = false WHERE id = $1`, idC,
	); err != nil {
		t.Fatalf("preparando filtros: %v", err)
	}

	elegible := true
	items, total, err := repo.ListarAdmin(ctx, domain.FiltroPostulacionesAdmin{
		Cargo: domain.CargoAfiliados, Estado: domain.EstadoEnviada, Elegible: &elegible,
		Pagina: 1, PorPagina: 1,
	})
	if err != nil || total != 2 || len(items) != 1 || items[0].Cargo != domain.CargoAfiliados || !items[0].Elegible {
		t.Fatalf("items = %+v, total = %d, err = %v", items, total, err)
	}
	segunda, total, err := repo.ListarAdmin(ctx, domain.FiltroPostulacionesAdmin{
		Cargo: domain.CargoAfiliados, Estado: domain.EstadoEnviada, Elegible: &elegible,
		Pagina: 2, PorPagina: 1,
	})
	if err != nil || total != 2 || len(segunda) != 1 || segunda[0].ID == items[0].ID {
		t.Fatalf("segunda página = %+v, total = %d, err = %v", segunda, total, err)
	}

	detalle, err := repo.ObtenerPorIDAdmin(ctx, idC)
	if err != nil || detalle.ID != idC || detalle.Estado != domain.EstadoBorrador || detalle.Formulario.Cargo != domain.CargoJubilados {
		t.Fatalf("detalle = %+v, err = %v", detalle, err)
	}
}
