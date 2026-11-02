package domain

// Valores permitidos de la sección de formación académica (EF-12, sección 5). Coinciden
// con los CHECK de formacion_academica en db/schema.sql.
const (
	NivelTercerNivel  = "tercer_nivel"
	NivelEspecialidad = "especialidad"
	NivelMaestria     = "maestria"
	NivelPhD          = "phd"

	InstitucionNacional   = "nacional"
	InstitucionExtranjera = "extranjera"

	AreaFinanzas               = "finanzas"
	AreaMercadoCapitales       = "mercado_capitales"
	AreaAdministracionEmpresas = "administracion_empresas"
	AreaEconomia               = "economia"
	AreaDerecho                = "derecho"
	AreaMateriasAfines         = "afines"
)

// Límites de la sección de formación académica.
const (
	MaxFormaciones                  = 20
	MaxLongitudInstitucionFormacion = 200
	MaxLongitudTituloFormacion      = 200
	MaxLongitudNumeroSenescyt       = 50
)

// FormacionAcademica representa un título registrado en la SENESCYT y su respaldo.
type FormacionAcademica struct {
	Nivel          string
	Institucion    string
	Tipo           string
	Titulo         string
	Area           string
	SenescytNumero string
	SenescytFecha  string // AAAA-MM-DD
	// DocumentoID es el PDF de respaldo (opcional: formacion_academica.documento_id admite
	// NULL y la especificación solo exige respaldo en Experiencia). Vacío si no hay.
	DocumentoID string
}

// NivelFormacionValido indica si nivel es uno de los niveles de instrucción permitidos.
func NivelFormacionValido(nivel string) bool {
	switch nivel {
	case NivelTercerNivel, NivelEspecialidad, NivelMaestria, NivelPhD:
		return true
	}
	return false
}

// TipoInstitucionValido indica si tipo es nacional o extranjera.
func TipoInstitucionValido(tipo string) bool {
	return tipo == InstitucionNacional || tipo == InstitucionExtranjera
}

// AreaFormacionValida indica si area es una de las áreas de estudio permitidas.
func AreaFormacionValida(area string) bool {
	switch area {
	case AreaFinanzas, AreaMercadoCapitales, AreaAdministracionEmpresas,
		AreaEconomia, AreaDerecho, AreaMateriasAfines:
		return true
	}
	return false
}
