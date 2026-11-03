package domain

import (
	"context"
	"time"
)

// ResultadoEvaluacion es el examen de un postulante con los datos que necesitan el
// administrador y el reporte PDF: quién es, cuándo lo rindió y qué nota obtuvo. Nota y
// FinalizadoEn son nil mientras el examen sigue en curso.
type ResultadoEvaluacion struct {
	IntentoID      string
	PostulacionID  string
	Codigo         string
	Cargo          string // vacío si la postulación no eligió cargo
	Nombres        string
	Apellidos      string
	Cedula         string
	Email          string
	Inicio         time.Time
	Cierre         time.Time
	FinalizadoEn   *time.Time
	Nota           *float64
	TotalPreguntas int
}

// RespuestasEvaluacion es el puerto de persistencia de las respuestas, de la calificación y
// de los resultados de la evaluación (tablas eval_respuestas y eval_intentos).
//
// Guardar una respuesta y finalizar el intento se ordenan con bloqueos sobre la fila del
// intento, dentro de la transacción abierta por TxManager: GuardarRespuesta la bloquea en modo
// compartido y BloquearParaFinalizar en modo exclusivo. Así una respuesta que llega mientras se
// califica entra antes de calificar o se rechaza, y la nota coincide siempre con las respuestas
// guardadas.
type RespuestasEvaluacion interface {
	// GuardarRespuesta guarda (o reemplaza), dentro de tx, la opción elegida para una pregunta
	// del intento, como índice de la opción en el banco. Antes bloquea el intento en modo
	// compartido y solo guarda si a la hora dada sigue abierto: devuelve false si ya terminó o
	// venció (también si lo finalizaron mientras esperaba el bloqueo).
	GuardarRespuesta(ctx context.Context, tx Tx, intentoID string, preguntaID, opcion int, ahora time.Time) (bool, error)
	// Respuestas devuelve las respuestas guardadas del intento: id de pregunta → índice de la
	// opción elegida en el banco. Es una lectura fuera de transacción.
	Respuestas(ctx context.Context, intentoID string) (map[int]int, error)
	// BloquearParaFinalizar bloquea el intento dentro de tx (SELECT ... FOR UPDATE) hasta el
	// COMMIT o ROLLBACK, si todavía no está finalizado. Devuelve false si ya lo estaba (por
	// ejemplo, porque otro proceso lo finalizó mientras esta esperaba el bloqueo).
	BloquearParaFinalizar(ctx context.Context, tx Tx, intentoID string) (bool, error)
	// RespuestasEnTx es Respuestas dentro de tx, para calificar con el intento ya bloqueado.
	RespuestasEnTx(ctx context.Context, tx Tx, intentoID string) (map[int]int, error)
	// Finalizar guarda, dentro de tx, la nota y la hora de finalización si el intento todavía no
	// estaba finalizado. Devuelve false si otro proceso lo finalizó antes.
	Finalizar(ctx context.Context, tx Tx, intentoID string, nota float64, finalizadoEn time.Time) (bool, error)
	// VencidosSinFinalizar devuelve los intentos cuya hora de cierre ya pasó y que nadie
	// finalizó todavía.
	VencidosSinFinalizar(ctx context.Context, ahora time.Time) ([]IntentoEvaluacion, error)
	// IntentoPorID devuelve un intento por su id; false si no existe.
	IntentoPorID(ctx context.Context, intentoID string) (IntentoEvaluacion, bool, error)
	// Resultados devuelve todos los exámenes con los datos del postulante, ordenados por
	// apellidos y nombres.
	Resultados(ctx context.Context) ([]ResultadoEvaluacion, error)
	// ResultadoDelIntento devuelve el resultado de un intento; false si no existe.
	ResultadoDelIntento(ctx context.Context, intentoID string) (ResultadoEvaluacion, bool, error)
}
