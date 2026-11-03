package domain

import "context"

// ConteosPorCargo agrupa las postulaciones por el cargo elegido.
type ConteosPorCargo struct {
	Afiliados int
	Jubilados int
	Empleador int
}

// ConteosPorEstado agrupa las postulaciones por su estado actual.
type ConteosPorEstado struct {
	Borrador int
	Enviada  int
	Cerrada  int
}

// DashboardPostulaciones contiene los indicadores del panel de administración.
type DashboardPostulaciones struct {
	PorCargo  ConteosPorCargo
	PorEstado ConteosPorEstado
	Total     int
}

// DashboardReader es el puerto de lectura de los indicadores de postulaciones.
type DashboardReader interface {
	ObtenerDashboard(ctx context.Context) (DashboardPostulaciones, error)
}
