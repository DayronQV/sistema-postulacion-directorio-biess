package main

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/config"
	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/domain"
	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/integration/mailer"
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

func TestCorreoDesdeConfigConMailerLogNoNecesitaClaves(t *testing.T) {
	if m, err := correoDesdeConfig(config.Config{Mailer: "log"}, logDescartado); err != nil {
		t.Errorf("MAILER=log no necesita nada más: %v", err)
	} else if _, ok := m.(*mailer.Log); !ok {
		t.Errorf("MAILER=log devolvió %T", m)
	}
}

func TestCorreoDesdeConfigPasaLasClavesYElRemitenteAMailjet(t *testing.T) {
	cfg := config.Config{Mailer: "mailjet", MailjetAPIKey: "k", MailjetSecretKey: "s", MailFrom: "concurso@gmail.com", MailFromName: "Concurso BIESS"}
	m, err := correoDesdeConfig(cfg, logDescartado)
	if err != nil {
		t.Fatalf("con claves y remitente válidos debe arrancar: %v", err)
	}
	if _, ok := m.(*mailer.Mailjet); !ok {
		t.Errorf("MAILER=mailjet devolvió %T", m)
	}
	for nombre, cambiar := range map[string]func(*config.Config){
		"sin MAILJET_API_KEY":    func(c *config.Config) { c.MailjetAPIKey = "" },
		"sin MAILJET_SECRET_KEY": func(c *config.Config) { c.MailjetSecretKey = "" },
		"sin MAIL_FROM":          func(c *config.Config) { c.MailFrom = "" },
	} {
		c := cfg
		cambiar(&c)
		if _, err := correoDesdeConfig(c, logDescartado); err == nil {
			t.Errorf("%s no debe arrancar", nombre)
		}
	}
}

func TestIntegracionesFallaSiElArchivoNoExiste(t *testing.T) {
	if _, _, err := integraciones(filepath.Join(t.TempDir(), "no-existe.json"), logDescartado); err == nil {
		t.Error("un archivo ausente debe detener el arranque")
	}
}
