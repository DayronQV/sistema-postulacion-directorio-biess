package postgres_test

import (
	"context"
	"slices"
	"sync"
	"testing"
	"time"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/domain"
	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/repository/postgres"
)

// intentoDePrueba arma un intento de dos preguntas con su orden de opciones.
func intentoDePrueba(postulacionID string, inicio time.Time) domain.IntentoEvaluacion {
	return domain.IntentoEvaluacion{
		PostulacionID: postulacionID,
		Preguntas: []domain.PreguntaDelIntento{
			{PreguntaID: 11, Orden: []int{2, 0, 1, 3}},
			{PreguntaID: 7, Orden: []int{1, 0}},
		},
		Inicio: inicio,
		Cierre: inicio.Add(30 * time.Minute),
	}
}

func TestEvaluacionIntentos_PostulacionDelPostulante(t *testing.T) {
	pool := newPool(t)
	repo := postgres.NewEvaluacionRepo(pool)
	ctx := context.Background()

	postulanteID := crearPostulante(t, pool, "1799999921")
	if _, ok, err := repo.PostulacionDelPostulante(ctx, postulanteID); err != nil || ok {
		t.Fatalf("sin postulación: ok = %v, err = %v", ok, err)
	}

	postulacionID := postulacionConEstado(t, pool, "1799999922", domain.EstadoEnviada)
	var postulanteDeEsa string
	if err := pool.QueryRow(ctx, `SELECT postulante_id::text FROM postulaciones WHERE id = $1`, postulacionID).Scan(&postulanteDeEsa); err != nil {
		t.Fatal(err)
	}
	p, ok, err := repo.PostulacionDelPostulante(ctx, postulanteDeEsa)
	if err != nil || !ok || p.ID != postulacionID || p.Estado != domain.EstadoEnviada || p.SuperoMerito {
		t.Fatalf("postulación = %+v, ok = %v, err = %v", p, ok, err)
	}
	if _, err := pool.Exec(ctx, `UPDATE postulaciones SET supero_merito = TRUE WHERE id = $1`, postulacionID); err != nil {
		t.Fatal(err)
	}
	if p, _, _ := repo.PostulacionDelPostulante(ctx, postulanteDeEsa); !p.SuperoMerito {
		t.Fatal("debe reflejar que superó la Fase de Méritos")
	}
}

func TestEvaluacionIntentos_CrearYLeerElIntento(t *testing.T) {
	pool := newPool(t)
	repo := postgres.NewEvaluacionRepo(pool)
	ctx := context.Background()

	postulacionID := postulacionConEstado(t, pool, "1799999923", domain.EstadoEnviada)
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), `DELETE FROM eval_intentos WHERE postulacion_id = $1`, postulacionID)
	})

	if _, ok, err := repo.ObtenerIntento(ctx, postulacionID); err != nil || ok {
		t.Fatalf("antes de crear: ok = %v, err = %v", ok, err)
	}
	inicio := time.Date(2026, 11, 20, 14, 0, 0, 0, time.UTC)
	creado, nuevo, err := repo.CrearIntento(ctx, intentoDePrueba(postulacionID, inicio))
	if err != nil || !nuevo || creado.ID == "" {
		t.Fatalf("crear: %+v, nuevo = %v, err = %v", creado, nuevo, err)
	}

	leido, ok, err := repo.ObtenerIntento(ctx, postulacionID)
	if err != nil || !ok {
		t.Fatalf("leer: ok = %v, err = %v", ok, err)
	}
	if leido.ID != creado.ID || leido.PostulacionID != postulacionID || !leido.Inicio.Equal(inicio) || !leido.Cierre.Equal(inicio.Add(30*time.Minute)) {
		t.Fatalf("intento leído = %+v", leido)
	}
	if leido.FinalizadoEn != nil || leido.Nota != nil {
		t.Fatalf("un intento nuevo no está finalizado ni calificado: %+v", leido)
	}
	if len(leido.Preguntas) != 2 || leido.Preguntas[0].PreguntaID != 11 || !slices.Equal(leido.Preguntas[0].Orden, []int{2, 0, 1, 3}) ||
		leido.Preguntas[1].PreguntaID != 7 || !slices.Equal(leido.Preguntas[1].Orden, []int{1, 0}) {
		t.Fatalf("las preguntas deben conservar su posición y su orden de opciones: %+v", leido.Preguntas)
	}
}

// Criterio: un solo intento por postulante, aunque lleguen dos peticiones a la vez.
func TestEvaluacionIntentos_UnSoloIntentoPorPostulacionAunqueSeaConcurrente(t *testing.T) {
	pool := newPool(t)
	repo := postgres.NewEvaluacionRepo(pool)
	ctx := context.Background()

	postulacionID := postulacionConEstado(t, pool, "1799999924", domain.EstadoEnviada)
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), `DELETE FROM eval_intentos WHERE postulacion_id = $1`, postulacionID)
	})

	const peticiones = 8
	var (
		wg      sync.WaitGroup
		mu      sync.Mutex
		nuevos  int
		ids     = map[string]bool{}
		errores []error
	)
	inicio := time.Date(2026, 11, 20, 14, 0, 0, 0, time.UTC)
	for i := 0; i < peticiones; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			intento := intentoDePrueba(postulacionID, inicio)
			intento.Preguntas[0].PreguntaID = 100 + i // cada petición trae un sorteo distinto
			got, nuevo, err := repo.CrearIntento(ctx, intento)
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				errores = append(errores, err)
				return
			}
			if nuevo {
				nuevos++
			}
			ids[got.ID] = true
		}(i)
	}
	wg.Wait()
	if len(errores) > 0 {
		t.Fatalf("errores: %v", errores)
	}
	if nuevos != 1 || len(ids) != 1 {
		t.Fatalf("debe crearse exactamente un intento: nuevos = %d, ids distintos = %d", nuevos, len(ids))
	}
	var filas int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM eval_intentos WHERE postulacion_id = $1`, postulacionID).Scan(&filas); err != nil || filas != 1 {
		t.Fatalf("filas = %d, err = %v", filas, err)
	}
	// Todos recibieron el mismo examen: el del que ganó.
	guardado, _, _ := repo.ObtenerIntento(ctx, postulacionID)
	if guardado.Preguntas[0].PreguntaID < 100 || guardado.Preguntas[0].PreguntaID >= 100+peticiones {
		t.Fatalf("el examen guardado no es de ninguna petición: %+v", guardado.Preguntas)
	}
}

func TestEvaluacionIntentos_UnaSegundaCreacionDevuelveElIntentoOriginal(t *testing.T) {
	pool := newPool(t)
	repo := postgres.NewEvaluacionRepo(pool)
	ctx := context.Background()

	postulacionID := postulacionConEstado(t, pool, "1799999925", domain.EstadoEnviada)
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), `DELETE FROM eval_intentos WHERE postulacion_id = $1`, postulacionID)
	})

	inicio := time.Date(2026, 11, 20, 14, 0, 0, 0, time.UTC)
	primero, _, err := repo.CrearIntento(ctx, intentoDePrueba(postulacionID, inicio))
	if err != nil {
		t.Fatal(err)
	}
	otro := intentoDePrueba(postulacionID, inicio.Add(time.Hour))
	otro.Preguntas = []domain.PreguntaDelIntento{{PreguntaID: 999, Orden: []int{0, 1}}}
	segundo, nuevo, err := repo.CrearIntento(ctx, otro)
	if err != nil || nuevo {
		t.Fatalf("nuevo = %v, err = %v", nuevo, err)
	}
	if segundo.ID != primero.ID || !segundo.Inicio.Equal(inicio) || len(segundo.Preguntas) != 2 || segundo.Preguntas[0].PreguntaID != 11 {
		t.Fatalf("debe devolver el intento original sin cambios: %+v", segundo)
	}
}

func TestEvaluacionIntentos_PreguntasParaSorteo(t *testing.T) {
	pool := newPool(t)
	repo := postgres.NewEvaluacionRepo(pool)
	ctx := context.Background()
	marca := marcaDePrueba()
	limpiarPreguntas(t, pool, marca)

	creadas, err := repo.CrearPreguntas(ctx, []domain.Pregunta{
		{Enunciado: marca + " a", Opciones: []string{"1", "2"}, RespuestaCorrecta: 0},
		{Enunciado: marca + " b", Opciones: []string{"1", "2", "3", "4", "5"}, RespuestaCorrecta: 3},
	})
	if err != nil {
		t.Fatal(err)
	}
	todas, err := repo.PreguntasParaSorteo(ctx)
	if err != nil {
		t.Fatal(err)
	}
	cuentas := map[int]int{}
	for i, p := range todas {
		cuentas[p.ID] = p.NumOpciones
		if i > 0 && todas[i-1].ID >= p.ID {
			t.Fatalf("deben venir ordenadas por id: %v", todas)
		}
	}
	if cuentas[creadas[0].ID] != 2 || cuentas[creadas[1].ID] != 5 {
		t.Fatalf("cantidad de opciones por pregunta = %v", cuentas)
	}
}
