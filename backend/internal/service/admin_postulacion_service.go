package service

import (
	"context"
	"regexp"
	"strings"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/domain"
)

// MsgMeritoSoloEnviada es el 409 de marcar el mérito de una postulación no enviada.
const MsgMeritoSoloEnviada = "Solo se puede marcar la Fase de Méritos de una postulación enviada"

// Un id de postulación es un UUID; con otro formato no hay nada que buscar (404).
var uuidRe = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

// ResultadoMerito es la postulación después de marcar su mérito.
type ResultadoMerito struct {
	ID            string
	SuperoMerito  bool
	EstadoVisible string // lo que ve el postulante: "Evaluación Técnica" si superó
}

// AdminPostulacionService reúne las acciones del administrador sobre las postulaciones.
type AdminPostulacionService struct {
	meritos domain.Meritos
}

// NewAdminPostulacionService crea el servicio con sus dependencias inyectadas.
func NewAdminPostulacionService(meritos domain.Meritos) *AdminPostulacionService {
	return &AdminPostulacionService{meritos: meritos}
}

// MarcarMerito registra si el postulante superó la Fase de Méritos (EF-13). Solo se marca
// una postulación enviada: 404 si no existe y 409 si está en borrador o cerrada. Que solo
// el administrador pueda hacerlo lo controla la ruta (grupo de escritura del panel).
func (s *AdminPostulacionService) MarcarMerito(ctx context.Context, postulacionID string, supero bool) (ResultadoMerito, error) {
	id := strings.TrimSpace(postulacionID)
	if !uuidRe.MatchString(id) {
		return ResultadoMerito{}, domain.NotFound("La postulación no existe")
	}
	estado, err := s.meritos.EstadoPorID(ctx, id)
	if err != nil {
		return ResultadoMerito{}, err
	}
	if estado != domain.EstadoEnviada {
		return ResultadoMerito{}, domain.Conflict(MsgMeritoSoloEnviada, nil)
	}
	actualizada, err := s.meritos.ActualizarMerito(ctx, id, supero)
	if err != nil {
		return ResultadoMerito{}, err
	}
	if !actualizada {
		// Cambió de estado entre la lectura y la escritura.
		return ResultadoMerito{}, domain.Conflict(MsgMeritoSoloEnviada, nil)
	}
	return ResultadoMerito{
		ID:            id,
		SuperoMerito:  supero,
		EstadoVisible: domain.EstadoVisible(domain.EstadoEnviada, supero),
	}, nil
}
