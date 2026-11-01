package service

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/domain"
)

// limpiarFamiliares devuelve una copia de las filas sin espacios sobrantes en los textos.
func limpiarFamiliares(familiares []domain.Familiar) []domain.Familiar {
	if familiares == nil {
		return nil
	}
	limpios := make([]domain.Familiar, len(familiares))
	for i, fam := range familiares {
		limpios[i] = domain.Familiar{
			NombresApellidos: strings.TrimSpace(fam.NombresApellidos),
			Parentesco:       strings.TrimSpace(fam.Parentesco),
			Institucion:      strings.TrimSpace(fam.Institucion),
			Area:             strings.TrimSpace(fam.Area),
			Cargo:            strings.TrimSpace(fam.Cargo),
		}
	}
	return limpios
}

// validarFamiliares agrega a fields los errores de la sección de familiares. Los campos
// de una fila se nombran familiares[i].campo (i desde 0), como en docs/api.md.
func validarFamiliares(f domain.Formulario, fields map[string]string) {
	switch {
	case !f.TieneFamiliares && len(f.Familiares) > 0:
		fields["familiares"] = "Hay familiares pero se indicó que no tiene"
		return
	case f.TieneFamiliares && len(f.Familiares) == 0:
		fields["familiares"] = "Agregue al menos un familiar"
		return
	case len(f.Familiares) > domain.MaxFamiliares:
		fields["familiares"] = fmt.Sprintf("Máximo %d familiares", domain.MaxFamiliares)
		return
	}
	for i, fam := range f.Familiares {
		campo := func(n string) string { return fmt.Sprintf("familiares[%d].%s", i, n) }
		textoFamiliar(fields, campo("nombres_apellidos"), fam.NombresApellidos, domain.MaxLongitudNombreFamiliar)
		switch {
		case fam.Parentesco == "":
			fields[campo("parentesco")] = "Campo obligatorio"
		case !domain.ParentescoValido(fam.Parentesco):
			fields[campo("parentesco")] = "Valor no permitido"
		}
		switch {
		case fam.Institucion == "":
			fields[campo("institucion")] = "Campo obligatorio"
		case !domain.InstitucionFamiliarValida(fam.Institucion):
			fields[campo("institucion")] = "Valor no permitido"
		}
		textoFamiliar(fields, campo("area"), fam.Area, domain.MaxLongitudAreaFamiliar)
		textoFamiliar(fields, campo("cargo"), fam.Cargo, domain.MaxLongitudCargoFamiliar)
	}
}

// textoFamiliar exige un texto no vacío de a lo sumo limite caracteres.
func textoFamiliar(fields map[string]string, campo, valor string, limite int) {
	switch {
	case valor == "":
		fields[campo] = "Campo obligatorio"
	case utf8.RuneCountInString(valor) > limite:
		fields[campo] = fmt.Sprintf("Máximo %d caracteres", limite)
	}
}
