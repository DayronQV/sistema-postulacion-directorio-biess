package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/domain"
)

// UsuariosAdminRepo implementa domain.UsuariosAdmin sobre la tabla usuarios_admin.
type UsuariosAdminRepo struct {
	db *pgxpool.Pool
}

var _ domain.UsuariosAdmin = (*UsuariosAdminRepo)(nil)

// NewUsuariosAdminRepo crea el repositorio de usuarios de administración.
func NewUsuariosAdminRepo(db *pgxpool.Pool) *UsuariosAdminRepo {
	return &UsuariosAdminRepo{db: db}
}

// BuscarPorUsuario devuelve el usuario con su hash y su rol.
func (r *UsuariosAdminRepo) BuscarPorUsuario(ctx context.Context, usuario string) (domain.UsuarioAdmin, error) {
	var (
		u   domain.UsuarioAdmin
		rol string
	)
	err := r.db.QueryRow(ctx,
		`SELECT id, usuario, password_hash, rol FROM usuarios_admin WHERE usuario = $1`, usuario,
	).Scan(&u.ID, &u.Usuario, &u.PasswordHash, &rol)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.UsuarioAdmin{}, domain.NotFound("El usuario no existe")
	}
	if err != nil {
		return domain.UsuarioAdmin{}, fmt.Errorf("buscando el usuario de administración: %w", err)
	}
	u.Rol = domain.Rol(rol)
	return u, nil
}
