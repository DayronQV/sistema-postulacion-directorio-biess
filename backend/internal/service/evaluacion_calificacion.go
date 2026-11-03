package service

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/domain"
)

// Mensajes de las respuestas, la calificación y los resultados.
const (
	MsgOpcionObligatoria  = "Campo obligatorio"
	MsgReporteEnCurso     = "La evaluación de este postulante todavía está en curso"
	MsgResultadoNoExiste  = "El resultado no existe"
	asuntoCorreoResultado = "Resultado de su evaluación teórica — Concurso del Directorio del BIESS"
)

// ResultadoFinal es lo que recibe el postulante al terminar su examen.
type ResultadoFinal struct {
	Nota         float64
	Aciertos     int
	Total        int
	FinalizadoEn time.Time
}

// ReportePDF es el reporte del resultado listo para descargar.
type ReportePDF struct {
	Nombre    string
	Contenido []byte
}

// GuardarRespuesta guarda la opción que el postulante eligió en la pregunta n (desde 1).
// opcion es la posición de la opción tal como se le mostró (desde 0); se guarda como el
// índice de esa opción en el banco, para calificar sin depender del orden de cada examen.
// Se puede cambiar la respuesta las veces que quiera mientras el examen siga abierto; después
// de la hora de cierre es 409.
func (s *EvaluacionService) GuardarRespuesta(ctx context.Context, postulanteID string, n int, opcion *int) error {
	intento, err := s.intentoDelPostulante(ctx, postulanteID)
	if err != nil {
		return err
	}
	ahora := s.ahora().Truncate(time.Microsecond)
	if !intento.Vigente(ahora) {
		if err := s.cerrarSiVencio(ctx, intento, ahora); err != nil {
			return err
		}
		return domain.Conflict(MsgEvaluacionTerminada, nil)
	}
	if n < 1 || n > len(intento.Preguntas) {
		return domain.NotFound(MsgPreguntaFueraDeRango)
	}
	sorteada := intento.Preguntas[n-1]
	switch {
	case opcion == nil:
		return domain.Validation("Datos inválidos", map[string]string{"opcion": MsgOpcionObligatoria})
	case *opcion < 0 || *opcion >= len(sorteada.Orden):
		return domain.Validation("Datos inválidos", map[string]string{
			"opcion": fmt.Sprintf("Debe ser una de las opciones de la pregunta, de 0 a %d", len(sorteada.Orden)-1),
		})
	}
	// En una transacción: el repositorio bloquea el intento en modo compartido antes de
	// guardar, así una respuesta no se cuela mientras otro proceso califica el examen.
	var guardada bool
	err = s.tx.Ejecutar(ctx, func(ctx context.Context, tx domain.Tx) error {
		var err error
		guardada, err = s.respuestas.GuardarRespuesta(ctx, tx, intento.ID, sorteada.PreguntaID, sorteada.Orden[*opcion], ahora)
		return err
	})
	if err != nil {
		return err
	}
	if !guardada {
		// Se cerró entre la lectura y la escritura.
		return domain.Conflict(MsgEvaluacionTerminada, nil)
	}
	return nil
}

// Finalizar termina el examen del postulante y lo califica con las respuestas guardadas. Si
// ya estaba finalizado (por él o por el cierre automático), devuelve el mismo resultado sin
// volver a enviar el correo.
func (s *EvaluacionService) Finalizar(ctx context.Context, postulanteID string) (ResultadoFinal, error) {
	intento, err := s.intentoDelPostulante(ctx, postulanteID)
	if err != nil {
		return ResultadoFinal{}, err
	}
	if intento.FinalizadoEn != nil {
		return s.resultadoGuardado(ctx, intento)
	}
	return s.finalizarIntento(ctx, intento, s.ahora().Truncate(time.Microsecond))
}

// FinalizarVencidos finaliza los exámenes cuya hora de cierre ya pasó y que nadie finalizó,
// con las respuestas guardadas hasta ese momento. Lo llama una tarea periódica y también se
// hace antes de mostrar los resultados al administrador. Devuelve cuántos finalizó.
//
// Un examen que falla no detiene a los demás: se sigue con el resto y los errores se
// devuelven juntos. Si no, un solo examen con un problema permanente dejaría a todos los
// siguientes sin finalizar en cada vuelta de la tarea.
func (s *EvaluacionService) FinalizarVencidos(ctx context.Context) (int, error) {
	ahora := s.ahora().Truncate(time.Microsecond)
	vencidos, err := s.respuestas.VencidosSinFinalizar(ctx, ahora)
	if err != nil {
		return 0, err
	}
	finalizados := 0
	var fallos []error
	for _, intento := range vencidos {
		if err := ctx.Err(); err != nil {
			fallos = append(fallos, err)
			break
		}
		if _, err := s.finalizarIntento(ctx, intento, ahora); err != nil {
			fallos = append(fallos, fmt.Errorf("finalizando el intento %s: %w", intento.ID, err))
			continue
		}
		finalizados++
	}
	return finalizados, errors.Join(fallos...)
}

// Resultados devuelve los exámenes de todos los postulantes para el administrador. Antes
// finaliza los que vencieron, para que ninguno aparezca "en curso" cuando ya terminó. Si
// alguno no se pudo finalizar, se registra y se muestran igual los resultados: ese examen
// queda "en curso" hasta la próxima vuelta, en lugar de dejar al administrador sin la lista.
func (s *EvaluacionService) Resultados(ctx context.Context) ([]domain.ResultadoEvaluacion, error) {
	if _, err := s.FinalizarVencidos(ctx); err != nil {
		if ctx.Err() != nil {
			return nil, err
		}
		if s.logger != nil {
			s.logger.Error("no se pudieron finalizar algunos exámenes vencidos", "error", err)
		}
	}
	return s.respuestas.Resultados(ctx)
}

// Reporte genera el PDF del resultado de un examen para el administrador. Un examen que
// sigue en curso todavía no tiene reporte (409).
func (s *EvaluacionService) Reporte(ctx context.Context, intentoID string) (ReportePDF, error) {
	if !uuidRe.MatchString(intentoID) {
		return ReportePDF{}, domain.NotFound(MsgResultadoNoExiste)
	}
	intento, existe, err := s.respuestas.IntentoPorID(ctx, intentoID)
	if err != nil {
		return ReportePDF{}, err
	}
	if !existe {
		return ReportePDF{}, domain.NotFound(MsgResultadoNoExiste)
	}
	if intento.FinalizadoEn == nil {
		ahora := s.ahora().Truncate(time.Microsecond)
		if intento.Vigente(ahora) {
			return ReportePDF{}, domain.Conflict(MsgReporteEnCurso, nil)
		}
		if _, err := s.finalizarIntento(ctx, intento, ahora); err != nil {
			return ReportePDF{}, err
		}
	}
	res, existe, err := s.respuestas.ResultadoDelIntento(ctx, intentoID)
	if err != nil {
		return ReportePDF{}, err
	}
	if !existe {
		return ReportePDF{}, domain.NotFound(MsgResultadoNoExiste)
	}
	detalle, aciertos, err := s.calificar(ctx, intento)
	if err != nil {
		return ReportePDF{}, err
	}
	pdf, err := GenerarReportePDF(s.datosReporte(res, detalle, aciertos))
	if err != nil {
		return ReportePDF{}, err
	}
	return ReportePDF{Nombre: nombreReporte(res.Codigo), Contenido: pdf}, nil
}

// intentoDelPostulante devuelve el examen del postulante, o 404 si todavía no lo inició.
func (s *EvaluacionService) intentoDelPostulante(ctx context.Context, postulanteID string) (domain.IntentoEvaluacion, error) {
	post, ok, err := s.intentos.PostulacionDelPostulante(ctx, postulanteID)
	if err != nil {
		return domain.IntentoEvaluacion{}, err
	}
	if !ok {
		return domain.IntentoEvaluacion{}, domain.NotFound(MsgEvaluacionNoIniciada)
	}
	intento, hay, err := s.intentos.ObtenerIntento(ctx, post.ID)
	if err != nil {
		return domain.IntentoEvaluacion{}, err
	}
	if !hay {
		return domain.IntentoEvaluacion{}, domain.NotFound(MsgEvaluacionNoIniciada)
	}
	return intento, nil
}

// cerrarSiVencio finaliza el examen si ya pasó su hora de cierre y nadie lo había finalizado:
// es el cierre automático "al consultar".
func (s *EvaluacionService) cerrarSiVencio(ctx context.Context, intento domain.IntentoEvaluacion, ahora time.Time) error {
	if intento.FinalizadoEn != nil || ahora.Before(intento.Cierre) {
		return nil
	}
	_, err := s.finalizarIntento(ctx, intento, ahora)
	return err
}

// finalizarIntento califica el examen, guarda la nota y envía el reporte por correo. Si otro
// proceso lo finalizó antes, devuelve lo que quedó guardado sin enviar el correo otra vez.
//
// Bloquear el intento, leer sus respuestas, calificar y guardar la nota van en una sola
// transacción: una respuesta que llega a la vez queda guardada antes de calificar o se
// rechaza (ver GuardarRespuesta), así la nota siempre coincide con las respuestas guardadas
// que después muestran el reporte y los resultados. El correo se envía tras el COMMIT.
func (s *EvaluacionService) finalizarIntento(ctx context.Context, intento domain.IntentoEvaluacion, ahora time.Time) (ResultadoFinal, error) {
	total := len(intento.Preguntas)
	// Un examen vencido termina a su hora de cierre, aunque se finalice después.
	fin := ahora
	if fin.After(intento.Cierre) {
		fin = intento.Cierre
	}
	var (
		detalle    []DetallePregunta
		aciertos   int
		nota       float64
		finalizado bool
	)
	err := s.tx.Ejecutar(ctx, func(ctx context.Context, tx domain.Tx) error {
		abierto, err := s.respuestas.BloquearParaFinalizar(ctx, tx, intento.ID)
		if err != nil || !abierto {
			return err
		}
		respuestas, err := s.respuestas.RespuestasEnTx(ctx, tx, intento.ID)
		if err != nil {
			return err
		}
		detalle, aciertos, err = s.calificarCon(ctx, intento, respuestas)
		if err != nil {
			return err
		}
		nota = notaSobre100(aciertos, total)
		finalizado, err = s.respuestas.Finalizar(ctx, tx, intento.ID, nota, fin)
		return err
	})
	if err != nil {
		return ResultadoFinal{}, err
	}
	if !finalizado {
		guardado, existe, err := s.respuestas.IntentoPorID(ctx, intento.ID)
		if err != nil {
			return ResultadoFinal{}, err
		}
		if !existe {
			return ResultadoFinal{}, fmt.Errorf("el intento %s desapareció mientras se finalizaba", intento.ID)
		}
		return s.resultadoGuardado(ctx, guardado)
	}
	s.enviarReporte(ctx, intento.ID, detalle, aciertos)
	return ResultadoFinal{Nota: nota, Aciertos: aciertos, Total: total, FinalizadoEn: fin}, nil
}

// resultadoGuardado arma el resultado de un examen que ya estaba finalizado.
func (s *EvaluacionService) resultadoGuardado(ctx context.Context, intento domain.IntentoEvaluacion) (ResultadoFinal, error) {
	_, aciertos, err := s.calificar(ctx, intento)
	if err != nil {
		return ResultadoFinal{}, err
	}
	r := ResultadoFinal{Aciertos: aciertos, Total: len(intento.Preguntas)}
	if intento.Nota != nil {
		r.Nota = *intento.Nota
	}
	if intento.FinalizadoEn != nil {
		r.FinalizadoEn = *intento.FinalizadoEn
	}
	return r, nil
}

// calificar lee las respuestas guardadas del intento y lo califica (ver calificarCon).
func (s *EvaluacionService) calificar(ctx context.Context, intento domain.IntentoEvaluacion) ([]DetallePregunta, int, error) {
	respuestas, err := s.respuestas.Respuestas(ctx, intento.ID)
	if err != nil {
		return nil, 0, err
	}
	return s.calificarCon(ctx, intento, respuestas)
}

// calificarCon compara cada respuesta con la respuesta correcta del banco. Devuelve el
// detalle por pregunta, en el orden del examen, y cuántas acertó.
func (s *EvaluacionService) calificarCon(ctx context.Context, intento domain.IntentoEvaluacion, respuestas map[int]int) ([]DetallePregunta, int, error) {
	detalle := make([]DetallePregunta, len(intento.Preguntas))
	aciertos := 0
	for i, sorteada := range intento.Preguntas {
		p, existe, err := s.banco.ObtenerPregunta(ctx, sorteada.PreguntaID)
		if err != nil {
			return nil, 0, err
		}
		if !existe {
			return nil, 0, fmt.Errorf("la pregunta %d del intento %s ya no está en el banco", sorteada.PreguntaID, intento.ID)
		}
		d := DetallePregunta{Numero: i + 1, Enunciado: p.Enunciado, Estado: SinResponder}
		if elegida, respondio := respuestas[sorteada.PreguntaID]; respondio {
			d.Estado = Incorrecta
			if elegida == p.RespuestaCorrecta {
				d.Estado = Correcta
				aciertos++
			}
		}
		detalle[i] = d
	}
	return detalle, aciertos, nil
}

// notaSobre100 es el porcentaje de aciertos con dos decimales; sin preguntas es 0.
func notaSobre100(aciertos, total int) float64 {
	if total == 0 {
		return 0
	}
	return math.Round(float64(aciertos)*10000/float64(total)) / 100
}

// enviarReporte genera el PDF y lo manda al correo del postulante. Un fallo aquí no deshace
// la nota, que ya quedó guardada: se registra en el log y el administrador puede descargar el
// reporte desde los resultados.
func (s *EvaluacionService) enviarReporte(ctx context.Context, intentoID string, detalle []DetallePregunta, aciertos int) {
	if s.mailer == nil {
		return
	}
	registrar := func(msg string, err error) {
		if s.logger != nil {
			s.logger.Error(msg, "intento", intentoID, "error", err)
		}
	}
	res, existe, err := s.respuestas.ResultadoDelIntento(ctx, intentoID)
	if err != nil || !existe {
		registrar("no se pudo leer el resultado para el reporte", err)
		return
	}
	pdf, err := GenerarReportePDF(s.datosReporte(res, detalle, aciertos))
	if err != nil {
		registrar("no se pudo generar el reporte PDF", err)
		return
	}
	nota := 0.0
	if res.Nota != nil {
		nota = *res.Nota
	}
	err = s.mailer.Enviar(ctx, domain.Mensaje{
		Para:   res.Email,
		Asunto: asuntoCorreoResultado,
		Texto: fmt.Sprintf("Estimado/a %s %s:\n\nSu evaluación teórica de la postulación %s fue calificada con %s sobre 100 (%d de %d respuestas correctas).\n\nAdjuntamos el reporte con el detalle.\n\nSistema de postulación al Directorio del BIESS",
			res.Nombres, res.Apellidos, res.Codigo, formatoNota(nota), aciertos, res.TotalPreguntas),
		Adjuntos: []domain.Adjunto{{Nombre: nombreReporte(res.Codigo), Contenido: pdf}},
	})
	if err != nil {
		registrar("no se pudo enviar el reporte por correo", err)
	}
}

// datosReporte arma los datos del reporte a partir del resultado guardado.
func (s *EvaluacionService) datosReporte(res domain.ResultadoEvaluacion, detalle []DetallePregunta, aciertos int) DatosReporte {
	d := DatosReporte{
		Codigo: res.Codigo, Nombres: res.Nombres, Apellidos: res.Apellidos, Cedula: res.Cedula, Cargo: res.Cargo,
		Inicio: res.Inicio, Fin: res.Cierre, Emitido: s.ahora(),
		Aciertos: aciertos, Total: res.TotalPreguntas, Detalle: detalle,
		Verificacion: codigoVerificacion(res.IntentoID),
	}
	if res.FinalizadoEn != nil {
		d.Fin = *res.FinalizadoEn
	}
	if res.Nota != nil {
		d.Nota = *res.Nota
	}
	return d
}
