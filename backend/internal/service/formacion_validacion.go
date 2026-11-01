package service

import (
	"fmt"
	"strings"
	"time"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/domain"
)

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

func validarFormacion(filas []domain.FormacionAcademica, hoy time.Time, fields map[string]string) {
	for i, fila := range filas {
		campo := func(nombre string) string { return fmt.Sprintf("formacion[%d].%s", i, nombre) }
		if !domain.NivelFormacionValido(fila.Nivel) {
			fields[campo("nivel")] = "Nivel de instrucción no válido"
		}
		if fila.Institucion == "" {
			fields[campo("institucion")] = "Campo obligatorio"
		}
		if !domain.TipoInstitucionValido(fila.Tipo) {
			fields[campo("tipo")] = "Tipo de institución no válido"
		}
		if fila.Titulo == "" {
			fields[campo("titulo")] = "Campo obligatorio"
		}
		if !domain.AreaFormacionValida(fila.Area) {
			fields[campo("area")] = "Área de estudios no válida"
		}
		if fila.SenescytNumero == "" {
			fields[campo("senescyt_numero")] = "Campo obligatorio"
		}
		fecha, err := time.Parse(time.DateOnly, fila.SenescytFecha)
		switch {
		case err != nil:
			fields[campo("senescyt_fecha")] = "Fecha no válida"
		case fecha.After(inicioDelDia(hoy)):
			fields[campo("senescyt_fecha")] = "La fecha no puede ser futura"
		}
		if fila.DocumentoID == "" {
			fields[campo("documento_id")] = "El documento de respaldo es obligatorio"
		}
	}
}

func inicioDelDia(fecha time.Time) time.Time {
	return time.Date(fecha.Year(), fecha.Month(), fecha.Day(), 0, 0, 0, 0, fecha.Location())
}
