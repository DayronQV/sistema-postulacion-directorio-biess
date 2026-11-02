package domain

// Valores permitidos y límites de la sección de publicaciones (EF-12, sección 7). Los
// tipos coinciden con el CHECK de publicaciones en db/schema.sql.
const (
	TipoPublicacion   = "publicacion"
	TipoInvestigacion = "investigacion"

	MaxPublicaciones             = 50
	MaxLongitudTituloPublicacion = 300
	MaxLongitudMedioPublicacion  = 200
)

// PublicacionAcademica representa una publicación o investigación declarada.
type PublicacionAcademica struct {
	Titulo      string
	Tipo        string
	Medio       string
	Fecha       string // AAAA-MM-DD
	Relacionada bool
	DocumentoID string
}

// TipoPublicacionValido indica si tipo es publicación o trabajo de investigación.
func TipoPublicacionValido(tipo string) bool {
	return tipo == TipoPublicacion || tipo == TipoInvestigacion
}
