package mailer_test

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/domain"
	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/integration/mailer"
)

const (
	claveMailjet   = "clave-publica-de-prueba"
	secretoMailjet = "clave-secreta-de-prueba"
)

// peticionMailjet es lo que el servidor falso recibió.
type peticionMailjet struct {
	metodo, usuario, clave, tipo string
	conAuth                      bool
	mensajes                     []map[string]any
}

// mailjetFalso imita la API de Mailjet: guarda la petición y responde con el código dado.
func mailjetFalso(t *testing.T, codigo int, respuesta string) (*httptest.Server, *peticionMailjet) {
	t.Helper()
	var p peticionMailjet
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p.metodo, p.tipo = r.Method, r.Header.Get("Content-Type")
		p.usuario, p.clave, p.conAuth = r.BasicAuth()
		datos, _ := io.ReadAll(r.Body)
		var cuerpo struct {
			Messages []map[string]any
		}
		_ = json.Unmarshal(datos, &cuerpo)
		p.mensajes = cuerpo.Messages
		w.WriteHeader(codigo)
		_, _ = w.Write([]byte(respuesta))
	}))
	t.Cleanup(srv.Close)
	return srv, &p
}

func nuevoMailjet(t *testing.T, url string) *mailer.Mailjet {
	t.Helper()
	logger, _ := logEnMemoria()
	m, err := mailer.NewMailjet(mailer.MailjetConfig{
		APIKey: claveMailjet, SecretKey: secretoMailjet,
		Remitente: "concurso@gmail.com", NombreRemitente: "Concurso BIESS", URL: url,
	}, logger)
	if err != nil {
		t.Fatalf("NewMailjet: %v", err)
	}
	return m
}

func TestMailjetEnviaElCorreoConElFormatoDeLaAPI(t *testing.T) {
	srv, p := mailjetFalso(t, http.StatusOK, `{"Messages":[{"Status":"success"}]}`)

	if err := nuevoMailjet(t, srv.URL).Enviar(context.Background(), mensaje()); err != nil {
		t.Fatalf("Enviar: %v", err)
	}

	if p.metodo != http.MethodPost || p.tipo != "application/json" {
		t.Errorf("petición inesperada: método %q, Content-Type %q", p.metodo, p.tipo)
	}
	if !p.conAuth || p.usuario != claveMailjet || p.clave != secretoMailjet {
		t.Errorf("autenticación básica incorrecta: %q / %q", p.usuario, p.clave)
	}
	if len(p.mensajes) != 1 {
		t.Fatalf("se esperaba 1 mensaje, llegaron %d", len(p.mensajes))
	}
	m := p.mensajes[0]
	de, _ := m["From"].(map[string]any)
	if de["Email"] != "concurso@gmail.com" || de["Name"] != "Concurso BIESS" {
		t.Errorf("remitente = %v", de)
	}
	para, _ := m["To"].([]any)
	if len(para) != 1 || para[0].(map[string]any)["Email"] != "postulante@correo.com" {
		t.Errorf("destinatarios = %v, se esperaba solo postulante@correo.com", para)
	}
	if m["Subject"] != "Restablecer contraseña" || m["HTMLPart"] != "<p>Restablecer</p>" ||
		!strings.Contains(m["TextPart"].(string), "token=abc123") {
		t.Errorf("asunto o cuerpo mal enviados: %v", m)
	}
}

func TestMailjetEnviaLosAdjuntosEnBase64ConSuTipo(t *testing.T) {
	srv, p := mailjetFalso(t, http.StatusOK, `{}`)
	if err := nuevoMailjet(t, srv.URL).Enviar(context.Background(), mensaje()); err != nil {
		t.Fatalf("Enviar: %v", err)
	}
	adjuntos, _ := p.mensajes[0]["Attachments"].([]any)
	if len(adjuntos) != 1 {
		t.Fatalf("se esperaba 1 adjunto, llegaron %v", p.mensajes[0]["Attachments"])
	}
	a := adjuntos[0].(map[string]any)
	contenido, err := base64.StdEncoding.DecodeString(a["Base64Content"].(string))
	if a["Filename"] != "postulacion.pdf" || a["ContentType"] != "application/pdf" || err != nil || len(contenido) != 2048 {
		t.Errorf("adjunto mal enviado: %v (%d bytes, error %v)", a, len(contenido), err)
	}
}

func TestMailjetSinAdjuntosNoMandaElCampo(t *testing.T) {
	srv, p := mailjetFalso(t, http.StatusOK, `{}`)
	m := mensaje()
	m.Adjuntos = nil
	if err := nuevoMailjet(t, srv.URL).Enviar(context.Background(), m); err != nil {
		t.Fatalf("Enviar: %v", err)
	}
	if _, ok := p.mensajes[0]["Attachments"]; ok {
		t.Errorf("sin adjuntos no debe ir el campo Attachments: %v", p.mensajes[0])
	}
}

func TestMailjetDevuelveElErrorDeLaAPISinLasClaves(t *testing.T) {
	srv, _ := mailjetFalso(t, http.StatusBadRequest, `{"Messages":[{"Status":"error","Errors":[{"ErrorMessage":"sender not validated"}]}]}`)
	err := nuevoMailjet(t, srv.URL).Enviar(context.Background(), mensaje())
	if err == nil || !strings.Contains(err.Error(), "400") || !strings.Contains(err.Error(), "sender not validated") {
		t.Errorf("se esperaba el código y el mensaje de Mailjet, got %v", err)
	}
	if err != nil && (strings.Contains(err.Error(), claveMailjet) || strings.Contains(err.Error(), secretoMailjet)) {
		t.Errorf("el error no debe mostrar las claves: %v", err)
	}
}

func TestMailjetRechazaUnMensajeInvalidoSinLlamarALaAPI(t *testing.T) {
	srv, p := mailjetFalso(t, http.StatusOK, `{}`)
	m := mensaje()
	m.Asunto = ""
	if err := nuevoMailjet(t, srv.URL).Enviar(context.Background(), m); !errors.Is(err, domain.ErrValidation) {
		t.Errorf("se esperaba un error de validación, got %v", err)
	}
	if p.metodo != "" {
		t.Error("un mensaje inválido no debe llegar a Mailjet")
	}
}

func TestMailjetRespetaElContexto(t *testing.T) {
	srv, p := mailjetFalso(t, http.StatusOK, `{}`)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := nuevoMailjet(t, srv.URL).Enviar(ctx, mensaje()); !errors.Is(err, context.Canceled) {
		t.Errorf("se esperaba context.Canceled, got %v", err)
	}
	if p.metodo != "" {
		t.Error("con el contexto cancelado no se debe llamar a Mailjet")
	}
}

func TestNewMailjetFallaSinClavesOSinRemitente(t *testing.T) {
	logger, _ := logEnMemoria()
	casos := []struct {
		cfg     mailer.MailjetConfig
		mensaje string
	}{
		{mailer.MailjetConfig{SecretKey: "s", Remitente: "a@b.com"}, "MAILJET_API_KEY"},
		{mailer.MailjetConfig{APIKey: "k", Remitente: "a@b.com"}, "MAILJET_SECRET_KEY"},
		{mailer.MailjetConfig{APIKey: "k", SecretKey: "s", Remitente: "no-es-correo"}, "MAIL_FROM"},
	}
	for _, k := range casos {
		if _, err := mailer.NewMailjet(k.cfg, logger); err == nil || !strings.Contains(err.Error(), k.mensaje) {
			t.Errorf("%+v: se esperaba un error que pida %s, got %v", k.cfg, k.mensaje, err)
		}
	}
}

func TestNewEligeMailjetConMailerMailjet(t *testing.T) {
	logger, _ := logEnMemoria()
	m, err := mailer.New("MailJet", mailer.MailjetConfig{APIKey: "k", SecretKey: "s", Remitente: "a@b.com"}, logger)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if _, ok := m.(*mailer.Mailjet); !ok {
		t.Errorf("New(mailjet) devolvió %T, se esperaba *mailer.Mailjet", m)
	}
	if _, err := mailer.New("mailjet", mailer.MailjetConfig{}, logger); err == nil {
		t.Error("MAILER=mailjet sin configuración debe fallar al arrancar")
	}
}
