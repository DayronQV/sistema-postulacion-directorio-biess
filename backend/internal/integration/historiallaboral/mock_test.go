package historiallaboral_test

import (
	"context"
	"errors"
	"testing"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/domain"
	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/integration/historiallaboral"
	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/integration/registrocivil"
)

func consultar(t *testing.T, cedula string) domain.Calidad {
	t.Helper()
	c, err := historiallaboral.NewMock().ConsultarCalidad(context.Background(), cedula)
	if err != nil {
		t.Fatalf("ConsultarCalidad(%s): %v", cedula, err)
	}
	return c
}

func TestMockCalidadDeCadaCedulaDePrueba(t *testing.T) {
	casos := []struct {
		nombre string
		cedula string
		quiere domain.Calidad
	}{
		{"afiliado", historiallaboral.CedulaAfiliado, domain.Calidad{EsAfiliado: true}},
		{"jubilado", historiallaboral.CedulaJubilado, domain.Calidad{EsJubilado: true}},
		{"empleador", historiallaboral.CedulaEmpleador, domain.Calidad{EsEmpleador: true}},
		{"ninguna condición", historiallaboral.CedulaNinguna, domain.Calidad{}},
		{"cédula desconocida se trata como afiliada", "1713579249", domain.Calidad{EsAfiliado: true}},
		{"cédula vacía se trata como afiliada", "", domain.Calidad{EsAfiliado: true}},
	}
	for _, k := range casos {
		t.Run(k.nombre, func(t *testing.T) {
			if got := consultar(t, k.cedula); got != k.quiere {
				t.Errorf("ConsultarCalidad(%q) = %+v, se esperaba %+v", k.cedula, got, k.quiere)
			}
		})
	}
}

// Las cédulas deben ser las mismas del Registro Civil: así una persona que pasa el
// registro tiene un resultado conocido en el Historial Laboral.
func TestMockUsaLasCedulasDelRegistroCivil(t *testing.T) {
	casos := []struct {
		nombre   string
		historia string
		registro string
	}{
		{"afiliado", historiallaboral.CedulaAfiliado, registrocivil.CedulaFlujoFeliz},
		{"jubilado", historiallaboral.CedulaJubilado, registrocivil.CedulaCasada},
		{"empleador", historiallaboral.CedulaEmpleador, registrocivil.CedulaUnionDeHecho},
		{"ninguna", historiallaboral.CedulaNinguna, registrocivil.CedulaSinCalidad},
	}
	rc := registrocivil.NewMock()
	for _, k := range casos {
		if k.historia != k.registro {
			t.Errorf("%s: Historial Laboral usa %s y Registro Civil usa %s", k.nombre, k.historia, k.registro)
		}
		if _, err := rc.ConsultarCiudadano(context.Background(), k.historia); err != nil {
			t.Errorf("%s: la cédula %s no existe en el Registro Civil simulado: %v", k.nombre, k.historia, err)
		}
	}
}

func TestMockLasCedulasDePruebaSonValidas(t *testing.T) {
	for _, cedula := range []string{
		historiallaboral.CedulaAfiliado, historiallaboral.CedulaJubilado,
		historiallaboral.CedulaEmpleador, historiallaboral.CedulaNinguna,
	} {
		if !domain.CedulaValida(cedula) {
			t.Errorf("%s no pasa el módulo 10", cedula)
		}
	}
}

func TestMockRespetaElContexto(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := historiallaboral.NewMock().ConsultarCalidad(ctx, historiallaboral.CedulaAfiliado)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("se esperaba context.Canceled, got %v", err)
	}
}
