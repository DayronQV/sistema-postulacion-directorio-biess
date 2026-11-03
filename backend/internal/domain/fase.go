package domain

import (
	"context"
	"time"
)

// Nombres de las fases del concurso (tabla fases; las siembra db/schema.sql).
const (
	FasePostulacion = "Postulación"
	FaseMeritos     = "Méritos"
	FaseEvaluacion  = "Evaluación"
)

// MsgPeriodoCerrado es el 403 de crear, guardar o enviar una postulación fuera del periodo.
const MsgPeriodoCerrado = "El periodo de postulación no está habilitado"

// Fase es una fase del concurso que el administrador publica (EF-13). Solo la fase
// Postulación habilita el formulario (EF-12). Las fechas son opcionales: sin fecha no
// hay límite por ese lado.
type Fase struct {
	ID                 int
	Nombre             string
	Orden              int
	Publicada          bool
	HabilitaFormulario bool
	FechaInicio        *time.Time
	FechaFin           *time.Time
}

// FormularioAbierto indica si con esta fase se puede crear, guardar o enviar una
// postulación en el instante ahora: publicada, con el formulario habilitado y dentro de
// sus fechas (ambos extremos incluidos).
func (f Fase) FormularioAbierto(ahora time.Time) bool {
	if !f.Publicada || !f.HabilitaFormulario {
		return false
	}
	if f.FechaInicio != nil && ahora.Before(*f.FechaInicio) {
		return false
	}
	if f.FechaFin != nil && ahora.After(*f.FechaFin) {
		return false
	}
	return true
}

// Fases es el puerto de la tabla fases.
type Fases interface {
	// Listar devuelve las fases en su orden.
	Listar(ctx context.Context) ([]Fase, error)
	// Obtener devuelve la fase, o un error que cumple errors.Is(err, ErrNotFound).
	Obtener(ctx context.Context, id int) (Fase, error)
	// PorNombre devuelve la fase, o un error que cumple errors.Is(err, ErrNotFound).
	PorNombre(ctx context.Context, nombre string) (Fase, error)
	// Actualizar guarda publicada, habilita_formulario y las fechas de la fase.
	Actualizar(ctx context.Context, f Fase) error
}
