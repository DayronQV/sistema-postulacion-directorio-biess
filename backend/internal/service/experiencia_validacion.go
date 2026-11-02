package service

import (
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/domain"
)

func limpiarExperiencia(filas []domain.ExperienciaProfesional) []domain.ExperienciaProfesional {
	if filas == nil {
		return nil
	}
	limpias := make([]domain.ExperienciaProfesional, len(filas))
	for i, fila := range filas {
		fila.Tipo = strings.TrimSpace(fila.Tipo)
		fila.Cargo = strings.TrimSpace(fila.Cargo)
		fila.Institucion = strings.TrimSpace(fila.Institucion)
		fila.FechaInicio = strings.TrimSpace(fila.FechaInicio)
		fila.FechaFin = strings.TrimSpace(fila.FechaFin)
		fila.Actividades = strings.TrimSpace(fila.Actividades)
		fila.Descripcion = strings.TrimSpace(fila.Descripcion)
		fila.DocumentoID = strings.TrimSpace(fila.DocumentoID)
		fila.Duracion = domain.Duracion{}
		limpias[i] = fila
	}
	return limpias
}

func validarExperiencia(filas []domain.ExperienciaProfesional, hoy time.Time, fields map[string]string) {
	if len(filas) > domain.MaxExperiencias {
		fields["experiencia"] = fmt.Sprintf("Máximo %d experiencias", domain.MaxExperiencias)
		return
	}
	documentos := map[string]int{}
	for i, fila := range filas {
		campo := func(nombre string) string { return fmt.Sprintf("experiencia[%d].%s", i, nombre) }
		if !domain.TipoExperienciaValido(fila.Tipo) {
			fields[campo("tipo")] = "Tipo de experiencia no válido"
		}
		validarTextoExperiencia(fields, campo("cargo"), fila.Cargo, domain.MaxLongitudCargoExperiencia)
		validarTextoExperiencia(fields, campo("institucion"), fila.Institucion, domain.MaxLongitudInstitucionExperiencia)
		if !domain.ActividadExperienciaValida(fila.Actividades) {
			fields[campo("actividades")] = "Actividad relacionada no válida"
		}
		validarTextoExperiencia(fields, campo("descripcion"), fila.Descripcion, domain.MaxLongitudDescripcionExperiencia)

		inicio, errInicio := time.Parse(time.DateOnly, fila.FechaInicio)
		fin, errFin := time.Parse(time.DateOnly, fila.FechaFin)
		if errInicio != nil {
			fields[campo("fecha_inicio")] = "Fecha no válida"
		} else if inicio.After(fechaSinHora(hoy)) {
			fields[campo("fecha_inicio")] = "La fecha no puede ser futura"
		}
		if errFin != nil {
			fields[campo("fecha_fin")] = "Fecha no válida"
		} else if fin.After(fechaSinHora(hoy)) {
			fields[campo("fecha_fin")] = "La fecha no puede ser futura"
		}
		if errInicio == nil && errFin == nil && fin.Before(inicio) {
			fields[campo("fecha_fin")] = "La fecha final no puede ser anterior a la inicial"
		}

		if fila.DocumentoID == "" {
			fields[campo("documento_id")] = "Documento de respaldo obligatorio"
		} else if primera, repetido := documentos[fila.DocumentoID]; repetido {
			fields[campo("documento_id")] = fmt.Sprintf("El documento ya se usa en la experiencia %d", primera+1)
		} else {
			documentos[fila.DocumentoID] = i
		}
	}
}

func fechaSinHora(fecha time.Time) time.Time {
	return time.Date(fecha.Year(), fecha.Month(), fecha.Day(), 0, 0, 0, 0, fecha.Location())
}

func validarTextoExperiencia(fields map[string]string, campo, valor string, limite int) {
	switch {
	case valor == "":
		fields[campo] = "Campo obligatorio"
	case utf8.RuneCountInString(valor) > limite:
		fields[campo] = fmt.Sprintf("Máximo %d caracteres", limite)
	}
}
