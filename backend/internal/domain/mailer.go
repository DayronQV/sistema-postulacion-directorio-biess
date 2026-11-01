package domain

import (
	"context"
	"net/mail"
	"strconv"
	"strings"
)

// Adjunto es un archivo que acompaña a un correo, por ejemplo el PDF de la postulación
// o el reporte del examen.
type Adjunto struct {
	Nombre    string
	Contenido []byte
}

// Mensaje es un correo por enviar. Debe traer el cuerpo en texto, en HTML o en ambos.
type Mensaje struct {
	Para     string // una sola dirección, por ejemplo "postulante@correo.com"
	Asunto   string
	Texto    string
	HTML     string
	Adjuntos []Adjunto
}

// Mailer es el puerto para enviar correos (recuperación de contraseña, envío de la
// postulación, resultado de la evaluación). Hoy se simula escribiendo en el log
// (internal/integration/mailer); en producción se cambia por un adaptador SMTP sin
// tocar los servicios que lo usan.
type Mailer interface {
	// Enviar valida el mensaje y lo envía. Un mensaje inválido devuelve un error que
	// cumple errors.Is(err, ErrValidation).
	Enviar(ctx context.Context, m Mensaje) error
}

// Validar comprueba el mensaje antes de enviarlo. Todos los adaptadores deben llamarlo,
// para que se rechacen los mismos mensajes sin importar cómo se envíen. Devuelve un
// error de validación (400) con `fields`, en el formato de docs/api.md.
func (m Mensaje) Validar() error {
	campos := map[string]string{}

	switch {
	case strings.TrimSpace(m.Para) == "":
		campos["para"] = "Campo obligatorio"
	case !direccionSimple(m.Para):
		campos["para"] = "Correo electrónico no válido"
	}

	switch {
	case strings.TrimSpace(m.Asunto) == "":
		campos["asunto"] = "Campo obligatorio"
	case tieneSaltoDeLinea(m.Asunto):
		// Un salto de línea en una cabecera permitiría inyectar otras cabeceras.
		campos["asunto"] = "No puede contener saltos de línea"
	}

	if strings.TrimSpace(m.Texto) == "" && strings.TrimSpace(m.HTML) == "" {
		campos["cuerpo"] = "El correo necesita un texto o un HTML"
	}

	for i, a := range m.Adjuntos {
		switch {
		case strings.TrimSpace(a.Nombre) == "" || tieneSaltoDeLinea(a.Nombre):
			campos[campoAdjunto(i, "nombre")] = "Nombre de archivo no válido"
		case len(a.Contenido) == 0:
			campos[campoAdjunto(i, "contenido")] = "El archivo está vacío"
		}
	}

	if len(campos) > 0 {
		return Validation("Datos inválidos", campos)
	}
	return nil
}

// direccionSimple acepta solo una dirección ("a@b.com") escrita exactamente así:
// rechaza nombres con formato ("Juan <a@b.com>"), listas de destinatarios, espacios
// alrededor y saltos de línea, incluso al final. No se recorta nada a propósito: lo que
// se valida es lo mismo que luego se escribe en la cabecera del correo.
func direccionSimple(s string) bool {
	addr, err := mail.ParseAddress(s)
	return err == nil && addr.Address == s
}

func tieneSaltoDeLinea(s string) bool {
	return strings.ContainsAny(s, "\r\n")
}

func campoAdjunto(i int, campo string) string {
	return "adjuntos[" + strconv.Itoa(i) + "]." + campo
}
