package postgres_test

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/domain"
	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/repository/postgres"
)

// postulacionConEstado crea una postulación mínima en el estado pedido y devuelve su id.
func postulacionConEstado(t *testing.T, pool *pgxpool.Pool, cedula, estado string) string {
	t.Helper()
	postulanteID := crearPostulante(t, pool, cedula)
	var id string
	err := pool.QueryRow(context.Background(), `
INSERT INTO postulaciones (codigo, postulante_id, estado, nombres, apellidos, genero,
                           fecha_nacimiento, nacionalidad, estado_civil)
VALUES ('BIESS-9999-' || lpad(nextval('postulacion_codigo_seq')::text, 6, '0'), $1, $2,
        'Prueba', 'Mérito', 'M', '1990-01-01', 'Ecuatoriana', 'Soltero/a')
RETURNING id::text`, postulanteID, estado).Scan(&id)
	if err != nil {
		t.Fatalf("creando la postulación: %v", err)
	}
	return id
}

func superoMerito(t *testing.T, pool *pgxpool.Pool, id string) bool {
	t.Helper()
	var supero bool
	if err := pool.QueryRow(context.Background(),
		`SELECT supero_merito FROM postulaciones WHERE id = $1`, id).Scan(&supero); err != nil {
		t.Fatalf("leyendo supero_merito: %v", err)
	}
	return supero
}

func TestMerito_SoloSeMarcaUnaPostulacionEnviada(t *testing.T) {
	pool := newPool(t)
	repo := postgres.NewPostulacionRepo(pool)
	ctx := context.Background()

	enviada := postulacionConEstado(t, pool, "1799999901", domain.EstadoEnviada)
	ok, err := repo.ActualizarMerito(ctx, enviada, true)
	if err != nil || !ok || !superoMerito(t, pool, enviada) {
		t.Fatalf("enviada: actualizada = %v, err = %v, marca = %v", ok, err, superoMerito(t, pool, enviada))
	}
	// Se puede quitar la marca.
	if ok, err := repo.ActualizarMerito(ctx, enviada, false); err != nil || !ok || superoMerito(t, pool, enviada) {
		t.Fatalf("quitar la marca: actualizada = %v, err = %v", ok, err)
	}

	// Una borrador o cerrada no se toca: la condición de estado va en el mismo UPDATE.
	for i, estado := range []string{domain.EstadoBorrador, domain.EstadoCerrada} {
		id := postulacionConEstado(t, pool, []string{"1799999902", "1799999903"}[i], estado)
		ok, err := repo.ActualizarMerito(ctx, id, true)
		if err != nil || ok {
			t.Fatalf("%s: se esperaba false sin error, llegó %v, %v", estado, ok, err)
		}
		if superoMerito(t, pool, id) {
			t.Fatalf("%s: no debe cambiar supero_merito", estado)
		}
	}
}

func TestMerito_EstadoPorID(t *testing.T) {
	pool := newPool(t)
	repo := postgres.NewPostulacionRepo(pool)
	ctx := context.Background()

	id := postulacionConEstado(t, pool, "1799999904", domain.EstadoCerrada)
	if estado, err := repo.EstadoPorID(ctx, id); err != nil || estado != domain.EstadoCerrada {
		t.Fatalf("estado = %q, err = %v", estado, err)
	}
	_, err := repo.EstadoPorID(ctx, "99999999-2222-3333-4444-555555555555")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("se esperaba ErrNotFound, llegó %v", err)
	}
}
