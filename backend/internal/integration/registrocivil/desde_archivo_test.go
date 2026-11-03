package registrocivil_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/domain"
	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/integration/registrocivil"
)

const padronDePrueba = `[{
  "cedula": "1713579249", "nombres": "Ana María", "apellidos": "Loor Mero", "genero": "Femenino",
  "fecha_nacimiento": "1990-05-17", "nacionalidad": "Ecuatoriana", "estado_civil": "Casado/a",
  "pareja": {"nombres_apellidos": "Pedro Vera Pico", "cedula": "0913579249"},
  "fecha_expedicion": "2020-02-03", "codigo_dactilar": "V1234V5678", "calidades": ["afiliado"]
}]`

func mockDesdeArchivo(t *testing.T) *registrocivil.Mock {
	t.Helper()
	ruta := filepath.Join(t.TempDir(), "padron.json")
	if err := os.WriteFile(ruta, []byte(padronDePrueba), 0o600); err != nil {
		t.Fatal(err)
	}
	m, err := registrocivil.NewDesdeArchivo(ruta)
	if err != nil {
		t.Fatalf("NewDesdeArchivo: %v", err)
	}
	return m
}

func TestDesdeArchivoDevuelveLasPersonasDelArchivo(t *testing.T) {
	c, err := mockDesdeArchivo(t).ConsultarCiudadano(context.Background(), "1713579249")
	if err != nil {
		t.Fatalf("ConsultarCiudadano: %v", err)
	}
	if c.Nombres != "Ana María" || c.CodigoDactilar != "V1234V5678" || c.EstadoCivil != domain.EstadoCasado ||
		c.Pareja == nil || c.Pareja.NombresApellidos != "Pedro Vera Pico" {
		t.Errorf("datos inesperados: %+v", c)
	}
}

func TestDesdeArchivoNoConoceLosCiudadanosDePrueba(t *testing.T) {
	m := mockDesdeArchivo(t)
	for _, cedula := range todasLasCedulas {
		if _, err := m.ConsultarCiudadano(context.Background(), cedula); !errors.Is(err, domain.ErrNotFound) {
			t.Errorf("%s: se esperaba ErrNotFound, got %v", cedula, err)
		}
	}
}

func TestDesdeArchivoFallaSiElArchivoNoSirve(t *testing.T) {
	if _, err := registrocivil.NewDesdeArchivo(filepath.Join(t.TempDir(), "no-existe.json")); err == nil {
		t.Error("un archivo que no existe debe dar error")
	}
}
