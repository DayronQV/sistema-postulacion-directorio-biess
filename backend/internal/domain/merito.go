package domain

import "context"

// Estados que ve el postulante. Coinciden con los de la base, salvo "Evaluación Técnica":
// una postulación enviada que el administrador marcó como que superó la Fase de Méritos
// (EF-14: solo esos postulantes acceden a la evaluación teórica).
const (
	EstadoVisibleBorrador          = "Borrador"
	EstadoVisibleEnviada           = "Enviada"
	EstadoVisibleCerrada           = "Cerrada"
	EstadoVisibleEvaluacionTecnica = "Evaluación Técnica"
)

// EstadoVisible es el estado de la postulación como lo ve el postulante.
func EstadoVisible(estado string, superoMerito bool) string {
	switch {
	case estado == EstadoEnviada && superoMerito:
		return EstadoVisibleEvaluacionTecnica
	case estado == EstadoEnviada:
		return EstadoVisibleEnviada
	case estado == EstadoCerrada:
		return EstadoVisibleCerrada
	default:
		return EstadoVisibleBorrador
	}
}

// Meritos es el puerto para marcar en postulaciones.supero_merito si el postulante superó
// la Fase de Méritos.
type Meritos interface {
	// EstadoPorID devuelve el estado de la postulación, o ErrNotFound si no existe.
	EstadoPorID(ctx context.Context, postulacionID string) (string, error)
	// ActualizarMerito guarda la marca solo si la postulación está enviada. Devuelve false
	// si no la actualizó (porque no estaba enviada al momento de escribir).
	ActualizarMerito(ctx context.Context, postulacionID string, supero bool) (bool, error)
}
