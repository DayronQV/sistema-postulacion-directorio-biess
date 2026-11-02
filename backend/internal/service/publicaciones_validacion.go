package service

import (
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/domain"
)

func limpiarPublicaciones(filas []domain.PublicacionAcademica) []domain.PublicacionAcademica {
	if filas == nil {
		return nil
	}
	limpias := make([]domain.PublicacionAcademica, len(filas))
	for i, fila := range filas {
		fila.Titulo = strings.TrimSpace(fila.Titulo)
		fila.Tipo = strings.TrimSpace(fila.Tipo)
		fila.Medio = strings.TrimSpace(fila.Medio)
		fila.Fecha = strings.TrimSpace(fila.Fecha)
		fila.DocumentoID = strings.TrimSpace(fila.DocumentoID)
		limpias[i] = fila
	}
	return limpias
}

func validarPublicaciones(f domain.Formulario, hoy time.Time, fields map[string]string) {
	if !f.TienePublicaciones {
		if len(f.Publicaciones) > 0 {
			fields["publicaciones"] = "No envíe publicaciones si la respuesta es No"
		}
		return
	}
	if len(f.Publicaciones) == 0 {
		fields["publicaciones"] = "Agregue al menos una publicación o investigación"
		return
	}
	if len(f.Publicaciones) > domain.MaxPublicaciones {
		fields["publicaciones"] = fmt.Sprintf("Máximo %d publicaciones", domain.MaxPublicaciones)
		return
	}
	documentos := map[string]int{}
	for i, fila := range f.Publicaciones {
		campo := func(nombre string) string { return fmt.Sprintf("publicaciones[%d].%s", i, nombre) }
		validarTextoPublicacion(fields, campo("titulo"), fila.Titulo, domain.MaxLongitudTituloPublicacion)
		if !domain.TipoPublicacionValido(fila.Tipo) {
			fields[campo("tipo")] = "Tipo no válido"
		}
		validarTextoPublicacion(fields, campo("medio"), fila.Medio, domain.MaxLongitudMedioPublicacion)
		fecha, err := time.Parse(time.DateOnly, fila.Fecha)
		switch {
		case err != nil:
			fields[campo("fecha")] = "Fecha no válida"
		case fecha.After(time.Date(hoy.Year(), hoy.Month(), hoy.Day(), 0, 0, 0, 0, hoy.Location())):
			fields[campo("fecha")] = "La fecha no puede ser futura"
		}
		if fila.DocumentoID == "" {
			fields[campo("documento_id")] = "Documento de respaldo obligatorio"
		} else if primera, repetido := documentos[fila.DocumentoID]; repetido {
			fields[campo("documento_id")] = fmt.Sprintf("El documento ya se usa en la publicación %d", primera+1)
		} else {
			documentos[fila.DocumentoID] = i
		}
	}
}

func validarTextoPublicacion(fields map[string]string, campo, valor string, limite int) {
	switch {
	case valor == "":
		fields[campo] = "Campo obligatorio"
	case utf8.RuneCountInString(valor) > limite:
		fields[campo] = fmt.Sprintf("Máximo %d caracteres", limite)
	}
}
