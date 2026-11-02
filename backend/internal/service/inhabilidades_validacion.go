package service

import (
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/domain"
)

// limpiarInhabilidades recorta los textos del detalle y lo descarta si la pregunta 7 no
// está respondida Sí: un detalle suelto no se guarda.
func limpiarInhabilidades(i domain.Inhabilidades) domain.Inhabilidades {
	if !i.Respuestas[domain.PreguntaRemovido] {
		i.Removido = domain.RemovidoDetalle{}
		return i
	}
	i.Removido.Institucion = strings.TrimSpace(i.Removido.Institucion)
	i.Removido.Cargo = strings.TrimSpace(i.Removido.Cargo)
	i.Removido.FechaFin = strings.TrimSpace(i.Removido.FechaFin)
	i.Removido.Motivo = strings.TrimSpace(i.Removido.Motivo)
	return i
}

// validarInhabilidades exige el detalle de la pregunta 7 cuando se responde Sí. Que las 13
// preguntas estén respondidas se exige al enviar la postulación, no al guardar el borrador.
func validarInhabilidades(i domain.Inhabilidades, fields map[string]string) {
	if !i.Respuestas[domain.PreguntaRemovido] {
		return
	}
	const base = "inhabilidades.removido_detalle."
	validarTextoRemovido(fields, base+"institucion", i.Removido.Institucion, domain.MaxLongitudInstitucionRemovido, true)
	validarTextoRemovido(fields, base+"cargo", i.Removido.Cargo, domain.MaxLongitudCargoRemovido, false)
	validarTextoRemovido(fields, base+"motivo", i.Removido.Motivo, domain.MaxLongitudMotivoRemovido, true)
	if i.Removido.FechaFin != "" {
		if _, err := time.Parse(time.DateOnly, i.Removido.FechaFin); err != nil {
			fields[base+"fecha_fin"] = "Fecha no válida"
		}
	}
}

func validarTextoRemovido(fields map[string]string, campo, valor string, limite int, obligatorio bool) {
	switch {
	case valor == "" && obligatorio:
		fields[campo] = "Campo obligatorio"
	case utf8.RuneCountInString(valor) > limite:
		fields[campo] = fmt.Sprintf("Máximo %d caracteres", limite)
	}
}
