package service

import (
	"fmt"
	"net/mail"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/domain"
)

// telefonoReferencia admite de 7 a 15 dígitos, con un "+" inicial opcional.
var telefonoReferencia = regexp.MustCompile(`^\+?[0-9]{7,15}$`)

// limpiarReferencias devuelve una copia de las filas sin espacios sobrantes en los textos.
func limpiarReferencias(referencias []domain.ReferenciaPersonal) []domain.ReferenciaPersonal {
	if referencias == nil {
		return nil
	}
	limpias := make([]domain.ReferenciaPersonal, len(referencias))
	for i, r := range referencias {
		limpias[i] = domain.ReferenciaPersonal{
			NombresApellidos: strings.TrimSpace(r.NombresApellidos),
			Relacion:         strings.TrimSpace(r.Relacion),
			Telefono:         strings.TrimSpace(r.Telefono),
			Correo:           strings.TrimSpace(r.Correo),
		}
	}
	return limpias
}

// validarReferencias agrega a fields los errores de la sección de referencias personales.
// Los campos de una fila se nombran referencias[i].campo (i desde 0), como en docs/api.md.
// En borrador la lista puede estar vacía para guardar el progreso; el mínimo de una
// referencia se exige al enviar la postulación (docs/especificacion.md).
func validarReferencias(f domain.Formulario, fields map[string]string) {
	if len(f.Referencias) > domain.MaxReferencias {
		fields["referencias"] = fmt.Sprintf("Máximo %d referencias", domain.MaxReferencias)
		return
	}
	for i, r := range f.Referencias {
		campo := func(n string) string { return fmt.Sprintf("referencias[%d].%s", i, n) }
		textoReferencia(fields, campo("nombres_apellidos"), r.NombresApellidos, domain.MaxLongitudNombreReferencia)
		textoReferencia(fields, campo("relacion"), r.Relacion, domain.MaxLongitudRelacionReferencia)
		switch {
		case r.Telefono == "":
			fields[campo("telefono")] = "Campo obligatorio"
		case !telefonoReferencia.MatchString(r.Telefono):
			fields[campo("telefono")] = "Teléfono no válido"
		}
		switch {
		case r.Correo == "":
			fields[campo("correo")] = "Campo obligatorio"
		case !correoValido(r.Correo):
			fields[campo("correo")] = "Correo no válido"
		}
	}
}

// correoValido acepta solo una dirección simple (usuario@dominio), sin nombre visible.
func correoValido(correo string) bool {
	direccion, err := mail.ParseAddress(correo)
	return err == nil && direccion.Address == correo && utf8.RuneCountInString(correo) <= domain.MaxLongitudCorreoReferencia
}

// textoReferencia exige un texto no vacío de a lo sumo limite caracteres.
func textoReferencia(fields map[string]string, campo, valor string, limite int) {
	switch {
	case valor == "":
		fields[campo] = "Campo obligatorio"
	case utf8.RuneCountInString(valor) > limite:
		fields[campo] = fmt.Sprintf("Máximo %d caracteres", limite)
	}
}
