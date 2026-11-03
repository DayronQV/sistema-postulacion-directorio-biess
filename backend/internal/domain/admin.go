package domain

import "context"

// UsuarioAdmin es un usuario del panel de administración: analista (Talento Humano) o
// administrador (Comisión de Selección). Se siembra en la base; no hay registro público.
type UsuarioAdmin struct {
	ID           string
	Usuario      string
	PasswordHash string
	Rol          Rol
}

// UsuariosAdmin es el puerto de la tabla usuarios_admin.
type UsuariosAdmin interface {
	// BuscarPorUsuario devuelve el usuario, o un error que cumple
	// errors.Is(err, ErrNotFound) si no existe.
	BuscarPorUsuario(ctx context.Context, usuario string) (UsuarioAdmin, error)
}

// FiltroPostulacionesAdmin acota la nómina del panel. Elegible nil no filtra.
type FiltroPostulacionesAdmin struct {
	Cargo     string
	Estado    string
	Elegible  *bool
	Pagina    int
	PorPagina int
}

// ResumenPostulacionAdmin es una fila de la nómina administrativa.
type ResumenPostulacionAdmin struct {
	ID        string
	Codigo    string
	Cedula    string
	Nombres   string
	Apellidos string
	Cargo     string
	Estado    string
	Elegible  bool
}

// PostulacionesAdmin es el puerto de lectura de la nómina y su detalle.
type PostulacionesAdmin interface {
	ListarAdmin(ctx context.Context, filtro FiltroPostulacionesAdmin) ([]ResumenPostulacionAdmin, int, error)
	ObtenerPorIDAdmin(ctx context.Context, postulacionID string) (Postulacion, error)
}

// DocumentosAdmin permite descargar un respaldo comprobando que pertenece a la
// postulación que se está consultando en el panel.
type DocumentosAdmin interface {
	ObtenerDePostulacion(ctx context.Context, postulacionID, documentoID string) (Documento, error)
}
