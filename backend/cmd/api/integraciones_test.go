package main

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/domain"
)

var logDescartado = slog.New(slog.NewTextHandler(io.Discard, nil))

func TestIntegracionesSinArchivoUsaLosCiudadanosDePrueba(t *testing.T) {
	registro, historial, err := integraciones("", logDescartado)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := registro.ConsultarCiudadano(context.Background(), "1712345675"); err != nil {
		t.Errorf("el ciudadano de prueba debe existir: %v", err)
	}
	calidad, _ := historial.ConsultarCalidad(context.Background(), "1713579249")
	if !calidad.EsAfiliado {
		t.Errorf("sin archivo, una cédula desconocida cuenta como afiliada: %+v", calidad)
	}
}

func TestIntegracionesConArchivoSoloConoceASuGente(t *testing.T) {
	ruta := filepath.Join(t.TempDir(), "padron.json")
	contenido := `[{"cedula": "1713579249", "nombres": "Ana", "apellidos": "Loor", "genero": "Femenino",
	  "fecha_nacimiento": "1990-05-17", "nacionalidad": "Ecuatoriana", "estado_civil": "Soltero/a",
	  "fecha_expedicion": "2020-02-03", "codigo_dactilar": "V1234V5678", "calidades": ["jubilado"]}]`
	if err := os.WriteFile(ruta, []byte(contenido), 0o600); err != nil {
		t.Fatal(err)
	}

	registro, historial, err := integraciones(ruta, logDescartado)
	if err != nil {
		t.Fatal(err)
	}
	if c, err := registro.ConsultarCiudadano(context.Background(), "1713579249"); err != nil || c.Nombres != "Ana" {
		t.Errorf("la persona del archivo debe existir: %+v, %v", c, err)
	}
	if _, err := registro.ConsultarCiudadano(context.Background(), "1712345675"); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("el ciudadano de prueba no debe existir con archivo: %v", err)
	}
	if calidad, _ := historial.ConsultarCalidad(context.Background(), "1713579249"); !calidad.EsJubilado || calidad.EsAfiliado {
		t.Errorf("calidad = %+v, se esperaba solo jubilado", calidad)
	}
}

func TestIntegracionesFallaSiElArchivoNoExiste(t *testing.T) {
	if _, _, err := integraciones(filepath.Join(t.TempDir(), "no-existe.json"), logDescartado); err == nil {
		t.Error("un archivo ausente debe detener el arranque")
	}
}
