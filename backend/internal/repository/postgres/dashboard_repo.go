package postgres

import (
	"context"
	"fmt"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/domain"
)

var _ domain.DashboardReader = (*PostulacionRepo)(nil)

// ObtenerDashboard calcula todos los indicadores en una sola consulta agregada.
func (r *PostulacionRepo) ObtenerDashboard(ctx context.Context) (domain.DashboardPostulaciones, error) {
	var d domain.DashboardPostulaciones
	err := r.db.QueryRow(ctx, `
SELECT COUNT(*)::int,
       COUNT(*) FILTER (WHERE cargo = 'afiliados')::int,
       COUNT(*) FILTER (WHERE cargo = 'jubilados')::int,
       COUNT(*) FILTER (WHERE cargo = 'empleador')::int,
       COUNT(*) FILTER (WHERE estado = 'borrador')::int,
       COUNT(*) FILTER (WHERE estado = 'enviada')::int,
       COUNT(*) FILTER (WHERE estado = 'cerrada')::int
FROM postulaciones`).Scan(
		&d.Total,
		&d.PorCargo.Afiliados,
		&d.PorCargo.Jubilados,
		&d.PorCargo.Empleador,
		&d.PorEstado.Borrador,
		&d.PorEstado.Enviada,
		&d.PorEstado.Cerrada,
	)
	if err != nil {
		return domain.DashboardPostulaciones{}, fmt.Errorf("consultando los indicadores del dashboard: %w", err)
	}
	return d, nil
}
