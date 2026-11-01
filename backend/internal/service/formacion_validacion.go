package service

import (
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/domain"
)

// limpiarFormacion devuelve una copia de las filas sin espacios sobrantes en los textos.
func limpiarFormacion(filas []domain.FormacionAcademica) []domain.FormacionAcademica {
	if filas == nil {
		return nil
	}
	limpias := make([]domain.FormacionAcademica, len(filas))
	for i, fila := range filas {
		fila.Nivel = strings.TrimSpace(fila.Nivel)
		fila.Institucion = strings.TrimSpace(fila.Institucion)
		fila.Tipo = strings.TrimSpace(fila.Tipo)
		fila.Titulo = strings.TrimSpace(fila.Titulo)
		fila.Area = strings.TrimSpace(fila.Area)
		fila.SenescytNumero = strings.TrimSpace(fila.SenescytNumero)
		fila.SenescytFecha = strings.TrimSpace(fila.SenescytFecha)
		fila.DocumentoID = strings.TrimSpace(fila.DocumentoID)
		limpias[i] = fila
	}
	return limpias
}

// validarFormacion agrega a fields los errores de la sección de formación académica. Los
// campos de una fila se nombran formacion[i].campo (i desde 0), como en docs/api.md. Que
// el documento sea del postulante y de tipo formacion lo comprueba el repositorio dentro
// de la transacción.
func validarFormacion(filas []domain.FormacionAcademica, hoy time.Time, fields map[string]string) {
	if len(filas) > domain.MaxFormaciones {
		fields["formacion"] = fmt.Sprintf("Máximo %d formaciones", domain.MaxFormaciones)
		return
	}
	documentos := map[string]int{} // documento → primera fila que lo usa
	for i, fila := range filas {
		campo := func(nombre string) string { return fmt.Sprintf("formacion[%d].%s", i, nombre) }
		if !domain.NivelFormacionValido(fila.Nivel) {
			fields[campo("nivel")] = "Nivel de instrucción no válido"
		}
		textoFormacion(fields, campo("institucion"), fila.Institucion, domain.MaxLongitudInstitucionFormacion)
		if !domain.TipoInstitucionValido(fila.Tipo) {
			fields[campo("tipo")] = "Tipo de institución no válido"
		}
		textoFormacion(fields, campo("titulo"), fila.Titulo, domain.MaxLongitudTituloFormacion)
		if !domain.AreaFormacionValida(fila.Area) {
			fields[campo("area")] = "Área de estudios no válida"
		}
		textoFormacion(fields, campo("senescyt_numero"), fila.SenescytNumero, domain.MaxLongitudNumeroSenescyt)
		fecha, err := time.Parse(time.DateOnly, fila.SenescytFecha)
		switch {
		case err != nil:
			fields[campo("senescyt_fecha")] = "Fecha no válida"
		case fecha.After(inicioDelDia(hoy)):
			fields[campo("senescyt_fecha")] = "La fecha no puede ser futura"
		}
		// El respaldo es opcional, pero un mismo PDF no puede respaldar dos títulos.
		if fila.DocumentoID != "" {
			if primera, repetido := documentos[fila.DocumentoID]; repetido {
				fields[campo("documento_id")] = fmt.Sprintf("El documento ya se usa en la formación %d", primera+1)
			} else {
				documentos[fila.DocumentoID] = i
			}
		}
	}
}

// textoFormacion exige un texto no vacío de a lo sumo limite caracteres.
func textoFormacion(fields map[string]string, campo, valor string, limite int) {
	switch {
	case valor == "":
		fields[campo] = "Campo obligatorio"
	case utf8.RuneCountInString(valor) > limite:
		fields[campo] = fmt.Sprintf("Máximo %d caracteres", limite)
	}
}

// inicioDelDia devuelve la medianoche de fecha en su misma zona horaria.
func inicioDelDia(fecha time.Time) time.Time {
	return time.Date(fecha.Year(), fecha.Month(), fecha.Day(), 0, 0, 0, 0, fecha.Location())
}
