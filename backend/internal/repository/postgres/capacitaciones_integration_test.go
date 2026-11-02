package postgres_test

import (
	"context"
	"testing"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/domain"
	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/integration/registrocivil"
	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/repository/postgres"
)

func TestCapacitaciones_GuardaVariasYConservaElOrden(t *testing.T) {
	pool := newPool(t)
	ctx := context.Background()
	postulanteID := crearPostulante(t, pool, registrocivil.CedulaFlujoFeliz)
	svc := newService(pool, postgres.NewCapacitacionesRepo(pool))
	filas := []domain.Capacitacion{
		{Tipo: domain.TipoCurso, Nombre: "Finanzas", Institucion: "Universidad", FechaInicio: "2024-01-01", FechaFin: "2024-01-05", Horas: 40},
		{Tipo: domain.TipoTaller, Nombre: "Mercado de capitales", Institucion: "Instituto", FechaInicio: "2024-02-01", FechaFin: "2024-02-02", Horas: 16},
	}
	p, err := svc.Guardar(ctx, postulanteID, domain.Formulario{Capacitaciones: filas})
	if err != nil {
		t.Fatal(err)
	}
	if len(p.Formulario.Capacitaciones) != 2 || p.Formulario.Capacitaciones[0].Nombre != "Finanzas" || p.Formulario.Capacitaciones[1].Nombre != "Mercado de capitales" {
		t.Fatalf("capacitaciones = %+v", p.Formulario.Capacitaciones)
	}
}
