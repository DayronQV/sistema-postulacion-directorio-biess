package mailer

import (
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/domain"
)

// Adaptadores disponibles, los valores de la variable MAILER.
const (
	AdaptadorLog     = "log"
	AdaptadorMailjet = "mailjet"
	AdaptadorSMTP    = "smtp"
)

// New elige el adaptador según el valor de MAILER ("log" si está vacío). Lo llama
// main.go una sola vez, así los servicios no deciden con `if` qué adaptador usar. mailjet
// solo se usa si MAILER=mailjet.
//
// Un valor desconocido, un adaptador que no existe o una configuración incompleta falla
// al arrancar, no al enviar el primer correo.
func New(adaptador string, mailjet MailjetConfig, logger *slog.Logger) (domain.Mailer, error) {
	switch strings.ToLower(strings.TrimSpace(adaptador)) {
	case "", AdaptadorLog:
		return NewLog(logger), nil
	case AdaptadorMailjet:
		return NewMailjet(mailjet, logger)
	case AdaptadorSMTP:
		// El servidor de producción no puede abrir conexiones SMTP salientes (DigitalOcean
		// bloquea los puertos 25, 465 y 587); por eso el correo real va por una API HTTPS.
		return nil, fmt.Errorf("MAILER=%s no está disponible: el servidor bloquea SMTP; use MAILER=%s", AdaptadorSMTP, AdaptadorMailjet)
	default:
		return nil, fmt.Errorf("MAILER=%q no es válido; las opciones son %q y %q", adaptador, AdaptadorLog, AdaptadorMailjet)
	}
}

// tamano muestra un tamaño en bytes de forma legible.
func tamano(bytes int) string {
	switch {
	case bytes >= 1<<20:
		return strconv.FormatFloat(float64(bytes)/(1<<20), 'f', 1, 64) + " MB"
	case bytes >= 1<<10:
		return strconv.FormatFloat(float64(bytes)/(1<<10), 'f', 1, 64) + " KB"
	default:
		return strconv.Itoa(bytes) + " B"
	}
}
