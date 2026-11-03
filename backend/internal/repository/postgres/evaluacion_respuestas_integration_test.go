package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/domain"
	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/repository/postgres"
)

// examenDePrueba crea preguntas en el banco, una postulación enviada y su intento, con cierre
// media hora después de inicio. Devuelve el intento y los ids de las preguntas.
func examenDePrueba(t *testing.T, repo *postgres.EvaluacionRepo, cedula string, inicio time.Time) (domain.IntentoEvaluacion, []int) {
	t.Helper()
	ctx := context.Background()
	pool := newPool(t)
	marca := marcaDePrueba()
	limpiarPreguntas(t, pool, marca)
	creadas, err := repo.CrearPreguntas(ctx, preguntasDePrueba(marca, 3))
	if err != nil {
		t.Fatal(err)
	}
	ids := []int{creadas[0].ID, creadas[1].ID, creadas[2].ID}
	postulacionID := postulacionConEstado(t, pool, cedula, domain.EstadoEnviada)
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), `DELETE FROM eval_intentos WHERE postulacion_id = $1`, postulacionID)
	})
	intento, _, err := repo.CrearIntento(ctx, domain.IntentoEvaluacion{
		PostulacionID: postulacionID,
		Preguntas: []domain.PreguntaDelIntento{
			{PreguntaID: ids[0], Orden: []int{1, 2, 0}}, {PreguntaID: ids[1], Orden: []int{0, 1, 2}}, {PreguntaID: ids[2], Orden: []int{2, 0, 1}},
		},
		Inicio: inicio, Cierre: inicio.Add(30 * time.Minute),
	})
	if err != nil {
		t.Fatal(err)
	}
	return intento, ids
}

func TestEvaluacionRespuestas_GuardarReemplazarYLeer(t *testing.T) {
	ctx := context.Background()
	repo := postgres.NewEvaluacionRepo(newPool(t))
	inicio := time.Date(2026, 11, 20, 14, 0, 0, 0, time.UTC)
	intento, ids := examenDePrueba(t, repo, "1799999931", inicio)
	durante := inicio.Add(5 * time.Minute)

	for _, r := range []struct{ pregunta, opcion int }{{ids[0], 2}, {ids[1], 0}, {ids[0], 1}} {
		if ok, err := repo.GuardarRespuesta(ctx, intento.ID, r.pregunta, r.opcion, durante); err != nil || !ok {
			t.Fatalf("guardar %+v: ok = %v, err = %v", r, ok, err)
		}
	}
	got, err := repo.Respuestas(ctx, intento.ID)
	if err != nil || len(got) != 2 || got[ids[0]] != 1 || got[ids[1]] != 0 {
		t.Fatalf("respuestas = %v, err = %v (la segunda respuesta a la misma pregunta la reemplaza)", got, err)
	}

	// A la hora de cierre exacta ya no se guarda.
	if ok, err := repo.GuardarRespuesta(ctx, intento.ID, ids[2], 0, intento.Cierre); err != nil || ok {
		t.Fatalf("en el cierre: ok = %v, err = %v", ok, err)
	}
	// Ni en un intento finalizado aunque no haya vencido.
	if ok, _ := repo.Finalizar(ctx, intento.ID, 50, durante); !ok {
		t.Fatal("no se finalizó")
	}
	if ok, err := repo.GuardarRespuesta(ctx, intento.ID, ids[2], 0, durante); err != nil || ok {
		t.Fatalf("en un intento finalizado: ok = %v, err = %v", ok, err)
	}
	if got, _ := repo.Respuestas(ctx, intento.ID); len(got) != 2 {
		t.Fatalf("no debe haberse guardado nada más: %v", got)
	}
}

func TestEvaluacionRespuestas_FinalizarUnaSolaVez(t *testing.T) {
	ctx := context.Background()
	repo := postgres.NewEvaluacionRepo(newPool(t))
	inicio := time.Date(2026, 11, 20, 14, 0, 0, 0, time.UTC)
	intento, _ := examenDePrueba(t, repo, "1799999932", inicio)
	fin := inicio.Add(12 * time.Minute)

	if ok, err := repo.Finalizar(ctx, intento.ID, 66.67, fin); err != nil || !ok {
		t.Fatalf("primera: ok = %v, err = %v", ok, err)
	}
	if ok, err := repo.Finalizar(ctx, intento.ID, 10, fin.Add(time.Minute)); err != nil || ok {
		t.Fatalf("segunda: ok = %v, err = %v", ok, err)
	}
	leido, existe, err := repo.IntentoPorID(ctx, intento.ID)
	if err != nil || !existe || leido.Nota == nil || *leido.Nota != 66.67 || leido.FinalizadoEn == nil || !leido.FinalizadoEn.Equal(fin) {
		t.Fatalf("debe quedar la primera nota: %+v, %v", leido, err)
	}
	if _, existe, err := repo.IntentoPorID(ctx, "99999999-9999-9999-9999-999999999999"); err != nil || existe {
		t.Fatalf("intento inexistente: existe = %v, err = %v", existe, err)
	}
}

func TestEvaluacionRespuestas_VencidosSinFinalizar(t *testing.T) {
	ctx := context.Background()
	repo := postgres.NewEvaluacionRepo(newPool(t))
	inicio := time.Date(2026, 11, 20, 14, 0, 0, 0, time.UTC)
	vencido, _ := examenDePrueba(t, repo, "1799999933", inicio)
	finalizado, _ := examenDePrueba(t, repo, "1799999934", inicio)
	enCurso, _ := examenDePrueba(t, repo, "1799999935", inicio.Add(time.Hour))
	if _, err := repo.Finalizar(ctx, finalizado.ID, 80, inicio.Add(10*time.Minute)); err != nil {
		t.Fatal(err)
	}

	lista, err := repo.VencidosSinFinalizar(ctx, inicio.Add(45*time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	ids := map[string]bool{}
	for _, i := range lista {
		ids[i.ID] = true
	}
	if !ids[vencido.ID] || ids[finalizado.ID] || ids[enCurso.ID] {
		t.Fatalf("solo el vencido sin finalizar: vencido %v, finalizado %v, en curso %v", ids[vencido.ID], ids[finalizado.ID], ids[enCurso.ID])
	}
}

func TestEvaluacionRespuestas_ResultadosConLosDatosDelPostulante(t *testing.T) {
	ctx := context.Background()
	pool := newPool(t)
	repo := postgres.NewEvaluacionRepo(pool)
	inicio := time.Date(2026, 11, 20, 14, 0, 0, 0, time.UTC)
	intento, _ := examenDePrueba(t, repo, "1799999936", inicio)
	if _, err := repo.Finalizar(ctx, intento.ID, 33.33, inicio.Add(5*time.Minute)); err != nil {
		t.Fatal(err)
	}
	var codigo string
	if err := pool.QueryRow(ctx, `SELECT codigo FROM postulaciones WHERE id = $1`, intento.PostulacionID).Scan(&codigo); err != nil {
		t.Fatal(err)
	}

	r, existe, err := repo.ResultadoDelIntento(ctx, intento.ID)
	if err != nil || !existe {
		t.Fatalf("existe = %v, err = %v", existe, err)
	}
	if r.Codigo != codigo || r.Cedula != "1799999936" || r.Email != "1799999936@prueba.test" || r.Nombres != "Prueba" ||
		r.TotalPreguntas != 3 || r.Nota == nil || *r.Nota != 33.33 || r.FinalizadoEn == nil || r.Cargo != "" {
		t.Fatalf("resultado = %+v", r)
	}
	todos, err := repo.Resultados(ctx)
	if err != nil {
		t.Fatal(err)
	}
	encontrado := false
	for _, x := range todos {
		encontrado = encontrado || x.IntentoID == intento.ID
	}
	if !encontrado {
		t.Fatal("el listado debe incluir el resultado")
	}
	if _, existe, err := repo.ResultadoDelIntento(ctx, "99999999-9999-9999-9999-999999999999"); err != nil || existe {
		t.Fatalf("inexistente: existe = %v, err = %v", existe, err)
	}
}
