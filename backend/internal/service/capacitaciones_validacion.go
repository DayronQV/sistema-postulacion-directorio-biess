package service

import (
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/domain"
)

func limpiarCapacitaciones(filas []domain.Capacitacion) []domain.Capacitacion {
	if filas == nil {
		return nil
	}
	limpias := make([]domain.Capacitacion, len(filas))
	for i, fila := range filas {
		fila.Tipo = strings.TrimSpace(fila.Tipo)
		fila.Nombre = strings.TrimSpace(fila.Nombre)
		fila.Institucion = strings.TrimSpace(fila.Institucion)
		fila.FechaInicio = strings.TrimSpace(fila.FechaInicio)
		fila.FechaFin = strings.TrimSpace(fila.FechaFin)
		fila.DocumentoID = strings.TrimSpace(fila.DocumentoID)
		limpias[i] = fila
	}
	return limpias
}

func validarCapacitaciones(filas []domain.Capacitacion, hoy time.Time, fields map[string]string) {
	if len(filas) > domain.MaxCapacitaciones {
		fields["capacitaciones"] = fmt.Sprintf("Máximo %d capacitaciones", domain.MaxCapacitaciones)
		return
	}
	documentos := map[string]int{}
	for i, fila := range filas {
		campo := func(nombre string) string { return fmt.Sprintf("capacitaciones[%d].%s", i, nombre) }
		if !domain.TipoCapacitacionValido(fila.Tipo) {
			fields[campo("tipo")] = "Tipo de evento no válido"
		}
		validarTextoCapacitacion(fields, campo("nombre"), fila.Nombre, domain.MaxLongitudNombreCapacitacion)
		validarTextoCapacitacion(fields, campo("institucion"), fila.Institucion, domain.MaxLongitudInstitucionCapacitacion)
		inicio, errInicio := time.Parse(time.DateOnly, fila.FechaInicio)
		fin, errFin := time.Parse(time.DateOnly, fila.FechaFin)
		limite := time.Date(hoy.Year(), hoy.Month(), hoy.Day(), 0, 0, 0, 0, hoy.Location())
		if errInicio != nil {
			fields[campo("fecha_inicio")] = "Fecha no válida"
		} else if inicio.After(limite) {
			fields[campo("fecha_inicio")] = "La fecha no puede ser futura"
		}
		if errFin != nil {
			fields[campo("fecha_fin")] = "Fecha no válida"
		} else if fin.After(limite) {
			fields[campo("fecha_fin")] = "La fecha no puede ser futura"
		}
		if errInicio == nil && errFin == nil && fin.Before(inicio) {
			fields[campo("fecha_fin")] = "La fecha final no puede ser anterior a la inicial"
		}
		if fila.Horas <= 0 {
			fields[campo("horas")] = "Las horas deben ser mayores a cero"
		}
		if fila.DocumentoID != "" {
			if primera, repetido := documentos[fila.DocumentoID]; repetido {
				fields[campo("documento_id")] = fmt.Sprintf("El documento ya se usa en la capacitación %d", primera+1)
			} else {
				documentos[fila.DocumentoID] = i
			}
		}
	}
}

func validarTextoCapacitacion(fields map[string]string, campo, valor string, limite int) {
	if valor == "" {
		fields[campo] = "Campo obligatorio"
	} else if utf8.RuneCountInString(valor) > limite {
		fields[campo] = fmt.Sprintf("Máximo %d caracteres", limite)
	}
}
