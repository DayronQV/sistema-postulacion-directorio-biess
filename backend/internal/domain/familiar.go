package domain

import "slices"

// Valores permitidos de la sección de familiares (EF-12, Información adicional).
const (
	InstitucionIESS  = "IESS"
	InstitucionBIESS = "BIESS"
)

// Límites de la sección de familiares.
const (
	MaxFamiliares             = 20
	MaxLongitudNombreFamiliar = 150
	MaxLongitudAreaFamiliar   = 100
	MaxLongitudCargoFamiliar  = 100
)

// Parentescos son los parentescos permitidos, en el orden en que se muestran.
var Parentescos = []string{
	"conviviente", "cónyuge", "cuñado/a", "hermano/a", "hijo político", "hijo/a",
	"madre", "nieto/a", "nuera", "padre", "primo hermano/a", "sobrino/a",
	"suegro/a", "tío/a", "yerno",
}

// ParentescoValido indica si p es uno de los parentescos permitidos.
func ParentescoValido(p string) bool {
	return slices.Contains(Parentescos, p)
}

// InstitucionFamiliarValida indica si i es IESS o BIESS.
func InstitucionFamiliarValida(i string) bool {
	return i == InstitucionIESS || i == InstitucionBIESS
}
