package postgres_test

import (
	"os"
	"regexp"
	"testing"
)

// El estado inicial de las fases se lee del propio script, sin base de datos: una base local
// creada con una versión anterior del script no lo refleja.
func TestSchema_LasFasesArrancanCerradas(t *testing.T) {
	sql, err := os.ReadFile("../../../../db/schema.sql")
	if err != nil {
		t.Fatalf("leyendo db/schema.sql: %v", err)
	}
	for _, fase := range []string{"Postulación", "Méritos", "Evaluación"} {
		fila := regexp.MustCompile(`\('` + fase + `',\s*\d+,\s*(TRUE|FALSE),\s*(TRUE|FALSE)\)`)
		m := fila.FindSubmatch(sql)
		if m == nil {
			t.Fatalf("db/schema.sql no siembra la fase %s", fase)
		}
		if string(m[1]) != "FALSE" || string(m[2]) != "FALSE" {
			t.Errorf("la fase %s debe arrancar sin publicar y sin formulario: publicada=%s, habilita_formulario=%s",
				fase, m[1], m[2])
		}
	}
}
