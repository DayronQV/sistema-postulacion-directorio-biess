package domain

import (
	"context"
	"time"
)

// Límites del banco de preguntas y de la configuración de la evaluación teórica (EF-14).
const (
	MinOpcionesPregunta  = 2
	MaxOpcionesPregunta  = 8
	MaxLargoEnunciado    = 1000
	MaxLargoOpcion       = 500
	MaxPreguntasPorCarga = 500
	MaxErroresReportados = 50
	MinDuracionMinutos   = 1
	MaxDuracionMinutos   = 480
)

// Pregunta es una pregunta de opción múltiple del banco. RespuestaCorrecta es el índice
// (desde 0) de la opción correcta dentro de Opciones.
//
// Esta es la vista del administrador: trae la respuesta correcta. Lo que ve el postulante
// es PreguntaPublica, que no la incluye.
type Pregunta struct {
	ID                int      `json:"id"`
	Enunciado         string   `json:"enunciado"`
	Opciones          []string `json:"opciones"`
	RespuestaCorrecta int      `json:"respuesta_correcta"`
}

// PreguntaPublica es lo único de una pregunta que se envía al postulante: la respuesta
// correcta nunca sale del servidor.
type PreguntaPublica struct {
	ID        int      `json:"id"`
	Enunciado string   `json:"enunciado"`
	Opciones  []string `json:"opciones"`
}

// Publica devuelve la pregunta sin la respuesta correcta.
func (p Pregunta) Publica() PreguntaPublica {
	return PreguntaPublica{ID: p.ID, Enunciado: p.Enunciado, Opciones: p.Opciones}
}

// ConfigEvaluacion son los parámetros de la evaluación: cuándo empieza, cuánto dura y
// cuántas preguntas se sortean por postulante. IniciaEn es nil mientras no se configure.
type ConfigEvaluacion struct {
	IniciaEn        *time.Time
	DuracionMinutos int
	NumPreguntas    int
}

// FiltroPreguntas filtra y pagina el listado del banco. Buscar compara con el enunciado
// sin distinguir mayúsculas; vacío trae todas.
type FiltroPreguntas struct {
	Buscar    string
	Pagina    int
	PorPagina int
}

// BancoEvaluacion es el puerto de persistencia del banco de preguntas y de la
// configuración de la evaluación (tablas eval_preguntas, eval_config y eval_intentos).
type BancoEvaluacion interface {
	// CrearPreguntas guarda todas las preguntas o ninguna, en una transacción, y las
	// devuelve con su id.
	CrearPreguntas(ctx context.Context, preguntas []Pregunta) ([]Pregunta, error)
	// ListarPreguntas devuelve una página del banco (ordenado por id) y el total que
	// coincide con el filtro.
	ListarPreguntas(ctx context.Context, f FiltroPreguntas) ([]Pregunta, int, error)
	// ObtenerPregunta devuelve una pregunta del banco; false si no existe.
	ObtenerPregunta(ctx context.Context, id int) (Pregunta, bool, error)
	// ActualizarPregunta reemplaza el contenido de la pregunta; false si no existe.
	ActualizarPregunta(ctx context.Context, p Pregunta) (bool, error)
	// EliminarPregunta borra la pregunta; false si no existe.
	EliminarPregunta(ctx context.Context, id int) (bool, error)
	// ContarPreguntas devuelve el tamaño del banco.
	ContarPreguntas(ctx context.Context) (int, error)
	// HayIntentos indica si algún postulante ya inició la evaluación.
	HayIntentos(ctx context.Context) (bool, error)
	// ObtenerConfig devuelve la configuración guardada; false si nunca se configuró.
	ObtenerConfig(ctx context.Context) (ConfigEvaluacion, bool, error)
	// GuardarConfig crea o reemplaza la configuración (una sola fila).
	GuardarConfig(ctx context.Context, c ConfigEvaluacion) error
}
