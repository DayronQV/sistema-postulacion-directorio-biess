package postgres_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/domain"
	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/repository/postgres"
)

// marcaDePrueba identifica las preguntas de esta ejecución para no mezclarlas con otras
// que ya estén en la base.
func marcaDePrueba() string { return fmt.Sprintf("zz-prueba-%d", time.Now().UnixNano()) }

// limpiarPreguntas borra las preguntas de la prueba al terminar.
func limpiarPreguntas(t *testing.T, pool *pgxpool.Pool, marca string) {
	t.Helper()
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), `DELETE FROM eval_preguntas WHERE enunciado LIKE $1`, marca+"%")
	})
}

func preguntasDePrueba(marca string, n int) []domain.Pregunta {
	ps := make([]domain.Pregunta, n)
	for i := range ps {
		ps[i] = domain.Pregunta{
			Enunciado: fmt.Sprintf("%s pregunta %d", marca, i+1), Opciones: []string{"Uno", "Dos", "Tres"}, RespuestaCorrecta: i % 3,
		}
	}
	return ps
}

func TestEvaluacion_CrearListarYLeerPreguntas(t *testing.T) {
	pool := newPool(t)
	repo := postgres.NewEvaluacionRepo(pool)
	ctx := context.Background()
	marca := marcaDePrueba()
	limpiarPreguntas(t, pool, marca)

	creadas, err := repo.CrearPreguntas(ctx, preguntasDePrueba(marca, 25))
	if err != nil || len(creadas) != 25 {
		t.Fatalf("creadas = %d, err = %v", len(creadas), err)
	}
	if creadas[0].ID == 0 || creadas[24].ID <= creadas[0].ID {
		t.Fatalf("los ids deben ser crecientes: %d … %d", creadas[0].ID, creadas[24].ID)
	}

	// Páginas de 10 ordenadas por id: 10 + 10 + 5, con el total de la búsqueda.
	vistos := 0
	for pagina, esperadas := range []int{10, 10, 5} {
		items, total, err := repo.ListarPreguntas(ctx, domain.FiltroPreguntas{Buscar: marca, Pagina: pagina + 1, PorPagina: 10})
		if err != nil || total != 25 || len(items) != esperadas {
			t.Fatalf("página %d: %d items, total %d, err %v", pagina+1, len(items), total, err)
		}
		for _, p := range items {
			vistos++
			if p.Enunciado != fmt.Sprintf("%s pregunta %d", marca, vistos) || len(p.Opciones) != 3 || p.RespuestaCorrecta != (vistos-1)%3 {
				t.Fatalf("pregunta %d leída = %+v", vistos, p)
			}
		}
	}
	// Más allá de la última página: sin items y el total intacto.
	if items, total, err := repo.ListarPreguntas(ctx, domain.FiltroPreguntas{Buscar: marca, Pagina: 9, PorPagina: 10}); err != nil || len(items) != 0 || total != 25 {
		t.Fatalf("página 9: %d items, total %d, err %v", len(items), total, err)
	}
	// Búsqueda sin distinguir mayúsculas.
	if _, total, err := repo.ListarPreguntas(ctx, domain.FiltroPreguntas{Buscar: marca + " PREGUNTA 25", Pagina: 1, PorPagina: 10}); err != nil || total != 1 {
		t.Fatalf("búsqueda sin mayúsculas: total %d, err %v", total, err)
	}

	p, existe, err := repo.ObtenerPregunta(ctx, creadas[4].ID)
	if err != nil || !existe || p.Enunciado != creadas[4].Enunciado || p.RespuestaCorrecta != 1 {
		t.Fatalf("obtener = %+v, existe = %v, err = %v", p, existe, err)
	}
	if _, existe, err := repo.ObtenerPregunta(ctx, 2_000_000_000); err != nil || existe {
		t.Fatalf("una pregunta inexistente: existe = %v, err = %v", existe, err)
	}
}

func TestEvaluacion_LaBusquedaNoUsaComodines(t *testing.T) {
	pool := newPool(t)
	repo := postgres.NewEvaluacionRepo(pool)
	ctx := context.Background()
	marca := marcaDePrueba()
	limpiarPreguntas(t, pool, marca)

	if _, err := repo.CrearPreguntas(ctx, []domain.Pregunta{
		{Enunciado: marca + " 100% seguro", Opciones: []string{"a", "b"}, RespuestaCorrecta: 0},
		{Enunciado: marca + " abc", Opciones: []string{"a", "b"}, RespuestaCorrecta: 0},
		{Enunciado: marca + " a_c", Opciones: []string{"a", "b"}, RespuestaCorrecta: 0},
	}); err != nil {
		t.Fatal(err)
	}
	for buscar, esperadas := range map[string]int{marca + " 100%": 1, marca + " a_c": 1, marca + " a%c": 0, marca + " a": 2} {
		if _, total, err := repo.ListarPreguntas(ctx, domain.FiltroPreguntas{Buscar: buscar, Pagina: 1, PorPagina: 10}); err != nil || total != esperadas {
			t.Fatalf("buscar %q: total %d, se esperaban %d (err %v)", buscar, total, esperadas, err)
		}
	}
}

func TestEvaluacion_LaCargaEnBloqueEsTodaONinguna(t *testing.T) {
	pool := newPool(t)
	repo := postgres.NewEvaluacionRepo(pool)
	ctx := context.Background()
	marca := marcaDePrueba()
	limpiarPreguntas(t, pool, marca)

	ps := preguntasDePrueba(marca, 5)
	ps[3].RespuestaCorrecta = 40000 // no cabe en la columna SMALLINT: la base la rechaza

	if _, err := repo.CrearPreguntas(ctx, ps); err == nil {
		t.Fatal("se esperaba un error de la base")
	}
	if _, total, err := repo.ListarPreguntas(ctx, domain.FiltroPreguntas{Buscar: marca, Pagina: 1, PorPagina: 10}); err != nil || total != 0 {
		t.Fatalf("no debe quedar ninguna pregunta de una carga fallida: total %d, err %v", total, err)
	}
}

func TestEvaluacion_ActualizarYEliminarPregunta(t *testing.T) {
	pool := newPool(t)
	repo := postgres.NewEvaluacionRepo(pool)
	ctx := context.Background()
	marca := marcaDePrueba()
	limpiarPreguntas(t, pool, marca)

	creadas, err := repo.CrearPreguntas(ctx, preguntasDePrueba(marca, 2))
	if err != nil {
		t.Fatal(err)
	}
	nueva := domain.Pregunta{ID: creadas[0].ID, Enunciado: marca + " editada", Opciones: []string{"x", "y"}, RespuestaCorrecta: 1}
	if ok, err := repo.ActualizarPregunta(ctx, nueva); err != nil || !ok {
		t.Fatalf("actualizar: ok = %v, err = %v", ok, err)
	}
	leida, _, _ := repo.ObtenerPregunta(ctx, nueva.ID)
	if leida.Enunciado != nueva.Enunciado || len(leida.Opciones) != 2 || leida.RespuestaCorrecta != 1 {
		t.Fatalf("lectura tras actualizar = %+v", leida)
	}
	otra, _, _ := repo.ObtenerPregunta(ctx, creadas[1].ID)
	if otra.Enunciado != creadas[1].Enunciado {
		t.Fatal("solo debe cambiar la pregunta indicada")
	}
	if ok, err := repo.ActualizarPregunta(ctx, domain.Pregunta{ID: 2_000_000_000, Enunciado: "x", Opciones: []string{"a", "b"}}); err != nil || ok {
		t.Fatalf("actualizar inexistente: ok = %v, err = %v", ok, err)
	}

	if ok, err := repo.EliminarPregunta(ctx, creadas[0].ID); err != nil || !ok {
		t.Fatalf("eliminar: ok = %v, err = %v", ok, err)
	}
	if _, existe, _ := repo.ObtenerPregunta(ctx, creadas[0].ID); existe {
		t.Fatal("la pregunta sigue existiendo")
	}
	if ok, err := repo.EliminarPregunta(ctx, creadas[0].ID); err != nil || ok {
		t.Fatalf("eliminar otra vez: ok = %v, err = %v", ok, err)
	}
}

func TestEvaluacion_ContarPreguntas(t *testing.T) {
	pool := newPool(t)
	repo := postgres.NewEvaluacionRepo(pool)
	ctx := context.Background()
	antes, err := repo.ContarPreguntas(ctx)
	if err != nil {
		t.Fatal(err)
	}
	marca := marcaDePrueba()
	limpiarPreguntas(t, pool, marca)
	if _, err := repo.CrearPreguntas(ctx, preguntasDePrueba(marca, 4)); err != nil {
		t.Fatal(err)
	}
	if despues, err := repo.ContarPreguntas(ctx); err != nil || despues != antes+4 {
		t.Fatalf("contar: antes %d, después %d, err %v", antes, despues, err)
	}
}

func TestEvaluacion_Configuracion(t *testing.T) {
	pool := newPool(t)
	repo := postgres.NewEvaluacionRepo(pool)
	ctx := context.Background()

	// Se restaura lo que hubiera en la base.
	previa, habia, err := repo.ObtenerConfig(ctx)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if habia {
			_ = repo.GuardarConfig(context.Background(), previa)
		} else {
			_, _ = pool.Exec(context.Background(), `DELETE FROM eval_config`)
		}
	})
	if _, err := pool.Exec(ctx, `DELETE FROM eval_config`); err != nil {
		t.Fatal(err)
	}

	if _, hay, err := repo.ObtenerConfig(ctx); err != nil || hay {
		t.Fatalf("sin configurar: hay = %v, err = %v", hay, err)
	}
	if err := repo.GuardarConfig(ctx, domain.ConfigEvaluacion{DuracionMinutos: 45, NumPreguntas: 8}); err != nil {
		t.Fatal(err)
	}
	c, hay, err := repo.ObtenerConfig(ctx)
	if err != nil || !hay || c.DuracionMinutos != 45 || c.NumPreguntas != 8 {
		t.Fatalf("config leída = %+v, hay = %v, err = %v", c, hay, err)
	}

	// Guardar de nuevo reemplaza la única fila.
	if err := repo.GuardarConfig(ctx, domain.ConfigEvaluacion{DuracionMinutos: 20, NumPreguntas: 3}); err != nil {
		t.Fatal(err)
	}
	c, _, _ = repo.ObtenerConfig(ctx)
	if c.DuracionMinutos != 20 || c.NumPreguntas != 3 {
		t.Fatalf("config tras reemplazar = %+v", c)
	}
	var filas int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM eval_config`).Scan(&filas); err != nil || filas != 1 {
		t.Fatalf("debe haber una sola fila de configuración, hay %d (err %v)", filas, err)
	}
}

func TestEvaluacion_HayIntentos(t *testing.T) {
	pool := newPool(t)
	repo := postgres.NewEvaluacionRepo(pool)
	ctx := context.Background()

	if hay, err := repo.HayIntentos(ctx); err != nil {
		t.Fatal(err)
	} else if hay {
		t.Skip("la base ya tiene intentos de evaluación: se salta la prueba")
	}
	postulacionID := postulacionConEstado(t, pool, "1799999911", domain.EstadoEnviada)
	if _, err := pool.Exec(ctx,
		`INSERT INTO eval_intentos (postulacion_id, preguntas, cierre) VALUES ($1, '[]', now() + interval '1 hour')`, postulacionID); err != nil {
		t.Fatalf("creando el intento: %v", err)
	}
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), `DELETE FROM eval_intentos WHERE postulacion_id = $1`, postulacionID)
	})
	if hay, err := repo.HayIntentos(ctx); err != nil || !hay {
		t.Fatalf("con un intento: hay = %v, err = %v", hay, err)
	}
}
