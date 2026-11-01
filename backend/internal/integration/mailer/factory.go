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
	AdaptadorLog  = "log"
	AdaptadorSMTP = "smtp"
)

// New elige el adaptador según el valor de MAILER ("log" si está vacío). Lo llama
// main.go una sola vez, así los servicios no deciden con `if` qué adaptador usar.
//
// Un valor desconocido o un adaptador que todavía no existe falla al arrancar, no al
// enviar el primer correo.
func New(adaptador string, logger *slog.Logger) (domain.Mailer, error) {
	switch strings.ToLower(strings.TrimSpace(adaptador)) {
	case "", AdaptadorLog:
		return NewLog(logger), nil
	case AdaptadorSMTP:
		// Punto de extensión: aquí se crea el adaptador SMTP en la issue #52.
		return nil, fmt.Errorf("MAILER=%s todavía no está disponible (issue #52); use MAILER=%s", AdaptadorSMTP, AdaptadorLog)
	default:
		return nil, fmt.Errorf("MAILER=%q no es válido; las opciones son %q y %q", adaptador, AdaptadorLog, AdaptadorSMTP)
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
