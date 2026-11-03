package mailer

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/domain"
)

// URLMailjet es el endpoint de Mailjet para enviar correos (API v3.1).
const URLMailjet = "https://api.mailjet.com/v3.1/send"

// timeoutMailjet limita cada envío, para que un Mailjet lento no deje colgada la petición.
const timeoutMailjet = 10 * time.Second

// MailjetConfig es lo que necesita el adaptador de Mailjet.
type MailjetConfig struct {
	APIKey          string // MAILJET_API_KEY
	SecretKey       string // MAILJET_SECRET_KEY
	Remitente       string // MAIL_FROM: dirección verificada en Mailjet como remitente
	NombreRemitente string // MAIL_FROM_NAME: lo que ve el destinatario como "De"
	URL             string // vacío = URLMailjet; las pruebas apuntan a un servidor local
	Cliente         *http.Client
}

// Mailjet es un Mailer que envía correos reales por la API HTTPS de Mailjet
// (https://www.mailjet.com). Se usa en producción porque el servidor no puede abrir
// conexiones SMTP salientes: DigitalOcean bloquea los puertos 25, 465 y 587, y la API va
// por el puerto 443.
type Mailjet struct {
	cfg    MailjetConfig
	logger *slog.Logger
}

var _ domain.Mailer = (*Mailjet)(nil)

// NewMailjet crea el adaptador. Falla si faltan las claves o el remitente no es una
// dirección válida, para detectarlo al arrancar y no al primer correo.
func NewMailjet(cfg MailjetConfig, logger *slog.Logger) (*Mailjet, error) {
	var problemas []string
	if strings.TrimSpace(cfg.APIKey) == "" {
		problemas = append(problemas, "falta MAILJET_API_KEY")
	}
	if strings.TrimSpace(cfg.SecretKey) == "" {
		problemas = append(problemas, "falta MAILJET_SECRET_KEY")
	}
	if !direccionValida(cfg.Remitente) {
		problemas = append(problemas, "MAIL_FROM debe ser una dirección de correo válida")
	}
	if len(problemas) > 0 {
		return nil, errors.New("MAILER=mailjet: " + strings.Join(problemas, "; "))
	}
	if cfg.URL == "" {
		cfg.URL = URLMailjet
	}
	if cfg.Cliente == nil {
		cfg.Cliente = &http.Client{Timeout: timeoutMailjet}
	}
	return &Mailjet{cfg: cfg, logger: logger}, nil
}

type contactoMailjet struct {
	Email string `json:"Email"`
	Name  string `json:"Name,omitempty"`
}

type adjuntoMailjet struct {
	ContentType   string `json:"ContentType"`
	Filename      string `json:"Filename"`
	Base64Content string `json:"Base64Content"`
}

type mensajeMailjet struct {
	From        contactoMailjet   `json:"From"`
	To          []contactoMailjet `json:"To"`
	Subject     string            `json:"Subject"`
	TextPart    string            `json:"TextPart,omitempty"`
	HTMLPart    string            `json:"HTMLPart,omitempty"`
	Attachments []adjuntoMailjet  `json:"Attachments,omitempty"`
}

type envioMailjet struct {
	Messages []mensajeMailjet `json:"Messages"`
}

// Enviar valida el mensaje y lo manda a Mailjet. Una respuesta que no es 2xx devuelve un
// error con el código y el detalle de Mailjet (nunca las claves).
func (mj *Mailjet) Enviar(ctx context.Context, m domain.Mensaje) error {
	if err := m.Validar(); err != nil {
		return err
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	mensaje := mensajeMailjet{
		From:     contactoMailjet{Email: mj.cfg.Remitente, Name: mj.cfg.NombreRemitente},
		To:       []contactoMailjet{{Email: m.Para}},
		Subject:  m.Asunto,
		TextPart: m.Texto,
		HTMLPart: m.HTML,
	}
	for _, a := range m.Adjuntos {
		mensaje.Attachments = append(mensaje.Attachments, adjuntoMailjet{
			ContentType:   tipoDeContenido(a.Nombre),
			Filename:      a.Nombre,
			Base64Content: base64.StdEncoding.EncodeToString(a.Contenido),
		})
	}
	cuerpo, err := json.Marshal(envioMailjet{Messages: []mensajeMailjet{mensaje}})
	if err != nil {
		return fmt.Errorf("armando el correo para Mailjet: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, mj.cfg.URL, bytes.NewReader(cuerpo))
	if err != nil {
		return fmt.Errorf("armando la petición a Mailjet: %w", err)
	}
	req.SetBasicAuth(mj.cfg.APIKey, mj.cfg.SecretKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := mj.cfg.Cliente.Do(req)
	if err != nil {
		return fmt.Errorf("enviando el correo por Mailjet: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		detalle, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return fmt.Errorf("mailjet respondió %d: %s", resp.StatusCode, strings.TrimSpace(string(detalle)))
	}

	adjuntos := make([]string, 0, len(m.Adjuntos))
	for _, a := range m.Adjuntos {
		adjuntos = append(adjuntos, descripcionAdjunto(a))
	}
	mj.logger.Info("correo enviado por Mailjet", "para", m.Para, "asunto", m.Asunto, "adjuntos", adjuntos)
	return nil
}

// tipoDeContenido deduce el tipo MIME del adjunto por su extensión. Mailjet lo exige; los
// adjuntos del sistema son PDF.
func tipoDeContenido(nombre string) string {
	if strings.HasSuffix(strings.ToLower(nombre), ".pdf") {
		return "application/pdf"
	}
	return "application/octet-stream"
}

// direccionValida acepta solo una dirección escrita tal cual ("a@b.com"), con la misma
// regla que los destinatarios de domain.Mensaje.
func direccionValida(s string) bool {
	return domain.Mensaje{Para: s, Asunto: "x", Texto: "x"}.Validar() == nil
}
