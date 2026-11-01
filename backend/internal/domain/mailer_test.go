package domain_test

import (
	"errors"
	"testing"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/domain"
)

func mensajeValido() domain.Mensaje {
	return domain.Mensaje{
		Para:   "postulante@correo.com",
		Asunto: "Su postulación fue enviada",
		Texto:  "Adjuntamos su formulario.",
		Adjuntos: []domain.Adjunto{
			{Nombre: "postulacion.pdf", Contenido: []byte("%PDF-1.4 contenido")},
		},
	}
}

func TestMensajeValido(t *testing.T) {
	if err := mensajeValido().Validar(); err != nil {
		t.Fatalf("un mensaje válido no debe dar error: %v", err)
	}

	soloHTML := domain.Mensaje{Para: "a@b.com", Asunto: "Hola", HTML: "<p>Hola</p>"}
	if err := soloHTML.Validar(); err != nil {
		t.Errorf("un mensaje solo con HTML debe ser válido: %v", err)
	}

	sinAdjuntos := domain.Mensaje{Para: "a@b.com", Asunto: "Hola", Texto: "Hola"}
	if err := sinAdjuntos.Validar(); err != nil {
		t.Errorf("los adjuntos son opcionales: %v", err)
	}
}

func TestMensajeInvalido(t *testing.T) {
	casos := []struct {
		nombre string
		cambia func(*domain.Mensaje)
		campo  string // campo de `fields` que debe aparecer
		texto  string // si no está vacío, el mensaje exacto de ese campo
	}{
		{"sin destinatario", func(m *domain.Mensaje) { m.Para = "" }, "para", "Campo obligatorio"},
		{"destinatario solo con espacios", func(m *domain.Mensaje) { m.Para = "   " }, "para", "Campo obligatorio"},
		{"destinatario con espacios alrededor", func(m *domain.Mensaje) { m.Para = " a@b.com " }, "para", "Correo electrónico no válido"},
		{"destinatario con salto de línea al final", func(m *domain.Mensaje) { m.Para = "a@b.com\r\n" }, "para", "Correo electrónico no válido"},
		{"destinatario con salto de línea simple al final", func(m *domain.Mensaje) { m.Para = "a@b.com\n" }, "para", "Correo electrónico no válido"},
		{"destinatario sin arroba", func(m *domain.Mensaje) { m.Para = "postulante.correo.com" }, "para", ""},
		{"destinatario con nombre", func(m *domain.Mensaje) { m.Para = "Juan <a@b.com>" }, "para", ""},
		{"varios destinatarios", func(m *domain.Mensaje) { m.Para = "a@b.com, c@d.com" }, "para", ""},
		{"destinatario con salto de línea", func(m *domain.Mensaje) { m.Para = "a@b.com\r\nBcc: otro@x.com" }, "para", ""},
		{"sin asunto", func(m *domain.Mensaje) { m.Asunto = "" }, "asunto", "Campo obligatorio"},
		{"asunto con salto de línea (inyección de cabeceras)", func(m *domain.Mensaje) { m.Asunto = "Hola\r\nBcc: otro@x.com" }, "asunto", ""},
		{"asunto con salto de línea simple", func(m *domain.Mensaje) { m.Asunto = "Hola\nmundo" }, "asunto", ""},
		{"sin cuerpo", func(m *domain.Mensaje) { m.Texto, m.HTML = "", "" }, "cuerpo", ""},
		{"cuerpo solo con espacios", func(m *domain.Mensaje) { m.Texto, m.HTML = "  \n ", " " }, "cuerpo", ""},
		{"adjunto sin nombre", func(m *domain.Mensaje) { m.Adjuntos[0].Nombre = "" }, "adjuntos[0].nombre", ""},
		{"adjunto con nombre que rompe cabecera", func(m *domain.Mensaje) { m.Adjuntos[0].Nombre = "a.pdf\r\nX: y" }, "adjuntos[0].nombre", ""},
		{"adjunto vacío", func(m *domain.Mensaje) { m.Adjuntos[0].Contenido = nil }, "adjuntos[0].contenido", ""},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			m := mensajeValido()
			c.cambia(&m)

			err := m.Validar()
			if !errors.Is(err, domain.ErrValidation) {
				t.Fatalf("se esperaba ErrValidation, got %v", err)
			}
			var de *domain.Error
			if !errors.As(err, &de) || de.Fields[c.campo] == "" {
				t.Fatalf("falta el campo %q en fields: %+v", c.campo, de)
			}
			if c.texto != "" && de.Fields[c.campo] != c.texto {
				t.Fatalf("fields[%q] = %q, se esperaba %q", c.campo, de.Fields[c.campo], c.texto)
			}
		})
	}
}

func TestMensajeInvalidoListaTodosLosErrores(t *testing.T) {
	m := domain.Mensaje{Adjuntos: []domain.Adjunto{{Nombre: "ok.pdf"}, {Nombre: ""}}}
	err := m.Validar()
	var de *domain.Error
	if !errors.As(err, &de) {
		t.Fatalf("se esperaba un error de dominio, got %v", err)
	}
	for _, campo := range []string{"para", "asunto", "cuerpo", "adjuntos[0].contenido", "adjuntos[1].nombre"} {
		if de.Fields[campo] == "" {
			t.Errorf("falta el campo %q: el error debe listar todo lo que está mal, no solo lo primero (%v)", campo, de.Fields)
		}
	}
}
