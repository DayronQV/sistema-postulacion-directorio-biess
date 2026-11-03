package historiallaboral_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/domain"
	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/integration/historiallaboral"
)

func persona(cedula, calidades string) string {
	return `{"cedula": "` + cedula + `", "nombres": "Ana", "apellidos": "Loor", "genero": "Femenino",
	  "fecha_nacimiento": "1990-05-17", "nacionalidad": "Ecuatoriana", "estado_civil": "Soltero/a",
	  "fecha_expedicion": "2020-02-03", "codigo_dactilar": "V1234V5678", "calidades": ` + calidades + `}`
}

func mockDesdeArchivo(t *testing.T) *historiallaboral.Mock {
	t.Helper()
	contenido := "[" + persona("1713579249", `["afiliado"]`) + "," + persona("0913579249", `["jubilado"]`) + "," +
		persona("1313579243", `["empleador"]`) + "," + persona("1712345675", `[]`) + "]"
	ruta := filepath.Join(t.TempDir(), "padron.json")
	if err := os.WriteFile(ruta, []byte(contenido), 0o600); err != nil {
		t.Fatal(err)
	}
	m, err := historiallaboral.NewDesdeArchivo(ruta)
	if err != nil {
		t.Fatalf("NewDesdeArchivo: %v", err)
	}
	return m
}

func TestDesdeArchivoDaLaCalidadDeCadaPersona(t *testing.T) {
	m := mockDesdeArchivo(t)
	casos := map[string]domain.Calidad{
		"1713579249": {EsAfiliado: true},
		"0913579249": {EsJubilado: true},
		"1313579243": {EsEmpleador: true},
		"1712345675": {},
	}
	for cedula, quiere := range casos {
		got, err := m.ConsultarCalidad(context.Background(), cedula)
		if err != nil {
			t.Fatalf("ConsultarCalidad(%s): %v", cedula, err)
		}
		if got != quiere {
			t.Errorf("ConsultarCalidad(%s) = %+v, se esperaba %+v", cedula, got, quiere)
		}
	}
}

func TestDesdeArchivoUnaCedulaQueNoConstaNoTieneNingunaCalidad(t *testing.T) {
	got, err := mockDesdeArchivo(t).ConsultarCalidad(context.Background(), "1700000000")
	if err != nil {
		t.Fatal(err)
	}
	if got != (domain.Calidad{}) {
		t.Errorf("calidad = %+v: con archivo, una cédula desconocida no debe contar como afiliada", got)
	}
}

func TestDesdeArchivoFallaSiElArchivoNoSirve(t *testing.T) {
	if _, err := historiallaboral.NewDesdeArchivo(filepath.Join(t.TempDir(), "no-existe.json")); err == nil {
		t.Error("un archivo que no existe debe dar error")
	}
}
