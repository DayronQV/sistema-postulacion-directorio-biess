package service_test

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/domain"
	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/service"
)

// fakeExamenes imita las postulaciones y los intentos en memoria; el banco es el fakeBanco.
type fakeExamenes struct {
	banco         *fakeBanco
	postulaciones map[string]domain.PostulacionEvaluacion // por id de postulante
	intentos      map[string]domain.IntentoEvaluacion     // por id de postulación
	// otroGano simula que otra petición creó el intento justo antes de esta.
	otroGano bool

	respuestas     map[string]map[int]int // por id de intento: pregunta → opción del banco
	finalizaciones int                    // veces que Finalizar guardó una nota
	// otroFinalizo simula que la tarea periódica finalizó el intento justo antes.
	otroFinalizo bool
	// alBloquear corre dentro de BloquearParaFinalizar, antes de devolver el bloqueo: simula
	// una respuesta que otra petición confirmó justo antes de que se obtuviera el bloqueo.
	alBloquear func()
	// fueraDeTx cuenta escrituras o lecturas para calificar hechas sin transacción.
	fueraDeTx int
	// fallaAlBloquear hace fallar BloquearParaFinalizar para ese id de intento.
	fallaAlBloquear string
}

// fakeTxEval imita el TxManager: ejecuta fn con una transacción falsa y cuenta las llamadas.
type fakeTxEval struct{ abiertas int }

type txFalsaEval struct{}

func (t *fakeTxEval) Ejecutar(ctx context.Context, fn func(context.Context, domain.Tx) error) error {
	t.abiertas++
	return fn(ctx, txFalsaEval{})
}

// enTx comprueba que el método se llamó dentro de una transacción del fakeTxEval.
func (f *fakeExamenes) enTx(tx domain.Tx) {
	if _, ok := tx.(txFalsaEval); !ok {
		f.fueraDeTx++
	}
}

// idIntento arma un id con forma de UUID a partir del id de la postulación.
func idIntento(postulacionID string) string {
	return fmt.Sprintf("00000000-0000-0000-0000-%012d", len(postulacionID)*1000+int(postulacionID[len(postulacionID)-1]))
}

func (f *fakeExamenes) porID(id string) (string, domain.IntentoEvaluacion, bool) {
	for clave, i := range f.intentos {
		if i.ID == id {
			return clave, i, true
		}
	}
	return "", domain.IntentoEvaluacion{}, false
}

func (f *fakeExamenes) GuardarRespuesta(_ context.Context, tx domain.Tx, intentoID string, preguntaID, opcion int, ahora time.Time) (bool, error) {
	f.enTx(tx)
	_, i, ok := f.porID(intentoID)
	if !ok || i.FinalizadoEn != nil || !ahora.Before(i.Cierre) {
		return false, nil
	}
	if f.respuestas == nil {
		f.respuestas = map[string]map[int]int{}
	}
	if f.respuestas[intentoID] == nil {
		f.respuestas[intentoID] = map[int]int{}
	}
	f.respuestas[intentoID][preguntaID] = opcion
	return true, nil
}

func (f *fakeExamenes) Respuestas(_ context.Context, intentoID string) (map[int]int, error) {
	copia := map[int]int{}
	for k, v := range f.respuestas[intentoID] {
		copia[k] = v
	}
	return copia, nil
}

func (f *fakeExamenes) RespuestasEnTx(ctx context.Context, tx domain.Tx, intentoID string) (map[int]int, error) {
	f.enTx(tx)
	return f.Respuestas(ctx, intentoID)
}

func (f *fakeExamenes) BloquearParaFinalizar(_ context.Context, tx domain.Tx, intentoID string) (bool, error) {
	f.enTx(tx)
	if intentoID == f.fallaAlBloquear {
		return false, errors.New("base de datos caída para este intento")
	}
	clave, i, ok := f.porID(intentoID)
	if !ok {
		return false, nil
	}
	if f.otroFinalizo && i.FinalizadoEn == nil {
		otraNota, otroFin := 12.5, i.Cierre.Add(-time.Second)
		i.Nota, i.FinalizadoEn = &otraNota, &otroFin
		f.intentos[clave] = i
	}
	if i.FinalizadoEn != nil {
		return false, nil
	}
	if f.alBloquear != nil {
		f.alBloquear()
	}
	return true, nil
}

func (f *fakeExamenes) Finalizar(_ context.Context, tx domain.Tx, intentoID string, nota float64, fin time.Time) (bool, error) {
	f.enTx(tx)
	clave, i, ok := f.porID(intentoID)
	if !ok {
		return false, nil
	}
	if i.FinalizadoEn != nil {
		return false, nil
	}
	i.Nota, i.FinalizadoEn = &nota, &fin
	f.intentos[clave] = i
	f.finalizaciones++
	return true, nil
}

func (f *fakeExamenes) VencidosSinFinalizar(_ context.Context, ahora time.Time) ([]domain.IntentoEvaluacion, error) {
	var vencidos []domain.IntentoEvaluacion
	for _, i := range f.intentos {
		if i.FinalizadoEn == nil && !ahora.Before(i.Cierre) {
			vencidos = append(vencidos, i)
		}
	}
	return vencidos, nil
}

func (f *fakeExamenes) IntentoPorID(_ context.Context, id string) (domain.IntentoEvaluacion, bool, error) {
	_, i, ok := f.porID(id)
	return i, ok, nil
}

func (f *fakeExamenes) resultado(i domain.IntentoEvaluacion) domain.ResultadoEvaluacion {
	return domain.ResultadoEvaluacion{
		IntentoID: i.ID, PostulacionID: i.PostulacionID, Codigo: "BIESS-2026-" + i.PostulacionID,
		Cargo: domain.CargoAfiliados, Nombres: "Ana María", Apellidos: "Núñez Íñiguez", Cedula: "1712345675",
		Email: "ana@correo.com", Inicio: i.Inicio, Cierre: i.Cierre, FinalizadoEn: i.FinalizadoEn, Nota: i.Nota,
		TotalPreguntas: len(i.Preguntas),
	}
}

func (f *fakeExamenes) Resultados(context.Context) ([]domain.ResultadoEvaluacion, error) {
	var rs []domain.ResultadoEvaluacion
	for _, i := range f.intentos {
		rs = append(rs, f.resultado(i))
	}
	return rs, nil
}

func (f *fakeExamenes) ResultadoDelIntento(_ context.Context, id string) (domain.ResultadoEvaluacion, bool, error) {
	_, i, ok := f.porID(id)
	if !ok {
		return domain.ResultadoEvaluacion{}, false, nil
	}
	return f.resultado(i), true, nil
}

func (f *fakeExamenes) PreguntasParaSorteo(context.Context) ([]domain.PreguntaSorteable, error) {
	ps := make([]domain.PreguntaSorteable, len(f.banco.preguntas))
	for i, p := range f.banco.preguntas {
		ps[i] = domain.PreguntaSorteable{ID: p.ID, NumOpciones: len(p.Opciones)}
	}
	return ps, nil
}

func (f *fakeExamenes) PostulacionDelPostulante(_ context.Context, id string) (domain.PostulacionEvaluacion, bool, error) {
	p, ok := f.postulaciones[id]
	return p, ok, nil
}

func (f *fakeExamenes) CrearIntento(_ context.Context, i domain.IntentoEvaluacion) (domain.IntentoEvaluacion, bool, error) {
	if existente, ok := f.intentos[i.PostulacionID]; ok {
		return existente, false, nil
	}
	i.ID = idIntento(i.PostulacionID)
	if f.otroGano {
		// El que ganó la carrera guardó un examen con el mismo contenido pero otra hora de cierre.
		i.Cierre = i.Cierre.Add(time.Minute)
		f.intentos[i.PostulacionID] = i
		return i, false, nil
	}
	f.intentos[i.PostulacionID] = i
	return i, true, nil
}

func (f *fakeExamenes) ObtenerIntento(_ context.Context, postulacionID string) (domain.IntentoEvaluacion, bool, error) {
	i, ok := f.intentos[postulacionID]
	return i, ok, nil
}

// desplazado es un Barajador determinista: devuelve 1, 2, ..., n-1, 0 (todo corrido una posición).
type desplazado struct{}

func (desplazado) Permutacion(n int) ([]int, error) {
	p := make([]int, n)
	for i := range p {
		p[i] = (i + 1) % n
	}
	return p, nil
}

var inicioExamen = time.Date(2026, 11, 20, 14, 0, 0, 0, time.UTC)

type escenarioEval struct {
	svc      *service.EvaluacionService
	banco    *fakeBanco
	intentos *fakeExamenes
	fases    *fakeFases
	correo   *fakeMailer
	reloj    *time.Time
	tx       *fakeTxEval
}

// finVentana es el fin de la fase Evaluación: dos horas después del inicio.
var finVentana = inicioExamen.Add(2 * time.Hour)

// nuevoEscenario arma un banco de n preguntas de 4 opciones, una evaluación de 30 minutos que
// pide `pide` preguntas, la fase Evaluación publicada de inicioExamen a finVentana, y un
// postulante "p1" con la postulación enviada y la Fase de Méritos superada. El reloj marca la
// hora de inicio.
func nuevoEscenario(n, pide int, barajador domain.Barajador) escenarioEval {
	banco := &fakeBanco{}
	for i := 0; i < n; i++ {
		_, _ = banco.CrearPreguntas(context.Background(), []domain.Pregunta{{
			Enunciado: fmt.Sprintf("Pregunta %d", i+1), Opciones: []string{"A", "B", "C", "D"}, RespuestaCorrecta: 0,
		}})
	}
	banco.cfg = &domain.ConfigEvaluacion{DuracionMinutos: 30, NumPreguntas: pide}
	fases := nuevasFases()
	inicio, fin := inicioExamen, finVentana
	fases.fases[2].Publicada, fases.fases[2].FechaInicio, fases.fases[2].FechaFin = true, &inicio, &fin
	intentos := &fakeExamenes{
		banco:         banco,
		postulaciones: map[string]domain.PostulacionEvaluacion{"p1": {ID: "post-1", Estado: domain.EstadoEnviada, SuperoMerito: true}},
		intentos:      map[string]domain.IntentoEvaluacion{},
	}
	reloj := inicioExamen
	correo := &fakeMailer{}
	tx := &fakeTxEval{}
	svc := service.NewEvaluacionService(banco, intentos, intentos, fases, tx, barajador, func() time.Time { return reloj }).
		ConMailer(correo, nil)
	return escenarioEval{
		svc: svc, banco: banco, intentos: intentos, fases: fases, correo: correo, reloj: &reloj, tx: tx,
	}
}

// evaluacion devuelve la fase Evaluación del escenario, para cambiarla en cada prueba.
func (e escenarioEval) evaluacion() *domain.Fase { return &e.fases.fases[2] }

func (e escenarioEval) con(f func()) escenarioEval {
	f()
	return e
}

func esperaError(t *testing.T, err error, base error, msg string) {
	t.Helper()
	if !errors.Is(err, base) || (msg != "" && mensaje(err) != msg) {
		t.Fatalf("se esperaba %v con %q, llegó %v", base, msg, err)
	}
}

// Criterio de la issue #43: no se puede iniciar sin haber superado méritos.
func TestIniciar_SoloConPostulacionEnviadaYMeritosSuperados(t *testing.T) {
	ctx := context.Background()

	t.Run("sin postulación", func(t *testing.T) {
		e := nuevoEscenario(10, 5, desplazado{})
		_, _, err := e.svc.Iniciar(ctx, "otro-postulante")
		esperaError(t, err, domain.ErrForbidden, service.MsgEvaluacionSinPostulacion)
	})
	for _, estado := range []string{domain.EstadoBorrador, domain.EstadoCerrada} {
		t.Run("postulación "+estado, func(t *testing.T) {
			e := nuevoEscenario(10, 5, desplazado{})
			e.intentos.postulaciones["p1"] = domain.PostulacionEvaluacion{ID: "post-1", Estado: estado, SuperoMerito: true}
			_, _, err := e.svc.Iniciar(ctx, "p1")
			esperaError(t, err, domain.ErrForbidden, service.MsgEvaluacionSinPostulacion)
		})
	}
	t.Run("sin superar méritos", func(t *testing.T) {
		e := nuevoEscenario(10, 5, desplazado{})
		e.intentos.postulaciones["p1"] = domain.PostulacionEvaluacion{ID: "post-1", Estado: domain.EstadoEnviada, SuperoMerito: false}
		_, _, err := e.svc.Iniciar(ctx, "p1")
		esperaError(t, err, domain.ErrForbidden, service.MsgEvaluacionSinMerito)
		if len(e.intentos.intentos) != 0 {
			t.Fatal("no debe crearse ningún intento")
		}
	})
}

// Criterio de la issue #43: no se puede iniciar fuera del horario. El horario lo fija la fase
// Evaluación: publicada y entre su fecha de inicio y su fecha de fin.
func TestIniciar_SoloConLaFaseEvaluacionPublicadaYDentroDeSusFechas(t *testing.T) {
	ctx := context.Background()

	t.Run("fase sin publicar", func(t *testing.T) {
		e := nuevoEscenario(10, 5, desplazado{})
		e.evaluacion().Publicada = false
		_, _, err := e.svc.Iniciar(ctx, "p1")
		esperaError(t, err, domain.ErrForbidden, service.MsgEvaluacionNoPublicada)
		if len(e.intentos.intentos) != 0 {
			t.Fatal("no debe crearse ningún intento")
		}
	})
	t.Run("sin la fase Evaluación en la base", func(t *testing.T) {
		e := nuevoEscenario(10, 5, desplazado{})
		e.fases.fases = e.fases.fases[:2]
		_, _, err := e.svc.Iniciar(ctx, "p1")
		esperaError(t, err, domain.ErrForbidden, service.MsgEvaluacionNoPublicada)
	})
	t.Run("un segundo antes del inicio", func(t *testing.T) {
		e := nuevoEscenario(10, 5, desplazado{})
		*e.reloj = inicioExamen.Add(-time.Second)
		_, _, err := e.svc.Iniciar(ctx, "p1")
		esperaError(t, err, domain.ErrForbidden, service.MsgEvaluacionNoComienza)
		var de *domain.Error
		if !errors.As(err, &de) || de.Fields["inicia_en"] != "2026-11-20T14:00:00Z" {
			t.Fatalf("el error debe decir cuándo empieza: %+v", de)
		}
		if len(e.intentos.intentos) != 0 {
			t.Fatal("no debe crearse ningún intento")
		}
	})
	t.Run("a la hora exacta de inicio", func(t *testing.T) {
		e := nuevoEscenario(10, 5, desplazado{})
		if _, creado, err := e.svc.Iniciar(ctx, "p1"); err != nil || !creado {
			t.Fatalf("a la hora exacta debe poder iniciar: creado = %v, err = %v", creado, err)
		}
	})
	t.Run("un segundo antes del fin", func(t *testing.T) {
		e := nuevoEscenario(10, 5, desplazado{})
		*e.reloj = finVentana.Add(-time.Second)
		if _, creado, err := e.svc.Iniciar(ctx, "p1"); err != nil || !creado {
			t.Fatalf("un segundo antes del fin todavía debe poder iniciar: creado = %v, err = %v", creado, err)
		}
	})
	for nombre, despues := range map[string]time.Duration{
		"justo en la fecha de fin": 2 * time.Hour,
		"horas después del fin":    5 * time.Hour,
		"días después del fin":     72 * time.Hour,
	} {
		t.Run(nombre, func(t *testing.T) {
			e := nuevoEscenario(10, 5, desplazado{})
			*e.reloj = inicioExamen.Add(despues)
			_, _, err := e.svc.Iniciar(ctx, "p1")
			esperaError(t, err, domain.ErrForbidden, service.MsgEvaluacionPlazoVencido)
			var de *domain.Error
			if !errors.As(err, &de) || de.Fields["cierra_en"] != "2026-11-20T16:00:00Z" {
				t.Fatalf("el error debe decir cuándo cerró: %+v", de)
			}
			if len(e.intentos.intentos) != 0 {
				t.Fatal("no debe crearse ningún intento")
			}
		})
	}
	t.Run("sin fechas no hay límite", func(t *testing.T) {
		e := nuevoEscenario(10, 5, desplazado{})
		e.evaluacion().FechaInicio, e.evaluacion().FechaFin = nil, nil
		*e.reloj = inicioExamen.Add(-100 * 24 * time.Hour)
		if _, creado, err := e.svc.Iniciar(ctx, "p1"); err != nil || !creado {
			t.Fatalf("sin fechas debe poder iniciar: creado = %v, err = %v", creado, err)
		}
	})
	t.Run("solo fecha de inicio: sin límite por el otro lado", func(t *testing.T) {
		e := nuevoEscenario(10, 5, desplazado{})
		e.evaluacion().FechaFin = nil
		*e.reloj = inicioExamen.Add(30 * 24 * time.Hour)
		s, creado, err := e.svc.Iniciar(ctx, "p1")
		if err != nil || !creado || s.TiempoRestante != 30*time.Minute {
			t.Fatalf("sin fecha de fin debe poder iniciar con su tiempo completo: %+v, creado = %v, err = %v", s, creado, err)
		}
	})
	t.Run("solo fecha de fin: sin límite por el otro lado", func(t *testing.T) {
		e := nuevoEscenario(10, 5, desplazado{})
		e.evaluacion().FechaInicio = nil
		*e.reloj = inicioExamen.Add(-100 * time.Hour)
		if _, creado, err := e.svc.Iniciar(ctx, "p1"); err != nil || !creado {
			t.Fatalf("sin fecha de inicio debe poder iniciar: creado = %v, err = %v", creado, err)
		}
	})
}

// La hora de cierre de cada examen es la menor entre su inicio más la duración y la fecha de
// fin de la fase: ventana de 14:00 a 16:00 y examen de una hora.
func TestIniciar_ElExamenNuncaPasaDelFinDeLaFase(t *testing.T) {
	casos := []struct {
		nombre   string
		inicia   time.Duration // después de inicioExamen
		cierraEn time.Duration
		restante time.Duration
	}{
		{"a la hora de apertura", 0, time.Hour, time.Hour},
		{"con la hora completa todavía", 50 * time.Minute, 110 * time.Minute, time.Hour},
		{"justo cuando la hora completa llega al fin", 60 * time.Minute, 2 * time.Hour, time.Hour},
		{"quedan 40 minutos", 80 * time.Minute, 2 * time.Hour, 40 * time.Minute},
		{"quedan 10 minutos", 110 * time.Minute, 2 * time.Hour, 10 * time.Minute},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			e := nuevoEscenario(10, 5, desplazado{})
			e.banco.cfg.DuracionMinutos = 60
			*e.reloj = inicioExamen.Add(c.inicia)
			s, creado, err := e.svc.Iniciar(context.Background(), "p1")
			if err != nil || !creado {
				t.Fatalf("creado = %v, err = %v", creado, err)
			}
			if !s.Cierre.Equal(inicioExamen.Add(c.cierraEn)) || s.TiempoRestante != c.restante {
				t.Fatalf("cierre = %v (restante %v), se esperaba %v (restante %v)", s.Cierre, s.TiempoRestante, inicioExamen.Add(c.cierraEn), c.restante)
			}
			if !e.intentos.intentos["post-1"].Cierre.Equal(s.Cierre) {
				t.Fatal("lo guardado debe ser igual a lo devuelto")
			}
		})
	}
}

// Quien ya empezó conserva su examen hasta su hora de cierre, aunque la fase se cierre o se
// despublique después; un examen que se acortó por el fin de la fase vence en ese instante.
func TestIniciar_QuienYaEmpezoConservaSuExamen(t *testing.T) {
	ctx := context.Background()
	e := nuevoEscenario(10, 5, desplazado{})
	*e.reloj = inicioExamen.Add(100 * time.Minute) // le tocarían 30 min, pero la fase cierra a los 120
	s, creado, err := e.svc.Iniciar(ctx, "p1")
	if err != nil || !creado || s.TiempoRestante != 20*time.Minute {
		t.Fatalf("se acorta al fin de la fase: %+v, creado = %v, err = %v", s, creado, err)
	}

	*e.reloj = inicioExamen.Add(110 * time.Minute)
	e.evaluacion().Publicada = false // el administrador despublica la fase
	s, creado, err = e.svc.Iniciar(ctx, "p1")
	if err != nil || creado || s.TiempoRestante != 10*time.Minute {
		t.Fatalf("debe devolver su examen con lo que le queda: %+v, creado = %v, err = %v", s, creado, err)
	}
	if _, err := e.svc.Pregunta(ctx, "p1", 1); err != nil {
		t.Fatalf("también puede seguir viendo sus preguntas: %v", err)
	}

	*e.reloj = inicioExamen.Add(120 * time.Minute) // llegó el fin de la fase
	_, _, err = e.svc.Iniciar(ctx, "p1")
	esperaError(t, err, domain.ErrConflict, service.MsgEvaluacionTerminada)
	_, err = e.svc.Pregunta(ctx, "p1", 1)
	esperaError(t, err, domain.ErrConflict, service.MsgEvaluacionTerminada)
}

// Mientras la evaluación no se configure se usan los valores por defecto que muestra el panel.
func TestIniciar_SinConfigurarUsaLosValoresPorDefecto(t *testing.T) {
	e := nuevoEscenario(12, 5, desplazado{})
	e.banco.cfg = nil
	s, creado, err := e.svc.Iniciar(context.Background(), "p1")
	if err != nil || !creado {
		t.Fatalf("creado = %v, err = %v", creado, err)
	}
	if s.TotalPreguntas != service.NumPreguntasPorDefecto || s.TiempoRestante != time.Duration(service.DuracionMinutosPorDefecto)*time.Minute {
		t.Fatalf("debe usar los valores por defecto: %+v", s)
	}
}

func TestIniciar_LasHorasSeGuardanConPrecisionDeMicrosegundos(t *testing.T) {
	e := nuevoEscenario(10, 4, desplazado{})
	*e.reloj = inicioExamen.Add(123_456_789 * time.Nanosecond)
	s, _, err := e.svc.Iniciar(context.Background(), "p1")
	if err != nil {
		t.Fatal(err)
	}
	esperado := inicioExamen.Add(123_456 * time.Microsecond)
	if !s.Inicio.Equal(esperado) || !s.Cierre.Equal(esperado.Add(30*time.Minute)) {
		t.Fatalf("las horas deben quedar en microsegundos: %+v", s)
	}
	if !e.intentos.intentos["post-1"].Cierre.Equal(s.Cierre) {
		t.Fatal("lo devuelto al crear debe ser igual a lo guardado")
	}
}

func TestIniciar_BancoConMenosPreguntasQueLaConfiguracionEs422(t *testing.T) {
	e := nuevoEscenario(3, 5, desplazado{})
	_, _, err := e.svc.Iniciar(context.Background(), "p1")
	esperaError(t, err, domain.ErrUnprocessable, service.MsgEvaluacionBancoCorto)
	if len(e.intentos.intentos) != 0 {
		t.Fatal("no debe crearse ningún intento")
	}
}

func TestIniciar_CreaElIntentoConElSorteoYLaHoraDeCierre(t *testing.T) {
	e := nuevoEscenario(10, 4, desplazado{})
	s, creado, err := e.svc.Iniciar(context.Background(), "p1")
	if err != nil || !creado {
		t.Fatalf("creado = %v, err = %v", creado, err)
	}
	if s.TotalPreguntas != 4 || !s.Inicio.Equal(inicioExamen) || !s.Cierre.Equal(inicioExamen.Add(30*time.Minute)) || s.TiempoRestante != 30*time.Minute {
		t.Fatalf("sesión = %+v", s)
	}
	intento := e.intentos.intentos["post-1"]
	if len(intento.Preguntas) != 4 {
		t.Fatalf("el examen debe tener las preguntas que pide la configuración: %d", len(intento.Preguntas))
	}
	// Con el barajador desplazado, el orden de preguntas es banco[1], banco[2], ... y las
	// opciones de cada una quedan corridas una posición: [1 2 3 0].
	for k, p := range intento.Preguntas {
		if p.PreguntaID != e.banco.preguntas[k+1].ID || !slices.Equal(p.Orden, []int{1, 2, 3, 0}) {
			t.Fatalf("pregunta %d del examen = %+v", k+1, p)
		}
	}
}

// Criterio de la issue #43: dos postulantes reciben exámenes distintos.
func TestIniciar_DosPostulantesRecibenExamenesDistintos(t *testing.T) {
	e := nuevoEscenario(30, 10, service.BarajadorCrypto{})
	for _, id := range []string{"p2", "p3", "p4", "p5"} {
		e.intentos.postulaciones[id] = domain.PostulacionEvaluacion{ID: "post-" + id, Estado: domain.EstadoEnviada, SuperoMerito: true}
	}
	var exámenes []string
	for _, id := range []string{"p1", "p2", "p3", "p4", "p5"} {
		if _, _, err := e.svc.Iniciar(context.Background(), id); err != nil {
			t.Fatalf("%s: %v", id, err)
		}
		intento := e.intentos.intentos[e.intentos.postulaciones[id].ID]
		vistas := map[int]bool{}
		var huella strings.Builder
		for _, p := range intento.Preguntas {
			if vistas[p.PreguntaID] {
				t.Fatalf("%s: la pregunta %d está repetida en su examen", id, p.PreguntaID)
			}
			vistas[p.PreguntaID] = true
			orden := slices.Clone(p.Orden)
			slices.Sort(orden)
			if !slices.Equal(orden, []int{0, 1, 2, 3}) {
				t.Fatalf("%s: el orden de opciones %v no es una permutación", id, p.Orden)
			}
			fmt.Fprintf(&huella, "%d:%v;", p.PreguntaID, p.Orden)
		}
		if len(intento.Preguntas) != 10 {
			t.Fatalf("%s: %d preguntas", id, len(intento.Preguntas))
		}
		exámenes = append(exámenes, huella.String())
	}
	distintos := map[string]bool{}
	for _, ex := range exámenes {
		distintos[ex] = true
	}
	if len(distintos) != len(exámenes) {
		t.Fatalf("los exámenes de cinco postulantes deben ser todos distintos: %d distintos de %d", len(distintos), len(exámenes))
	}
}

func TestIniciar_UnSoloIntentoPorPostulante(t *testing.T) {
	ctx := context.Background()
	e := nuevoEscenario(10, 4, service.BarajadorCrypto{})

	primera, creado, err := e.svc.Iniciar(ctx, "p1")
	if err != nil || !creado {
		t.Fatalf("primera: creado = %v, err = %v", creado, err)
	}
	guardado := e.intentos.intentos["post-1"]

	// Diez minutos después, y aunque cambie la configuración, devuelve el mismo examen.
	*e.reloj = inicioExamen.Add(10 * time.Minute)
	e.banco.cfg.NumPreguntas = 2
	e.banco.cfg.DuracionMinutos = 5
	segunda, creado, err := e.svc.Iniciar(ctx, "p1")
	if err != nil || creado {
		t.Fatalf("segunda: creado = %v, err = %v", creado, err)
	}
	if segunda.TotalPreguntas != 4 || !segunda.Cierre.Equal(primera.Cierre) || segunda.TiempoRestante != 20*time.Minute {
		t.Fatalf("debe ser el mismo examen con el tiempo que queda: %+v vs %+v", segunda, primera)
	}
	if len(e.intentos.intentos) != 1 || !slices.Equal(e.intentos.intentos["post-1"].Preguntas[0].Orden, guardado.Preguntas[0].Orden) {
		t.Fatal("no debe crearse ni cambiarse el intento")
	}
}

func TestIniciar_UnIntentoTerminadoOVencidoNoSeReabre(t *testing.T) {
	ctx := context.Background()

	t.Run("vencido", func(t *testing.T) {
		e := nuevoEscenario(10, 4, desplazado{})
		if _, _, err := e.svc.Iniciar(ctx, "p1"); err != nil {
			t.Fatal(err)
		}
		*e.reloj = inicioExamen.Add(30 * time.Minute) // justo a la hora de cierre
		_, _, err := e.svc.Iniciar(ctx, "p1")
		esperaError(t, err, domain.ErrConflict, service.MsgEvaluacionTerminada)
	})
	t.Run("finalizado antes de tiempo", func(t *testing.T) {
		e := nuevoEscenario(10, 4, desplazado{})
		if _, _, err := e.svc.Iniciar(ctx, "p1"); err != nil {
			t.Fatal(err)
		}
		i := e.intentos.intentos["post-1"]
		fin := inicioExamen.Add(5 * time.Minute)
		i.FinalizadoEn = &fin
		e.intentos.intentos["post-1"] = i
		*e.reloj = inicioExamen.Add(6 * time.Minute)
		_, _, err := e.svc.Iniciar(ctx, "p1")
		esperaError(t, err, domain.ErrConflict, service.MsgEvaluacionTerminada)
	})
}

func TestIniciar_SiOtraPeticionCreoElIntentoPrimeroDevuelveEseMismo(t *testing.T) {
	e := nuevoEscenario(10, 4, desplazado{})
	e.intentos.otroGano = true
	s, creado, err := e.svc.Iniciar(context.Background(), "p1")
	if err != nil || creado {
		t.Fatalf("creado = %v, err = %v", creado, err)
	}
	if !s.Cierre.Equal(inicioExamen.Add(31 * time.Minute)) {
		t.Fatalf("debe devolver el intento que ganó la carrera: %+v", s)
	}
}

func TestPregunta_MuestraLaPreguntaConLasOpcionesEnSuOrdenYSinLaRespuesta(t *testing.T) {
	ctx := context.Background()
	e := nuevoEscenario(6, 3, desplazado{})
	// La pregunta 2 del banco, con la respuesta correcta en la opción 0 ("A").
	e.banco.preguntas[1].Opciones = []string{"opción-0", "opción-1", "opción-2", "opción-3"}
	if _, _, err := e.svc.Iniciar(ctx, "p1"); err != nil {
		t.Fatal(err)
	}
	*e.reloj = inicioExamen.Add(12*time.Minute + 30*time.Second)

	p, err := e.svc.Pregunta(ctx, "p1", 1)
	if err != nil {
		t.Fatal(err)
	}
	// El examen empieza con banco[1]; con el orden [1 2 3 0] se le muestran sus opciones corridas.
	esperadas := []string{"opción-1", "opción-2", "opción-3", "opción-0"}
	if p.Numero != 1 || p.Total != 3 || p.Pregunta.ID != e.banco.preguntas[1].ID || p.Pregunta.Enunciado != "Pregunta 2" ||
		!slices.Equal(p.Pregunta.Opciones, esperadas) {
		t.Fatalf("pregunta = %+v", p)
	}
	if p.TiempoRestante != 17*time.Minute+30*time.Second || !p.Cierre.Equal(inicioExamen.Add(30*time.Minute)) {
		t.Fatalf("el tiempo lo calcula el servidor: %+v", p)
	}
	// El banco no se alteró al barajar las opciones de la vista.
	if e.banco.preguntas[1].Opciones[0] != "opción-0" {
		t.Fatal("barajar la vista no debe cambiar el banco")
	}
}

func TestPregunta_NumeroFueraDeRangoEs404(t *testing.T) {
	ctx := context.Background()
	e := nuevoEscenario(6, 3, desplazado{})
	if _, _, err := e.svc.Iniciar(ctx, "p1"); err != nil {
		t.Fatal(err)
	}
	for _, n := range []int{-1, 0, 4, 100} {
		_, err := e.svc.Pregunta(ctx, "p1", n)
		esperaError(t, err, domain.ErrNotFound, service.MsgPreguntaFueraDeRango)
	}
	for _, n := range []int{1, 2, 3} {
		if _, err := e.svc.Pregunta(ctx, "p1", n); err != nil {
			t.Fatalf("pregunta %d: %v", n, err)
		}
	}
}

func TestPregunta_SinHaberIniciadoEs404(t *testing.T) {
	ctx := context.Background()
	e := nuevoEscenario(6, 3, desplazado{})
	_, err := e.svc.Pregunta(ctx, "p1", 1)
	esperaError(t, err, domain.ErrNotFound, service.MsgEvaluacionNoIniciada)
	_, err = e.svc.Pregunta(ctx, "desconocido", 1)
	esperaError(t, err, domain.ErrNotFound, service.MsgEvaluacionNoIniciada)
}

func TestPregunta_ConElTiempoAgotadoOFinalizadoEs409(t *testing.T) {
	ctx := context.Background()
	e := nuevoEscenario(6, 3, desplazado{})
	if _, _, err := e.svc.Iniciar(ctx, "p1"); err != nil {
		t.Fatal(err)
	}
	*e.reloj = inicioExamen.Add(29*time.Minute + 59*time.Second)
	if _, err := e.svc.Pregunta(ctx, "p1", 1); err != nil {
		t.Fatalf("un segundo antes del cierre todavía se puede ver: %v", err)
	}
	*e.reloj = inicioExamen.Add(30 * time.Minute)
	_, err := e.svc.Pregunta(ctx, "p1", 1)
	esperaError(t, err, domain.ErrConflict, service.MsgEvaluacionTerminada)
}

func TestPregunta_UnIntentoInconsistenteEsUnErrorInternoNoUna404(t *testing.T) {
	ctx := context.Background()
	e := nuevoEscenario(6, 3, desplazado{})
	if _, _, err := e.svc.Iniciar(ctx, "p1"); err != nil {
		t.Fatal(err)
	}
	// La pregunta desapareció del banco.
	e.banco.preguntas = e.banco.preguntas[:0]
	_, err := e.svc.Pregunta(ctx, "p1", 1)
	if err == nil || errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("se esperaba un error interno, llegó %v", err)
	}
}

func TestBarajadorCrypto_DevuelvePermutacionesValidasYVariadas(t *testing.T) {
	b := service.BarajadorCrypto{}
	if p, err := b.Permutacion(0); err != nil || len(p) != 0 {
		t.Fatalf("n=0: %v, %v", p, err)
	}
	if p, err := b.Permutacion(1); err != nil || !slices.Equal(p, []int{0}) {
		t.Fatalf("n=1: %v, %v", p, err)
	}
	vistas := map[string]bool{}
	for i := 0; i < 2000; i++ {
		p, err := b.Permutacion(3)
		if err != nil {
			t.Fatal(err)
		}
		orden := slices.Clone(p)
		slices.Sort(orden)
		if !slices.Equal(orden, []int{0, 1, 2}) {
			t.Fatalf("%v no es una permutación de 0..2", p)
		}
		vistas[fmt.Sprint(p)] = true
	}
	if len(vistas) != 6 {
		t.Fatalf("en 2000 sorteos deben salir las 6 permutaciones de 3 elementos, salieron %d", len(vistas))
	}
}
