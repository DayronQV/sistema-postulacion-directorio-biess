package domain

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

func TipoPublicacionValido(tipo string) bool {
	return tipo == TipoPublicacion || tipo == TipoInvestigacion
}
