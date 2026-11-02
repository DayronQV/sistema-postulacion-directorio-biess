package domain

// Valores permitidos y límites de la sección de capacitación (EF-12, sección 8). Los
// tipos coinciden con el CHECK de capacitaciones en db/schema.sql.
const (
	TipoCurso     = "curso"
	TipoSeminario = "seminario"
	TipoTaller    = "taller"

	MaxCapacitaciones                  = 50
	MaxLongitudNombreCapacitacion      = 300
	MaxLongitudInstitucionCapacitacion = 200
	// MaxHorasCapacitacion acota las horas de un evento: sin tope, un valor fuera del rango
	// de INTEGER en PostgreSQL haría fallar el INSERT con un 500 en lugar de un 400.
	MaxHorasCapacitacion = 10000
)

// Capacitacion representa un curso, seminario o taller declarado por el postulante.
type Capacitacion struct {
	Tipo        string
	Nombre      string
	Institucion string
	FechaInicio string // AAAA-MM-DD
	FechaFin    string // AAAA-MM-DD
	Horas       int
	DocumentoID string // opcional según db/schema.sql
}

// TipoCapacitacionValido indica si tipo es curso, seminario o taller.
func TipoCapacitacionValido(tipo string) bool {
	return tipo == TipoCurso || tipo == TipoSeminario || tipo == TipoTaller
}
