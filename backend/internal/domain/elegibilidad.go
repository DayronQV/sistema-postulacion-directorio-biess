package domain

// Sección Inhabilidades (issue #33): 13 preguntas Sí/No. La pregunta 1 pide una condición
// que se debe cumplir (Sí) y de la 2 a la 13 son impedimentos (No).
const (
	TotalPreguntasInhabilidades = 13
	PreguntaRemovido            = 7 // removido por un organismo de control en los últimos 5 años

	MaxLongitudInstitucionRemovido = 150
	MaxLongitudCargoRemovido       = 150
	MaxLongitudMotivoRemovido      = 500
)

// RemovidoDetalle es el detalle que se pide cuando la pregunta 7 se responde Sí.
type RemovidoDetalle struct {
	Institucion string
	Cargo       string
	FechaFin    string // AAAA-MM-DD; opcional
	Motivo      string
}

// Inhabilidades son las respuestas del postulante. Una pregunta sin responder no tiene
// entrada en Respuestas: en borrador se puede guardar la sección a medias.
type Inhabilidades struct {
	Respuestas map[int]bool
	Removido   RemovidoDetalle // solo tiene sentido si Respuestas[PreguntaRemovido] es true
}

// Completa indica si las 13 preguntas tienen respuesta.
func (i Inhabilidades) Completa() bool {
	for n := 1; n <= TotalPreguntasInhabilidades; n++ {
		if _, ok := i.Respuestas[n]; !ok {
			return false
		}
	}
	return true
}

// Elegible aplica la regla del requerimiento: la pregunta 1 en Sí y de la 2 a la 13 en No.
// Mientras falte alguna respuesta no hay elegibilidad que afirmar, así que devuelve false.
// No descalifica a nadie: el analista revisa los casos en la fase de méritos (EF-13).
func (i Inhabilidades) Elegible() bool {
	if !i.Completa() {
		return false
	}
	for n := 1; n <= TotalPreguntasInhabilidades; n++ {
		if i.Respuestas[n] != (n == 1) {
			return false
		}
	}
	return true
}
