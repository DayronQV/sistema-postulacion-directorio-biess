package postgres_test

import (
	"context"
	"errors"
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
	pool := newPool(t)
	repo := postgres.NewEvaluacionRepo(pool)
	tm := postgres.NewTxManager(pool)
	inicio := time.Date(2026, 11, 20, 14, 0, 0, 0, time.UTC)
	intento, ids := examenDePrueba(t, repo, "1799999931", inicio)
	durante := inicio.Add(5 * time.Minute)

	for _, r := range []struct{ pregunta, opcion int }{{ids[0], 2}, {ids[1], 0}, {ids[0], 1}} {
		if ok, err := guardarEnTx(t, tm, repo, intento.ID, r.pregunta, r.opcion, durante); err != nil || !ok {
			t.Fatalf("guardar %+v: ok = %v, err = %v", r, ok, err)
		}
	}
	got, err := repo.Respuestas(ctx, intento.ID)
	if err != nil || len(got) != 2 || got[ids[0]] != 1 || got[ids[1]] != 0 {
		t.Fatalf("respuestas = %v, err = %v (la segunda respuesta a la misma pregunta la reemplaza)", got, err)
	}

	// A la hora de cierre exacta ya no se guarda.
	if ok, err := guardarEnTx(t, tm, repo, intento.ID, ids[2], 0, intento.Cierre); err != nil || ok {
		t.Fatalf("en el cierre: ok = %v, err = %v", ok, err)
	}
	// Ni en un intento finalizado aunque no haya vencido.
	if ok, _ := finalizarEnTx(t, tm, repo, intento.ID, 50, durante); !ok {
		t.Fatal("no se finalizó")
	}
	if ok, err := guardarEnTx(t, tm, repo, intento.ID, ids[2], 0, durante); err != nil || ok {
		t.Fatalf("en un intento finalizado: ok = %v, err = %v", ok, err)
	}
	if got, _ := repo.Respuestas(ctx, intento.ID); len(got) != 2 {
		t.Fatalf("no debe haberse guardado nada más: %v", got)
	}
}

func TestEvaluacionRespuestas_FinalizarUnaSolaVez(t *testing.T) {
	ctx := context.Background()
	pool := newPool(t)
	repo := postgres.NewEvaluacionRepo(pool)
	tm := postgres.NewTxManager(pool)
	inicio := time.Date(2026, 11, 20, 14, 0, 0, 0, time.UTC)
	intento, _ := examenDePrueba(t, repo, "1799999932", inicio)
	fin := inicio.Add(12 * time.Minute)

	if ok, err := finalizarEnTx(t, tm, repo, intento.ID, 66.67, fin); err != nil || !ok {
		t.Fatalf("primera: ok = %v, err = %v", ok, err)
	}
	if ok, err := finalizarEnTx(t, tm, repo, intento.ID, 10, fin.Add(time.Minute)); err != nil || ok {
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
	pool := newPool(t)
	repo := postgres.NewEvaluacionRepo(pool)
	tm := postgres.NewTxManager(pool)
	inicio := time.Date(2026, 11, 20, 14, 0, 0, 0, time.UTC)
	vencido, _ := examenDePrueba(t, repo, "1799999933", inicio)
	finalizado, _ := examenDePrueba(t, repo, "1799999934", inicio)
	enCurso, _ := examenDePrueba(t, repo, "1799999935", inicio.Add(time.Hour))
	if _, err := finalizarEnTx(t, tm, repo, finalizado.ID, 80, inicio.Add(10*time.Minute)); err != nil {
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
	tm := postgres.NewTxManager(pool)
	inicio := time.Date(2026, 11, 20, 14, 0, 0, 0, time.UTC)
	intento, _ := examenDePrueba(t, repo, "1799999936", inicio)
	if _, err := finalizarEnTx(t, tm, repo, intento.ID, 33.33, inicio.Add(5*time.Minute)); err != nil {
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

// guardarEnTx guarda una respuesta en su propia transacción, como hace el servicio.
func guardarEnTx(t *testing.T, tm *postgres.TxManager, repo *postgres.EvaluacionRepo, intentoID string, preguntaID, opcion int, ahora time.Time) (bool, error) {
	t.Helper()
	var guardada bool
	err := tm.Ejecutar(context.Background(), func(ctx context.Context, tx domain.Tx) error {
		var err error
		guardada, err = repo.GuardarRespuesta(ctx, tx, intentoID, preguntaID, opcion, ahora)
		return err
	})
	return guardada, err
}

// finalizarEnTx bloquea el intento y lo finaliza en una transacción, como hace el servicio.
// Devuelve false si ya estaba finalizado.
func finalizarEnTx(t *testing.T, tm *postgres.TxManager, repo *postgres.EvaluacionRepo, intentoID string, nota float64, fin time.Time) (bool, error) {
	t.Helper()
	var finalizado bool
	err := tm.Ejecutar(context.Background(), func(ctx context.Context, tx domain.Tx) error {
		abierto, err := repo.BloquearParaFinalizar(ctx, tx, intentoID)
		if err != nil || !abierto {
			return err
		}
		finalizado, err = repo.Finalizar(ctx, tx, intentoID, nota, fin)
		return err
	})
	return finalizado, err
}

// Mientras un proceso finaliza el examen (con el intento bloqueado), una respuesta que llega a
// la vez tiene que esperar y, al soltarse el bloqueo, rechazarse: nunca puede quedar guardada
// sin haber entrado en la nota.
func TestEvaluacionRespuestas_UnaRespuestaQueLlegaMientrasSeFinalizaEsperaYSeRechaza(t *testing.T) {
	ctx := context.Background()
	pool := newPool(t)
	repo := postgres.NewEvaluacionRepo(pool)
	tm := postgres.NewTxManager(pool)
	inicio := time.Date(2026, 11, 20, 14, 0, 0, 0, time.UTC)
	intento, ids := examenDePrueba(t, repo, "1799999937", inicio)
	durante := inicio.Add(5 * time.Minute)

	bloqueado := make(chan struct{})
	soltar := make(chan struct{})
	finalizacion := make(chan error, 1)
	go func() {
		finalizacion <- tm.Ejecutar(ctx, func(ctx context.Context, tx domain.Tx) error {
			abierto, err := repo.BloquearParaFinalizar(ctx, tx, intento.ID)
			if err != nil {
				return err
			}
			if !abierto {
				return errors.New("el intento debía estar abierto")
			}
			close(bloqueado)
			<-soltar
			_, err = repo.Finalizar(ctx, tx, intento.ID, 0, durante)
			return err
		})
	}()
	<-bloqueado

	type resultadoGuardar struct {
		ok  bool
		err error
	}
	guardado := make(chan resultadoGuardar, 1)
	go func() {
		ok, err := guardarEnTx(t, tm, repo, intento.ID, ids[0], 1, durante)
		guardado <- resultadoGuardar{ok, err}
	}()

	// Con el intento bloqueado para finalizar, la respuesta no puede terminar.
	select {
	case r := <-guardado:
		t.Fatalf("la respuesta se guardó sin esperar a que terminara la finalización: %+v", r)
	case <-time.After(500 * time.Millisecond):
	}
	close(soltar)
	if err := <-finalizacion; err != nil {
		t.Fatal(err)
	}
	r := <-guardado
	if r.err != nil || r.ok {
		t.Fatalf("la respuesta debía rechazarse al encontrar el examen finalizado: %+v", r)
	}
	if got, err := repo.Respuestas(ctx, intento.ID); err != nil || len(got) != 0 {
		t.Fatalf("no debe quedar ninguna respuesta fuera de la nota: %v, %v", got, err)
	}
}
