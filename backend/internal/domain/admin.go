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
