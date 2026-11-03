package padron_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/domain"
	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/integration/padron"
)

// Cédulas válidas por módulo 10 que no pertenecen a nadie conocido.
const (
	cedulaA = "1713579249"
	cedulaB = "0913579249"
	cedulaC = "1313579243"
)

const personaSoltera = `{
  "cedula": "1713579249", "nombres": "Ana María", "apellidos": "Loor Mero", "genero": "Femenino",
  "fecha_nacimiento": "1990-05-17", "nacionalidad": "Ecuatoriana", "estado_civil": "Soltero/a",
  "fecha_expedicion": "2020-02-03", "codigo_dactilar": "V1234V5678", "calidades": ["afiliado"]
}`

const personaCasada = `{
  "cedula": "0913579249", "nombres": "Luis", "apellidos": "Pico Zambrano", "genero": "Masculino",
  "fecha_nacimiento": "1980-01-02", "nacionalidad": "Ecuatoriana", "estado_civil": "Casado/a",
  "pareja": {"nombres_apellidos": "Rosa Vera Mero", "cedula": "1313579243"},
  "fecha_expedicion": "2018-07-09", "codigo_dactilar": "E4321I8765", "calidades": ["jubilado", "empleador"]
}`

func escribir(t *testing.T, contenido string) string {
	t.Helper()
	ruta := filepath.Join(t.TempDir(), "padron.json")
	if err := os.WriteFile(ruta, []byte(contenido), 0o600); err != nil {
		t.Fatal(err)
	}
	return ruta
}

func TestCargarConvierteLosCamposDeCadaPersona(t *testing.T) {
	registros, err := padron.Cargar(escribir(t, "["+personaSoltera+","+personaCasada+"]"))
	if err != nil {
		t.Fatalf("Cargar: %v", err)
	}
	if len(registros) != 2 {
		t.Fatalf("se esperaban 2 personas, hay %d", len(registros))
	}

	a := registros[0]
	if a.Ciudadano.Cedula != cedulaA || a.Ciudadano.Nombres != "Ana María" || a.Ciudadano.Apellidos != "Loor Mero" ||
		a.Ciudadano.Genero != "Femenino" || a.Ciudadano.Nacionalidad != "Ecuatoriana" ||
		a.Ciudadano.EstadoCivil != domain.EstadoSoltero || a.Ciudadano.CodigoDactilar != "V1234V5678" {
		t.Errorf("datos de la primera persona mal leídos: %+v", a.Ciudadano)
	}
	if !a.Ciudadano.FechaNacimiento.Equal(time.Date(1990, 5, 17, 0, 0, 0, 0, time.UTC)) ||
		!a.Ciudadano.FechaExpedicion.Equal(time.Date(2020, 2, 3, 0, 0, 0, 0, time.UTC)) {
		t.Errorf("fechas mal leídas: %v y %v", a.Ciudadano.FechaNacimiento, a.Ciudadano.FechaExpedicion)
	}
	if a.Ciudadano.Pareja != nil {
		t.Errorf("una persona soltera no lleva pareja: %+v", a.Ciudadano.Pareja)
	}
	if a.Calidad != (domain.Calidad{EsAfiliado: true}) {
		t.Errorf("calidad = %+v, se esperaba solo afiliado", a.Calidad)
	}

	b := registros[1]
	if b.Ciudadano.Pareja == nil || b.Ciudadano.Pareja.NombresApellidos != "Rosa Vera Mero" || b.Ciudadano.Pareja.Cedula != cedulaC {
		t.Errorf("pareja mal leída: %+v", b.Ciudadano.Pareja)
	}
	if b.Calidad != (domain.Calidad{EsJubilado: true, EsEmpleador: true}) {
		t.Errorf("calidad = %+v, se esperaba jubilado y empleador", b.Calidad)
	}
}

func TestCargarSinCalidadesNoDaNingunaCalidad(t *testing.T) {
	json := strings.Replace(personaSoltera, `"calidades": ["afiliado"]`, `"calidades": []`, 1)
	registros, err := padron.Cargar(escribir(t, "["+json+"]"))
	if err != nil {
		t.Fatalf("Cargar: %v", err)
	}
	if registros[0].Calidad != (domain.Calidad{}) {
		t.Errorf("calidad = %+v, se esperaba ninguna", registros[0].Calidad)
	}
}

func TestCargarRechazaLoQueEstaMal(t *testing.T) {
	casos := []struct {
		nombre    string
		contenido string
		mensaje   string
	}{
		{"cédula con mal dígito verificador", strings.Replace(personaSoltera, cedulaA, "1713579240", 1), "cédula no válida"},
		{"cédula repetida", "[" + personaSoltera + "," + personaSoltera + "]", "cédula repetida"},
		{"género desconocido", strings.Replace(personaSoltera, `"Femenino"`, `"F"`, 1), "genero"},
		{"fecha de nacimiento ilegible", strings.Replace(personaSoltera, "1990-05-17", "17/05/1990", 1), "fecha_nacimiento"},
		{"fecha de expedición en el futuro", strings.Replace(personaSoltera, "2020-02-03", "2999-01-01", 1), "fecha_expedicion"},
		{"código dactilar corto", strings.Replace(personaSoltera, "V1234V5678", "V1234", 1), "codigo_dactilar"},
		{"código dactilar en minúscula", strings.Replace(personaSoltera, "V1234V5678", "v1234v5678", 1), "codigo_dactilar"},
		{"estado civil desconocido", strings.Replace(personaSoltera, "Soltero/a", "Complicado", 1), "estado_civil"},
		{"casado sin pareja", strings.Replace(personaSoltera, "Soltero/a", "Casado/a", 1), "pareja"},
		{"soltero con pareja", strings.Replace(personaCasada, "Casado/a", "Soltero/a", 1), "pareja"},
		{"pareja con cédula inválida", strings.Replace(personaCasada, cedulaC, "1313579240", 1), "pareja"},
		{"calidad desconocida", strings.Replace(personaSoltera, `"afiliado"`, `"pensionista"`, 1), "calidades"},
		{"campo que no existe", strings.Replace(personaSoltera, `"genero"`, `"sexo"`, 1), "sexo"},
	}
	for _, k := range casos {
		t.Run(k.nombre, func(t *testing.T) {
			contenido := k.contenido
			if !strings.HasPrefix(contenido, "[") {
				contenido = "[" + contenido + "]"
			}
			_, err := padron.Cargar(escribir(t, contenido))
			if err == nil {
				t.Fatal("se esperaba un error")
			}
			if !strings.Contains(err.Error(), k.mensaje) {
				t.Errorf("el error %q no menciona %q", err, k.mensaje)
			}
		})
	}
}

func TestCargarListaTodosLosProblemasSinMostrarDatosPersonales(t *testing.T) {
	malo := strings.Replace(personaSoltera, cedulaA, "1713579240", 1)
	malo = strings.Replace(malo, "V1234V5678", "x", 1)
	_, err := padron.Cargar(escribir(t, "["+malo+"]"))
	if err == nil {
		t.Fatal("se esperaba un error")
	}
	for _, quiere := range []string{"cédula no válida", "codigo_dactilar"} {
		if !strings.Contains(err.Error(), quiere) {
			t.Errorf("el error %q no menciona %q", err, quiere)
		}
	}
	for _, privado := range []string{"Ana María", "Loor Mero", "1713579240"} {
		if strings.Contains(err.Error(), privado) {
			t.Errorf("el error no debe repetir datos personales (%q): %v", privado, err)
		}
	}
}

func TestCargarRechazaArchivosInservibles(t *testing.T) {
	casos := map[string]string{
		"vacío":                "",
		"no es JSON":           "esto no es json",
		"no es un arreglo":     personaSoltera,
		"arreglo sin personas": "[]",
	}
	for nombre, contenido := range casos {
		t.Run(nombre, func(t *testing.T) {
			if _, err := padron.Cargar(escribir(t, contenido)); err == nil {
				t.Error("se esperaba un error")
			}
		})
	}
	if _, err := padron.Cargar(filepath.Join(t.TempDir(), "no-existe.json")); err == nil {
		t.Error("un archivo que no existe debe dar error")
	}
	if _, err := padron.Cargar(t.TempDir()); err == nil {
		t.Error("una carpeta en lugar del archivo debe dar error")
	}
}
