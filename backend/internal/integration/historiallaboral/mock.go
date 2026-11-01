// Package historiallaboral contiene los adaptadores del puerto domain.HistorialLaboral.
// Hoy solo existe el simulador (Mock): el Historial Laboral real del IESS no tiene
// acceso público para este proyecto.
package historiallaboral

import (
	"context"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/domain"
)

// Cédulas de prueba. Son las mismas del simulador del Registro Civil
// (internal/integration/registrocivil), que las valida con una prueba.
const (
	CedulaAfiliado  = "1712345675"
	CedulaJubilado  = "0912345675"
	CedulaEmpleador = "1312345679"
	CedulaNinguna   = "1709876542"
)

// Mock es un Historial Laboral simulado con casos de prueba fijos.
type Mock struct {
	calidades map[string]domain.Calidad
}

var _ domain.HistorialLaboral = (*Mock)(nil)

// NewMock crea el simulador con las cuatro cédulas de prueba.
func NewMock() *Mock {
	return &Mock{calidades: map[string]domain.Calidad{
		CedulaAfiliado:  {EsAfiliado: true},
		CedulaJubilado:  {EsJubilado: true},
		CedulaEmpleador: {EsEmpleador: true},
		CedulaNinguna:   {},
	}}
}

// ConsultarCalidad devuelve la calidad de la cédula de prueba. Cualquier otra
// cédula se trata como afiliada, para que las cédulas nuevas puedan postular al
// sector de Afiliados sin tocar el simulador.
func (m *Mock) ConsultarCalidad(ctx context.Context, cedula string) (domain.Calidad, error) {
	if err := ctx.Err(); err != nil {
		return domain.Calidad{}, err
	}
	if c, ok := m.calidades[cedula]; ok {
		return c, nil
	}
	return domain.Calidad{EsAfiliado: true}, nil
}
