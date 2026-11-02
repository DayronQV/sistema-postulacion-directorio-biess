package domain

const (
	TipoExperienciaGeneral   = "general"
	TipoExperienciaDireccion = "direccion"

	ActividadFinanzas               = "finanzas"
	ActividadMercadoCapitales       = "mercado_capitales"
	ActividadAdministracionEmpresas = "administracion_empresas"
	ActividadEconomia               = "economia"
	ActividadDerecho                = "derecho"
	ActividadMateriasAfines         = "afines"

	MaxExperiencias                   = 50
	MaxLongitudCargoExperiencia       = 200
	MaxLongitudInstitucionExperiencia = 200
	MaxLongitudDescripcionExperiencia = 1000
)

// ExperienciaProfesional representa un periodo laboral y el respaldo obligatorio.
type ExperienciaProfesional struct {
	Tipo        string
	Cargo       string
	Institucion string
	FechaInicio string // AAAA-MM-DD
	FechaFin    string // AAAA-MM-DD
	Actividades string
	Descripcion string
	DocumentoID string
	Duracion    Duracion
}

func TipoExperienciaValido(tipo string) bool {
	return tipo == TipoExperienciaGeneral || tipo == TipoExperienciaDireccion
}

func ActividadExperienciaValida(actividad string) bool {
	switch actividad {
	case ActividadFinanzas, ActividadMercadoCapitales, ActividadAdministracionEmpresas,
		ActividadEconomia, ActividadDerecho, ActividadMateriasAfines:
		return true
	}
	return false
}
