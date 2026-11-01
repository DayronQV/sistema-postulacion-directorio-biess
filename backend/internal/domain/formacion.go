package domain

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

// FormacionAcademica representa un título registrado en la SENESCYT y su respaldo.
type FormacionAcademica struct {
	Nivel          string
	Institucion    string
	Tipo           string
	Titulo         string
	Area           string
	SenescytNumero string
	SenescytFecha  string // AAAA-MM-DD
	DocumentoID    string
}

func NivelFormacionValido(nivel string) bool {
	switch nivel {
	case NivelTercerNivel, NivelEspecialidad, NivelMaestria, NivelPhD:
		return true
	}
	return false
}

func TipoInstitucionValido(tipo string) bool {
	return tipo == InstitucionNacional || tipo == InstitucionExtranjera
}

func AreaFormacionValida(area string) bool {
	switch area {
	case AreaFinanzas, AreaMercadoCapitales, AreaAdministracionEmpresas,
		AreaEconomia, AreaDerecho, AreaMateriasAfines:
		return true
	}
	return false
}
