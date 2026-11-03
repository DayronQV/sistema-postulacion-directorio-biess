package service_test

import (
	"bytes"
	"context"
	"errors"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/domain"
	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/service"
)

// correoCaido es un mailer que siempre falla.
type correoCaido struct{ intentos int }

func (c *correoCaido) Enviar(context.Context, domain.Mensaje) error {
	c.intentos++
	return errors.New("servidor de correo caído")
}

// iniciado arma un escenario con el examen de "p1" ya iniciado: n preguntas en el banco,
// pide de ellas y opciones barajadas con desplazado (la opción correcta del banco, la 0, se
// muestra en la posición 3).
func iniciado(t *testing.T, n, pide int) escenarioEval {
	t.Helper()
	e := nuevoEscenario(n, pide, desplazado{})
	if _, _, err := e.svc.Iniciar(context.Background(), "p1"); err != nil {
		t.Fatal(err)
	}
	return e
}

func (e escenarioEval) intento() domain.IntentoEvaluacion { return e.intentos.intentos["post-1"] }

func opc(n int) *int { return &n }

const posicionCorrecta = 3 // con desplazado, la opción 0 del banco se muestra en la posición 3

func TestGuardarRespuesta_GuardaLaOpcionDelBancoYSeVuelveAMostrar(t *testing.T) {
	ctx := context.Background()
	e := iniciado(t, 6, 3)

	if err := e.svc.GuardarRespuesta(ctx, "p1", 1, opc(posicionCorrecta)); err != nil {
		t.Fatal(err)
	}
	intento := e.intento()
	pregunta := intento.Preguntas[0].PreguntaID
	if got := e.intentos.respuestas[intento.ID][pregunta]; got != 0 {
		t.Fatalf("se guarda el índice del banco: se esperaba 0, se guardó %d", got)
	}
	p, err := e.svc.Pregunta(ctx, "p1", 1)
	if err != nil || p.Respuesta == nil || *p.Respuesta != posicionCorrecta {
		t.Fatalf("la pregunta debe volver con la opción marcada en su posición: %+v, %v", p.Respuesta, err)
	}
	// Una pregunta sin responder no trae respuesta.
	if p, _ := e.svc.Pregunta(ctx, "p1", 2); p.Respuesta != nil {
		t.Fatalf("sin responder debe ser nil: %v", *p.Respuesta)
	}

	// Se puede cambiar de opinión mientras el examen siga abierto.
	if err := e.svc.GuardarRespuesta(ctx, "p1", 1, opc(0)); err != nil {
		t.Fatal(err)
	}
	if got := e.intentos.respuestas[intento.ID][pregunta]; got != 1 {
		t.Fatalf("la posición 0 es la opción 1 del banco, se guardó %d", got)
	}
}

func TestGuardarRespuesta_DatosInvalidos(t *testing.T) {
	ctx := context.Background()
	e := iniciado(t, 6, 3)
	for nombre, c := range map[string]struct {
		n      int
		opcion *int
		base   error
	}{
		"sin opción":            {1, nil, domain.ErrValidation},
		"opción negativa":       {1, opc(-1), domain.ErrValidation},
		"opción fuera de rango": {1, opc(4), domain.ErrValidation},
		"pregunta 0":            {0, opc(0), domain.ErrNotFound},
		"pregunta de más":       {4, opc(0), domain.ErrNotFound},
	} {
		t.Run(nombre, func(t *testing.T) {
			err := e.svc.GuardarRespuesta(ctx, "p1", c.n, c.opcion)
			if !errors.Is(err, c.base) {
				t.Fatalf("se esperaba %v, llegó %v", c.base, err)
			}
			if c.base == domain.ErrValidation && camposEval(t, err)["opcion"] == "" {
				t.Fatalf("falta fields.opcion: %v", err)
			}
		})
	}
	if len(e.intentos.respuestas[e.intento().ID]) != 0 {
		t.Fatal("nada inválido debe guardarse")
	}
	// Sin haber iniciado.
	otro := nuevoEscenario(6, 3, desplazado{})
	if err := otro.svc.GuardarRespuesta(ctx, "p1", 1, opc(0)); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("sin examen se esperaba 404, llegó %v", err)
	}
}

func TestGuardarRespuesta_DespuesDelCierreEs409YCierraElExamen(t *testing.T) {
	ctx := context.Background()
	e := iniciado(t, 6, 3)
	if err := e.svc.GuardarRespuesta(ctx, "p1", 1, opc(posicionCorrecta)); err != nil {
		t.Fatal(err)
	}
	cierre := e.intento().Cierre
	*e.reloj = cierre.Add(5 * time.Minute)

	err := e.svc.GuardarRespuesta(ctx, "p1", 2, opc(posicionCorrecta))
	esperaError(t, err, domain.ErrConflict, service.MsgEvaluacionTerminada)
	intento := e.intento()
	if intento.FinalizadoEn == nil || !intento.FinalizadoEn.Equal(cierre) {
		t.Fatalf("el examen vencido se finaliza a su hora de cierre: %+v", intento.FinalizadoEn)
	}
	if *intento.Nota != 33.33 {
		t.Fatalf("con la respuesta guardada antes del cierre (1 de 3), la nota es 33,33: %v", *intento.Nota)
	}
	if len(e.intentos.respuestas[intento.ID]) != 1 {
		t.Fatal("la respuesta tardía no debe guardarse")
	}
	if len(e.correo.enviados) != 1 {
		t.Fatalf("debe enviarse el reporte una vez, se enviaron %d", len(e.correo.enviados))
	}
}

func TestFinalizar_CalificaEnviaElReporteYEsIdempotente(t *testing.T) {
	ctx := context.Background()
	e := iniciado(t, 6, 3)
	_ = e.svc.GuardarRespuesta(ctx, "p1", 1, opc(posicionCorrecta)) // correcta
	_ = e.svc.GuardarRespuesta(ctx, "p1", 2, opc(posicionCorrecta)) // correcta
	_ = e.svc.GuardarRespuesta(ctx, "p1", 3, opc(0))                // incorrecta
	*e.reloj = inicioExamen.Add(10 * time.Minute)

	r, err := e.svc.Finalizar(ctx, "p1")
	if err != nil {
		t.Fatal(err)
	}
	if r.Nota != 66.67 || r.Aciertos != 2 || r.Total != 3 || !r.FinalizadoEn.Equal(inicioExamen.Add(10*time.Minute)) {
		t.Fatalf("resultado = %+v", r)
	}
	intento := e.intento()
	if intento.FinalizadoEn == nil || *intento.Nota != 66.67 {
		t.Fatalf("la nota debe quedar guardada: %+v", intento)
	}

	// El correo lleva el reporte PDF.
	if len(e.correo.enviados) != 1 {
		t.Fatalf("se esperaba un correo, hubo %d", len(e.correo.enviados))
	}
	m := e.correo.enviados[0]
	if m.Para != "ana@correo.com" || !strings.Contains(m.Texto, "66,67") || len(m.Adjuntos) != 1 {
		t.Fatalf("correo = %+v", m)
	}
	if m.Adjuntos[0].Nombre != "reporte-evaluacion-BIESS-2026-post-1.pdf" || !bytes.HasPrefix(m.Adjuntos[0].Contenido, []byte("%PDF")) {
		t.Fatalf("el adjunto debe ser el reporte PDF: %s", m.Adjuntos[0].Nombre)
	}

	// Finalizar otra vez devuelve lo mismo y no vuelve a enviar ni a calificar.
	*e.reloj = inicioExamen.Add(20 * time.Minute)
	r2, err := e.svc.Finalizar(ctx, "p1")
	if err != nil || r2 != r {
		t.Fatalf("segunda vez = %+v, %v (antes %+v)", r2, err, r)
	}
	if len(e.correo.enviados) != 1 || e.intentos.finalizaciones != 1 {
		t.Fatalf("no debe repetirse: correos %d, finalizaciones %d", len(e.correo.enviados), e.intentos.finalizaciones)
	}
	// Ya no se pueden guardar respuestas ni ver preguntas.
	esperaError(t, e.svc.GuardarRespuesta(ctx, "p1", 3, opc(posicionCorrecta)), domain.ErrConflict, service.MsgEvaluacionTerminada)
	_, err = e.svc.Pregunta(ctx, "p1", 1)
	esperaError(t, err, domain.ErrConflict, service.MsgEvaluacionTerminada)
}

func TestFinalizar_SinRespuestasEsCeroYSinExamenEs404(t *testing.T) {
	ctx := context.Background()
	e := iniciado(t, 6, 4)
	r, err := e.svc.Finalizar(ctx, "p1")
	if err != nil || r.Nota != 0 || r.Aciertos != 0 || r.Total != 4 {
		t.Fatalf("resultado = %+v, %v", r, err)
	}
	otro := nuevoEscenario(6, 3, desplazado{})
	_, err = otro.svc.Finalizar(ctx, "p1")
	esperaError(t, err, domain.ErrNotFound, service.MsgEvaluacionNoIniciada)
}

func TestFinalizar_DespuesDelCierreTerminaASuHoraDeCierre(t *testing.T) {
	ctx := context.Background()
	e := iniciado(t, 6, 4)
	cierre := e.intento().Cierre
	*e.reloj = cierre.Add(3 * time.Hour)
	r, err := e.svc.Finalizar(ctx, "p1")
	if err != nil || !r.FinalizadoEn.Equal(cierre) {
		t.Fatalf("se esperaba la hora de cierre %v, llegó %+v, %v", cierre, r, err)
	}
}

func TestFinalizar_SiOtroProcesoLoFinalizoDevuelveLoGuardadoSinCorreo(t *testing.T) {
	e := iniciado(t, 6, 4)
	e.intentos.otroFinalizo = true
	r, err := e.svc.Finalizar(context.Background(), "p1")
	if err != nil || r.Nota != 12.5 {
		t.Fatalf("debe devolver la nota que quedó guardada: %+v, %v", r, err)
	}
	if len(e.correo.enviados) != 0 {
		t.Fatal("el correo lo envía quien finalizó, no esta llamada")
	}
}

func TestFinalizar_SiFallaElCorreoLaNotaQuedaGuardada(t *testing.T) {
	e := iniciado(t, 6, 4)
	caido := &correoCaido{}
	e.svc.ConMailer(caido, nil)
	r, err := e.svc.Finalizar(context.Background(), "p1")
	if err != nil {
		t.Fatalf("un fallo del correo no debe deshacer la nota: %v", err)
	}
	if caido.intentos != 1 || e.intento().FinalizadoEn == nil || r.Total != 4 {
		t.Fatalf("intentos de envío = %d, intento = %+v", caido.intentos, e.intento())
	}
}

func TestFinalizarVencidos_SoloLosQueYaCerraron(t *testing.T) {
	ctx := context.Background()
	e := iniciado(t, 10, 4)
	e.intentos.postulaciones["p2"] = domain.PostulacionEvaluacion{ID: "post-2", Estado: domain.EstadoEnviada, SuperoMerito: true}
	*e.reloj = inicioExamen.Add(20 * time.Minute) // p2 empieza 20 minutos después
	if _, _, err := e.svc.Iniciar(ctx, "p2"); err != nil {
		t.Fatal(err)
	}

	*e.reloj = inicioExamen.Add(35 * time.Minute) // el de p1 ya cerró (30 min); el de p2 no
	n, err := e.svc.FinalizarVencidos(ctx)
	if err != nil || n != 1 {
		t.Fatalf("finalizados = %d, err = %v", n, err)
	}
	if e.intentos.intentos["post-1"].FinalizadoEn == nil || e.intentos.intentos["post-2"].FinalizadoEn != nil {
		t.Fatal("solo debe finalizar el que ya cerró")
	}
	if len(e.correo.enviados) != 1 {
		t.Fatalf("el reporte se envía al finalizar: %d correos", len(e.correo.enviados))
	}
	// Una segunda vuelta no hace nada más.
	if n, _ := e.svc.FinalizarVencidos(ctx); n != 0 {
		t.Fatalf("segunda vuelta = %d", n)
	}
}

func TestResultados_FinalizaLosVencidosAntesDeListar(t *testing.T) {
	e := iniciado(t, 6, 3)
	*e.reloj = inicioExamen.Add(time.Hour)
	rs, err := e.svc.Resultados(context.Background())
	if err != nil || len(rs) != 1 || rs[0].FinalizadoEn == nil || rs[0].Nota == nil {
		t.Fatalf("resultados = %+v, %v", rs, err)
	}
}

func TestReporte(t *testing.T) {
	ctx := context.Background()

	t.Run("id inválido o inexistente es 404", func(t *testing.T) {
		e := iniciado(t, 6, 3)
		for _, id := range []string{"abc", "", "11111111-2222-3333-4444-555555555555"} {
			_, err := e.svc.Reporte(ctx, id)
			esperaError(t, err, domain.ErrNotFound, service.MsgResultadoNoExiste)
		}
	})
	t.Run("examen en curso es 409", func(t *testing.T) {
		e := iniciado(t, 6, 3)
		_, err := e.svc.Reporte(ctx, e.intento().ID)
		esperaError(t, err, domain.ErrConflict, service.MsgReporteEnCurso)
	})
	t.Run("examen vencido sin finalizar se finaliza y da el PDF", func(t *testing.T) {
		e := iniciado(t, 6, 3)
		*e.reloj = inicioExamen.Add(time.Hour)
		r, err := e.svc.Reporte(ctx, e.intento().ID)
		if err != nil || !bytes.HasPrefix(r.Contenido, []byte("%PDF")) || e.intento().FinalizadoEn == nil {
			t.Fatalf("reporte = %d bytes, err = %v", len(r.Contenido), err)
		}
	})
	t.Run("examen finalizado da el PDF con su nombre", func(t *testing.T) {
		e := iniciado(t, 6, 3)
		if _, err := e.svc.Finalizar(ctx, "p1"); err != nil {
			t.Fatal(err)
		}
		r, err := e.svc.Reporte(ctx, e.intento().ID)
		if err != nil || r.Nombre != "reporte-evaluacion-BIESS-2026-post-1.pdf" || !bytes.HasPrefix(r.Contenido, []byte("%PDF")) {
			t.Fatalf("reporte = %q (%d bytes), err = %v", r.Nombre, len(r.Contenido), err)
		}
	})
}

var paginaPDF = regexp.MustCompile(`/Type\s*/Page[^s]`)

func TestGenerarReportePDF_UnExamenLargoOcupaVariasPaginas(t *testing.T) {
	d := service.DatosReporte{
		Codigo: "BIESS-2026-000001", Nombres: "Ana María", Apellidos: "Núñez", Cedula: "1712345675",
		Cargo: domain.CargoJubilados, Inicio: inicioExamen, Fin: inicioExamen.Add(time.Hour), Emitido: inicioExamen,
		Nota: 75, Aciertos: 45, Total: 60, Verificacion: "0000-0000-0001",
	}
	for i := 1; i <= 60; i++ {
		d.Detalle = append(d.Detalle, service.DetallePregunta{
			Numero: i, Enunciado: strings.Repeat("¿Pregunta larga con tildes y ñ? ", 8), Estado: service.EstadoRespuesta(i % 3),
		})
	}
	pdf, err := service.GenerarReportePDF(d)
	if err != nil || !bytes.HasPrefix(pdf, []byte("%PDF")) {
		t.Fatalf("pdf = %d bytes, err = %v", len(pdf), err)
	}
	if paginas := len(paginaPDF.FindAll(pdf, -1)); paginas < 2 {
		t.Fatalf("60 preguntas no caben en una página: el reporte tiene %d", paginas)
	}
}

func TestFinalizar_CalificaConLasRespuestasLeidasConElExamenBloqueado(t *testing.T) {
	ctx := context.Background()
	e := iniciado(t, 6, 4)
	intento := e.intento()
	// Otra petición confirmó una respuesta correcta justo antes de que finalizar obtuviera el
	// bloqueo: la nota tiene que contarla, y el reporte después debe decir lo mismo.
	e.intentos.alBloquear = func() {
		if e.intentos.respuestas == nil {
			e.intentos.respuestas = map[string]map[int]int{}
		}
		e.intentos.respuestas[intento.ID] = map[int]int{intento.Preguntas[0].PreguntaID: 0}
	}

	r, err := e.svc.Finalizar(ctx, "p1")
	if err != nil || r.Aciertos != 1 || r.Nota != 25 {
		t.Fatalf("la nota debe contar la respuesta confirmada antes del bloqueo: %+v, %v", r, err)
	}
	guardado := e.intento()
	if guardado.Nota == nil || *guardado.Nota != r.Nota {
		t.Fatalf("nota guardada = %v; se esperaba %v", guardado.Nota, r.Nota)
	}
	// Leer de nuevo el resultado (como el reporte y los resultados) da los mismos aciertos.
	otra, err := e.svc.Finalizar(ctx, "p1")
	if err != nil || otra.Aciertos != r.Aciertos || otra.Nota != r.Nota {
		t.Fatalf("el resultado guardado no coincide con la nota: %+v, %v", otra, err)
	}
	if e.intentos.fueraDeTx != 0 {
		t.Fatalf("bloquear, leer las respuestas y finalizar deben ir en la transacción (%d fuera)", e.intentos.fueraDeTx)
	}
}

func TestGuardarRespuesta_UsaUnaTransaccion(t *testing.T) {
	e := iniciado(t, 6, 3)
	antes := e.tx.abiertas
	if err := e.svc.GuardarRespuesta(context.Background(), "p1", 1, opc(posicionCorrecta)); err != nil {
		t.Fatal(err)
	}
	if e.tx.abiertas != antes+1 || e.intentos.fueraDeTx != 0 {
		t.Fatalf("guardar debe abrir una transacción: abiertas %d → %d, fuera de tx %d", antes, e.tx.abiertas, e.intentos.fueraDeTx)
	}
}

func TestFinalizarVencidos_UnoQueFallaNoDetieneALosDemas(t *testing.T) {
	ctx := context.Background()
	e := iniciado(t, 10, 4)
	e.intentos.postulaciones["p2"] = domain.PostulacionEvaluacion{ID: "post-2", Estado: domain.EstadoEnviada, SuperoMerito: true}
	if _, _, err := e.svc.Iniciar(ctx, "p2"); err != nil {
		t.Fatal(err)
	}
	e.intentos.fallaAlBloquear = e.intentos.intentos["post-1"].ID

	*e.reloj = inicioExamen.Add(time.Hour) // los dos ya cerraron
	n, err := e.svc.FinalizarVencidos(ctx)

	if n != 1 || e.intentos.intentos["post-2"].FinalizadoEn == nil {
		t.Fatalf("debe finalizar el que no falla: finalizados = %d", n)
	}
	if err == nil || !strings.Contains(err.Error(), e.intentos.fallaAlBloquear) {
		t.Fatalf("debe informar el intento que falló: %v", err)
	}
}

func TestResultados_ListaAunqueUnVencidoNoSePuedaFinalizar(t *testing.T) {
	e := iniciado(t, 6, 3)
	e.intentos.fallaAlBloquear = e.intento().ID
	*e.reloj = inicioExamen.Add(time.Hour)

	rs, err := e.svc.Resultados(context.Background())

	if err != nil || len(rs) != 1 {
		t.Fatalf("el administrador debe ver la lista aunque un examen no se pueda finalizar: %+v, %v", rs, err)
	}
}
