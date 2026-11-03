package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/domain"
	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/service"
)

type fakeDashboard struct {
	datos domain.DashboardPostulaciones
	err   error
}

func (f fakeDashboard) ObtenerDashboard(context.Context) (domain.DashboardPostulaciones, error) {
	return f.datos, f.err
}

func TestDashboard_DevuelveLosConteosDelRepositorio(t *testing.T) {
	esperado := domain.DashboardPostulaciones{
		PorCargo:  domain.ConteosPorCargo{Afiliados: 4, Jubilados: 3, Empleador: 2},
		PorEstado: domain.ConteosPorEstado{Borrador: 2, Enviada: 6, Cerrada: 1},
		Total:     9,
	}
	got, err := service.NewDashboardService(fakeDashboard{datos: esperado}).Obtener(context.Background())
	if err != nil || got != esperado {
		t.Fatalf("resultado = %+v, err = %v", got, err)
	}
}

func TestDashboard_ConservaElErrorDelRepositorio(t *testing.T) {
	boom := errors.New("base no disponible")
	_, err := service.NewDashboardService(fakeDashboard{err: boom}).Obtener(context.Background())
	if !errors.Is(err, boom) {
		t.Fatalf("se esperaba el error original, llegó %v", err)
	}
}
