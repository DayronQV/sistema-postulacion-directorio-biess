package service

import (
	"context"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/domain"
)

// DashboardService obtiene los indicadores de postulaciones para el panel.
type DashboardService struct {
	dashboard domain.DashboardReader
}

// NewDashboardService crea el servicio con su repositorio inyectado.
func NewDashboardService(dashboard domain.DashboardReader) *DashboardService {
	return &DashboardService{dashboard: dashboard}
}

// Obtener devuelve los conteos actuales por cargo y estado.
func (s *DashboardService) Obtener(ctx context.Context) (domain.DashboardPostulaciones, error) {
	return s.dashboard.ObtenerDashboard(ctx)
}
