package domain

import (
	"context"
	"time"
)

// PreguntaDelIntento es una pregunta sorteada para un postulante. Orden dice en qué orden
// se le muestran las opciones: Orden[k] es el índice, en la pregunta del banco, de la opción
// que se le presenta en la posición k (desde 0). Al calificar, la opción elegida en la
// posición k vale como la opción Orden[k] del banco.
type PreguntaDelIntento struct {
	PreguntaID int
	Orden      []int
}

// IntentoEvaluacion es el examen de un postulante: qué preguntas le tocaron y en qué orden,
// cuándo empezó y a qué hora se cierra. Hay uno solo por postulación. Las preguntas están
// en el orden en que se le muestran: la posición n de la lista es la pregunta n del examen.
type IntentoEvaluacion struct {
	ID            string
	PostulacionID string
	Preguntas     []PreguntaDelIntento
	Inicio        time.Time
	Cierre        time.Time
	FinalizadoEn  *time.Time
	Nota          *float64
}

// Vigente indica si el intento sigue abierto a la hora dada: no se finalizó y todavía no
// llegó la hora de cierre.
func (i IntentoEvaluacion) Vigente(ahora time.Time) bool {
	return i.FinalizadoEn == nil && ahora.Before(i.Cierre)
}

// PostulacionEvaluacion es lo que importa de una postulación para decidir si su dueño puede
// rendir la evaluación.
type PostulacionEvaluacion struct {
	ID           string
	Estado       string
	SuperoMerito bool
}

// PreguntaSorteable es lo único que el sorteo necesita saber de una pregunta del banco.
type PreguntaSorteable struct {
	ID          int
	NumOpciones int
}

// IntentosEvaluacion es el puerto de persistencia de los intentos (tabla eval_intentos).
type IntentosEvaluacion interface {
	// PreguntasParaSorteo devuelve todas las preguntas del banco (id y cantidad de opciones),
	// ordenadas por id.
	PreguntasParaSorteo(ctx context.Context) ([]PreguntaSorteable, error)
	// PostulacionDelPostulante devuelve la postulación del postulante; false si no tiene.
	PostulacionDelPostulante(ctx context.Context, postulanteID string) (PostulacionEvaluacion, bool, error)
	// CrearIntento guarda el intento solo si la postulación todavía no tiene uno. Devuelve
	// el intento que quedó guardado y true si lo creó esta llamada; si ya existía, devuelve
	// el existente y false. Dos llamadas simultáneas nunca crean dos.
	CrearIntento(ctx context.Context, i IntentoEvaluacion) (IntentoEvaluacion, bool, error)
	// ObtenerIntento devuelve el intento de la postulación; false si no hay.
	ObtenerIntento(ctx context.Context, postulacionID string) (IntentoEvaluacion, bool, error)
}

// Barajador produce los órdenes aleatorios del sorteo. La implementación real usa una fuente
// criptográfica; en las pruebas se reemplaza por una determinista.
type Barajador interface {
	// Permutacion devuelve los números 0..n-1 en un orden aleatorio.
	Permutacion(n int) ([]int, error)
}
