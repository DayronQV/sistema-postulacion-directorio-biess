package domain

const (
	TipoCurso     = "curso"
	TipoSeminario = "seminario"
	TipoTaller    = "taller"

	MaxCapacitaciones                  = 50
	MaxLongitudNombreCapacitacion      = 300
	MaxLongitudInstitucionCapacitacion = 200
)

type Capacitacion struct {
	Tipo        string
	Nombre      string
	Institucion string
	FechaInicio string // AAAA-MM-DD
	FechaFin    string // AAAA-MM-DD
	Horas       int
	DocumentoID string // opcional según db/schema.sql
}

func TipoCapacitacionValido(tipo string) bool {
	return tipo == TipoCurso || tipo == TipoSeminario || tipo == TipoTaller
}
