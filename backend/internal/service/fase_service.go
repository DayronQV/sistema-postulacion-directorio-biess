package service

import (
	"context"
	"errors"
	"time"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/domain"
)

// CambiosFase son los campos que el administrador cambia de una fase.
type CambiosFase struct {
	Publicada          bool
	HabilitaFormulario bool
	FechaInicio        *time.Time
	FechaFin           *time.Time
}

// FaseService gestiona las fases del concurso (EF-13): publicarlas, fijar sus fechas y
// habilitar el formulario de postulación.
type FaseService struct {
	fases domain.Fases
}

// NewFaseService crea el servicio con sus dependencias inyectadas.
func NewFaseService(fases domain.Fases) *FaseService {
	return &FaseService{fases: fases}
}

// Listar devuelve las fases en su orden.
func (s *FaseService) Listar(ctx context.Context) ([]domain.Fase, error) {
	return s.fases.Listar(ctx)
}

// Actualizar publica o despublica la fase y fija sus fechas. Solo la fase Postulación
// puede habilitar el formulario, y solo si está publicada; la fecha de fin no puede ser
// anterior a la de inicio.
func (s *FaseService) Actualizar(ctx context.Context, id int, c CambiosFase) (domain.Fase, error) {
	f, err := s.fases.Obtener(ctx, id)
	if err != nil {
		return domain.Fase{}, err
	}
	fields := map[string]string{}
	if c.HabilitaFormulario && f.Nombre != domain.FasePostulacion {
		fields["habilita_formulario"] = "Solo la fase de Postulación habilita el formulario"
	} else if c.HabilitaFormulario && !c.Publicada {
		fields["habilita_formulario"] = "Publique la fase para habilitar el formulario"
	}
	if c.FechaInicio != nil && c.FechaFin != nil && c.FechaFin.Before(*c.FechaInicio) {
		fields["fecha_fin"] = "La fecha de fin no puede ser anterior a la de inicio"
	}
	if len(fields) > 0 {
		return domain.Fase{}, domain.Validation("Datos inválidos", fields)
	}

	f.Publicada = c.Publicada
	f.HabilitaFormulario = c.HabilitaFormulario
	f.FechaInicio = c.FechaInicio
	f.FechaFin = c.FechaFin
	if err := s.fases.Actualizar(ctx, f); err != nil {
		return domain.Fase{}, err
	}
	return f, nil
}

// formularioAbierto indica si en el instante ahora se puede crear, guardar o enviar una
// postulación. Sin fases configuradas, o sin la fase Postulación en la base, no hay
// restricción.
func formularioAbierto(ctx context.Context, fases domain.Fases, ahora time.Time) (bool, error) {
	if fases == nil {
		return true, nil
	}
	f, err := fases.PorNombre(ctx, domain.FasePostulacion)
	if errors.Is(err, domain.ErrNotFound) {
		return true, nil
	}
	if err != nil {
		return false, err
	}
	return f.FormularioAbierto(ahora), nil
}
