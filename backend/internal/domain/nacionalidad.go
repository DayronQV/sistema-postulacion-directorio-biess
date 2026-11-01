package domain

import (
	"strings"
	"unicode"
)

// sinTildes pliega las vocales con tilde o diéresis y la ñ a su letra base. Alcanza para
// los gentilicios en español; evita depender de golang.org/x/text solo para esto.
var sinTildes = strings.NewReplacer(
	"á", "a", "é", "e", "í", "i", "ó", "o", "ú", "u", "ü", "u", "ñ", "n",
)

// normalizarTexto deja un texto comparable: sin espacios sobrantes (también internos),
// en minúsculas y sin tildes.
func normalizarTexto(s string) string {
	return sinTildes.Replace(strings.ToLower(strings.Join(strings.FieldsFunc(s, unicode.IsSpace), " ")))
}

// EsNacionalidadEcuatoriana indica si la nacionalidad que informa el Registro Civil es la
// ecuatoriana. Ignora mayúsculas, tildes y espacios sobrantes, y acepta las dos
// concordancias de género ("Ecuatoriana" y "Ecuatoriano").
func EsNacionalidadEcuatoriana(nacionalidad string) bool {
	switch normalizarTexto(nacionalidad) {
	case "ecuatoriana", "ecuatoriano":
		return true
	}
	return false
}
