// Package mailer contiene los adaptadores del puerto domain.Mailer: el simulador (Log),
// para desarrollo, y Mailjet, que envía correos reales en producción.
package mailer

import (
	"context"
	"log/slog"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/domain"
)

// Log es un Mailer que no envía nada: escribe el correo en el log estructurado. Sirve
// para desarrollar y probar sin servidor de correo; por ejemplo, el enlace para
// restablecer la contraseña se copia del log.
//
// Solo debe usarse fuera de producción: el log muestra el cuerpo completo del correo,
// que puede traer enlaces de restablecimiento.
type Log struct {
	logger *slog.Logger
}

var _ domain.Mailer = (*Log)(nil)

// NewLog crea el simulador que escribe en logger.
func NewLog(logger *slog.Logger) *Log {
	return &Log{logger: logger}
}

// Enviar valida el mensaje y lo escribe en el log. De los adjuntos se registra el
// nombre y el tamaño, no el contenido.
func (l *Log) Enviar(ctx context.Context, m domain.Mensaje) error {
	if err := m.Validar(); err != nil {
		return err
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	adjuntos := make([]string, 0, len(m.Adjuntos))
	for _, a := range m.Adjuntos {
		adjuntos = append(adjuntos, descripcionAdjunto(a))
	}

	l.logger.Info("correo simulado: no se envía",
		"para", m.Para,
		"asunto", m.Asunto,
		"texto", m.Texto,
		"html", m.HTML,
		"adjuntos", adjuntos,
	)
	return nil
}

func descripcionAdjunto(a domain.Adjunto) string {
	return a.Nombre + " (" + tamano(len(a.Contenido)) + ")"
}
