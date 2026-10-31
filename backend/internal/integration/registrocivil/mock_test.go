package registrocivil_test

import (
	"context"
	"errors"
	"regexp"
	"testing"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/domain"
	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/integration/registrocivil"
)

var todasLasCedulas = []string{
	registrocivil.CedulaFlujoFeliz,
	registrocivil.CedulaCasada,
	registrocivil.CedulaUnionDeHecho,
	registrocivil.CedulaExtranjera,
	registrocivil.CedulaSinCalidad,
	registrocivil.CedulaBloqueoPorIntento,
	registrocivil.CedulaRecuperacion,
}

func consultar(t *testing.T, cedula string) domain.Ciudadano {
	t.Helper()
	c, err := registrocivil.NewMock().ConsultarCiudadano(context.Background(), cedula)
	if err != nil {
		t.Fatalf("ConsultarCiudadano(%s): %v", cedula, err)
	}
	return c
}

func TestMockDevuelveLosSieteCiudadanos(t *testing.T) {
	for _, cedula := range todasLasCedulas {
		c := consultar(t, cedula)
		if c.Cedula != cedula {
			t.Errorf("devolvió la cédula %s al consultar %s", c.Cedula, cedula)
		}
		if c.Nombres == "" || c.Apellidos == "" || c.Genero == "" || c.Nacionalidad == "" {
			t.Errorf("%s: faltan datos personales: %+v", cedula, c)
		}
		if c.FechaNacimiento.IsZero() || c.FechaExpedicion.IsZero() {
			t.Errorf("%s: faltan fechas", cedula)
		}
	}
}

func TestMockCedulaInexistenteDaErrNotFound(t *testing.T) {
	m := registrocivil.NewMock()
	for _, cedula := range []string{
		"1713579249", // cédula válida por módulo 10 que no es una de las 7
		"0913579249", // cédula de una pareja: está en los datos, pero no es un ciudadano
		"1700000000", "1712345676", "", "no-es-una-cedula",
	} {
		_, err := m.ConsultarCiudadano(context.Background(), cedula)
		if !errors.Is(err, domain.ErrNotFound) {
			t.Errorf("ConsultarCiudadano(%q): se esperaba ErrNotFound, got %v", cedula, err)
		}
	}
}

func TestMockRespetaElContexto(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := registrocivil.NewMock().ConsultarCiudadano(ctx, registrocivil.CedulaFlujoFeliz)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("se esperaba context.Canceled, got %v", err)
	}
}

// Todas deben pasar la validación del dominio, y los datos de identidad deben tener
// el formato real y no repetirse entre ciudadanos.
func TestMockDatosCoherentes(t *testing.T) {
	codigo := regexp.MustCompile(`^[A-Z0-9]{10}$`)
	vistos := map[string]string{}
	for _, cedula := range todasLasCedulas {
		c := consultar(t, cedula)
		if !domain.CedulaValida(c.Cedula) {
			t.Errorf("%s: la cédula no pasa el módulo 10", cedula)
		}
		if !codigo.MatchString(c.CodigoDactilar) {
			t.Errorf("%s: código dactilar %q no son 10 caracteres alfanuméricos en mayúscula", cedula, c.CodigoDactilar)
		}
		if otra, repetido := vistos[c.CodigoDactilar]; repetido {
			t.Errorf("el código dactilar %s se repite en %s y %s", c.CodigoDactilar, otra, cedula)
		}
		vistos[c.CodigoDactilar] = cedula
	}
}

func TestMockCasosDeNegocio(t *testing.T) {
	casos := []struct {
		cedula       string
		nacionalidad string
		estado       domain.EstadoCivil
		conPareja    bool
	}{
		{registrocivil.CedulaFlujoFeliz, "Ecuatoriana", domain.EstadoSoltero, false},
		{registrocivil.CedulaCasada, "Ecuatoriana", domain.EstadoCasado, true},
		{registrocivil.CedulaUnionDeHecho, "Ecuatoriana", domain.EstadoUnionDeHecho, true},
		{registrocivil.CedulaExtranjera, "Colombiana", domain.EstadoSoltero, false},
		{registrocivil.CedulaSinCalidad, "Ecuatoriana", domain.EstadoSoltero, false},
		{registrocivil.CedulaBloqueoPorIntento, "Ecuatoriana", domain.EstadoSoltero, false},
		{registrocivil.CedulaRecuperacion, "Ecuatoriana", domain.EstadoDivorciado, false},
	}
	for _, k := range casos {
		c := consultar(t, k.cedula)
		if c.Nacionalidad != k.nacionalidad {
			t.Errorf("%s: nacionalidad %q, se esperaba %q", k.cedula, c.Nacionalidad, k.nacionalidad)
		}
		if c.EstadoCivil != k.estado {
			t.Errorf("%s: estado civil %q, se esperaba %q", k.cedula, c.EstadoCivil, k.estado)
		}
		if (c.Pareja != nil) != k.conPareja {
			t.Errorf("%s: pareja presente = %v, se esperaba %v", k.cedula, c.Pareja != nil, k.conPareja)
		}
		if c.EstadoCivil.ConPareja() != k.conPareja {
			t.Errorf("%s: ConPareja() = %v, se esperaba %v", k.cedula, c.EstadoCivil.ConPareja(), k.conPareja)
		}
		if c.Pareja != nil && (c.Pareja.NombresApellidos == "" || !domain.CedulaValida(c.Pareja.Cedula)) {
			t.Errorf("%s: datos de la pareja incompletos o con cédula inválida: %+v", k.cedula, c.Pareja)
		}
	}
}

// Quien consulta no debe poder alterar los datos del simulador.
func TestMockDevuelveCopiaDeLaPareja(t *testing.T) {
	m := registrocivil.NewMock()
	c1, _ := m.ConsultarCiudadano(context.Background(), registrocivil.CedulaCasada)
	c1.Pareja.NombresApellidos = "Alterado"

	c2, _ := m.ConsultarCiudadano(context.Background(), registrocivil.CedulaCasada)
	if c2.Pareja.NombresApellidos == "Alterado" {
		t.Fatal("modificar la respuesta cambió los datos del simulador")
	}
}
