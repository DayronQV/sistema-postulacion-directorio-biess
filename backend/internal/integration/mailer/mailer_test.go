package mailer_test

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"strings"
	"testing"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/domain"
	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/integration/mailer"
)

// logEnMemoria devuelve un logger que escribe en un buffer para poder leer lo que se
// registró.
func logEnMemoria() (*slog.Logger, *bytes.Buffer) {
	var buf bytes.Buffer
	return slog.New(slog.NewTextHandler(&buf, nil)), &buf
}

func mensaje() domain.Mensaje {
	return domain.Mensaje{
		Para:   "postulante@correo.com",
		Asunto: "Restablecer contraseña",
		Texto:  "Ingrese a https://biess.example/restablecer?token=abc123",
		HTML:   "<p>Restablecer</p>",
		Adjuntos: []domain.Adjunto{
			{Nombre: "postulacion.pdf", Contenido: bytes.Repeat([]byte("x"), 2048)},
		},
	}
}

func TestLogEscribeElCorreoEnElLog(t *testing.T) {
	logger, buf := logEnMemoria()
	var m domain.Mailer = mailer.NewLog(logger)

	if err := m.Enviar(context.Background(), mensaje()); err != nil {
		t.Fatalf("Enviar: %v", err)
	}

	salida := buf.String()
	for _, esperado := range []string{
		"correo simulado",
		"postulante@correo.com",
		"Restablecer contraseña",
		"token=abc123", // el enlace de restablecimiento se copia del log
		"<p>Restablecer</p>",
		"postulacion.pdf (2.0 KB)", // nombre y tamaño del adjunto
	} {
		if !strings.Contains(salida, esperado) {
			t.Errorf("el log no contiene %q:\n%s", esperado, salida)
		}
	}
}

func TestLogNoEscribeElContenidoDeLosAdjuntos(t *testing.T) {
	logger, buf := logEnMemoria()
	m := mensaje()
	m.Adjuntos[0].Contenido = []byte("CONTENIDO-SECRETO-DEL-PDF")

	if err := mailer.NewLog(logger).Enviar(context.Background(), m); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(buf.String(), "CONTENIDO-SECRETO-DEL-PDF") {
		t.Fatal("el log no debe contener los bytes de los adjuntos")
	}
}

func TestLogFuncionaSinAdjuntos(t *testing.T) {
	logger, buf := logEnMemoria()
	m := domain.Mensaje{Para: "a@b.com", Asunto: "Hola", Texto: "Hola"}

	if err := mailer.NewLog(logger).Enviar(context.Background(), m); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "a@b.com") {
		t.Fatalf("no se registró el correo: %s", buf.String())
	}
}

func TestLogRechazaUnMensajeInvalidoSinRegistrarlo(t *testing.T) {
	logger, buf := logEnMemoria()
	m := mensaje()
	m.Para = "no-es-un-correo"

	err := mailer.NewLog(logger).Enviar(context.Background(), m)
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("se esperaba ErrValidation, got %v", err)
	}
	if buf.Len() != 0 {
		t.Fatalf("un mensaje inválido no debe escribirse en el log: %s", buf.String())
	}
}

func TestLogRespetaElContexto(t *testing.T) {
	logger, buf := logEnMemoria()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := mailer.NewLog(logger).Enviar(ctx, mensaje())
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("se esperaba context.Canceled, got %v", err)
	}
	if buf.Len() != 0 {
		t.Fatal("con el contexto cancelado no se debe registrar nada")
	}
}

func TestNewEligeElAdaptadorPorConfiguracion(t *testing.T) {
	logger, _ := logEnMemoria()

	// Vacío, "log" y cualquier combinación de mayúsculas o espacios dan el simulador.
	for _, valor := range []string{"", "log", "LOG", " Log "} {
		m, err := mailer.New(valor, mailer.MailjetConfig{}, logger)
		if err != nil {
			t.Errorf("New(%q): %v", valor, err)
			continue
		}
		if _, ok := m.(*mailer.Log); !ok {
			t.Errorf("New(%q) devolvió %T, se esperaba *mailer.Log", valor, m)
		}
	}
}

func TestNewFallaAlArrancarConUnValorNoDisponible(t *testing.T) {
	logger, _ := logEnMemoria()

	// SMTP no se puede usar en el servidor (puertos bloqueados): se pide usar Mailjet.
	_, err := mailer.New("smtp", mailer.MailjetConfig{}, logger)
	if err == nil || !strings.Contains(err.Error(), "mailjet") {
		t.Errorf("MAILER=smtp debe explicar que se use mailjet, got %v", err)
	}

	// Un valor desconocido lista las opciones válidas.
	_, err = mailer.New("sendgrid", mailer.MailjetConfig{}, logger)
	if err == nil || !strings.Contains(err.Error(), "sendgrid") || !strings.Contains(err.Error(), "log") {
		t.Errorf("un valor desconocido debe mencionar el valor y las opciones, got %v", err)
	}
}
