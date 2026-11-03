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
type RespuestasEvaluacion interface {
	// GuardarRespuesta guarda (o reemplaza) la opción elegida para una pregunta del intento,
	// como índice de la opción en el banco. Solo la guarda si a la hora dada el intento sigue
	// abierto: devuelve false si ya terminó o venció.
	GuardarRespuesta(ctx context.Context, intentoID string, preguntaID, opcion int, ahora time.Time) (bool, error)
	// Respuestas devuelve las respuestas guardadas del intento: id de pregunta → índice de la
	// opción elegida en el banco.
	Respuestas(ctx context.Context, intentoID string) (map[int]int, error)
	// Finalizar guarda la nota y la hora de finalización si el intento todavía no estaba
	// finalizado. Devuelve false si otro proceso lo finalizó antes.
	Finalizar(ctx context.Context, intentoID string, nota float64, finalizadoEn time.Time) (bool, error)
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
