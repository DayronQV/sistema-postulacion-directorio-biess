package postgres_test

import (
	"context"
	"testing"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/domain"
	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/integration/registrocivil"
	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/repository/postgres"
)

func TestPostulacion_MarcarEnviadaGuardaEstadoYFecha(t *testing.T) {
	pool := newPool(t)
	postulanteID := crearPostulante(t, pool, registrocivil.CedulaFlujoFeliz)
	if _, err := newService(pool).Obtener(context.Background(), postulanteID); err != nil {
		t.Fatal(err)
	}
	repo := postgres.NewPostulacionRepo(pool)
	ctx := context.Background()

	err := postgres.NewTxManager(pool).Ejecutar(ctx, func(ctx context.Context, tx domain.Tx) error {
		p, err := repo.BloquearPorPostulante(ctx, tx, postulanteID)
		if err != nil {
			return err
		}
		return repo.MarcarEnviada(ctx, tx, p.ID)
	})
	if err != nil {
		t.Fatalf("marcando como enviada: %v", err)
	}
	p, err := repo.ObtenerPorPostulante(ctx, postulanteID)
	if err != nil {
		t.Fatal(err)
	}
	if p.Estado != domain.EstadoEnviada || p.EnviadaEn == nil {
		t.Fatalf("estado = %s, enviada_at = %v", p.Estado, p.EnviadaEn)
	}
}
