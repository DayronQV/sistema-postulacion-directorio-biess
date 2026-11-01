package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/domain"
)

// Restricciones UNIQUE de postulantes (nombres por defecto de PostgreSQL).
const (
	uniqueCedula = "postulantes_cedula_key"
	uniqueEmail  = "postulantes_email_key"
)

// PostulanteRepo implementa domain.Postulantes sobre la tabla postulantes.
type PostulanteRepo struct {
	db *pgxpool.Pool
}

var _ domain.Postulantes = (*PostulanteRepo)(nil)

// NewPostulanteRepo crea el repositorio de postulantes.
func NewPostulanteRepo(db *pgxpool.Pool) *PostulanteRepo {
	return &PostulanteRepo{db: db}
}

// ExisteCedula indica si la cédula ya tiene una cuenta registrada.
func (r *PostulanteRepo) ExisteCedula(ctx context.Context, cedula string) (bool, error) {
	var existe bool
	err := r.db.QueryRow(ctx,
		`SELECT EXISTS (SELECT 1 FROM postulantes WHERE cedula = $1)`, cedula,
	).Scan(&existe)
	if err != nil {
		return false, fmt.Errorf("consultando la cédula en postulantes: %w", err)
	}
	return existe, nil
}

// ExisteEmail indica si el correo ya pertenece a una cuenta.
func (r *PostulanteRepo) ExisteEmail(ctx context.Context, email string) (bool, error) {
	var existe bool
	err := r.db.QueryRow(ctx,
		`SELECT EXISTS (SELECT 1 FROM postulantes WHERE email = $1)`, email,
	).Scan(&existe)
	if err != nil {
		return false, fmt.Errorf("consultando el correo en postulantes: %w", err)
	}
	return existe, nil
}

// Crear inserta la cuenta. Una violación de UNIQUE se traduce a ErrConflict.
func (r *PostulanteRepo) Crear(ctx context.Context, p domain.NuevoPostulante) (string, error) {
	var id string
	err := r.db.QueryRow(ctx,
		`INSERT INTO postulantes (cedula, email, password_hash) VALUES ($1, $2, $3) RETURNING id`,
		p.Cedula, p.Email, p.PasswordHash,
	).Scan(&id)

	// El servicio revisa antes de insertar; esto cubre dos registros simultáneos.
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		switch pgErr.ConstraintName {
		case uniqueCedula:
			return "", domain.ConflictoCedula()
		case uniqueEmail:
			return "", domain.ConflictoEmail()
		}
	}
	if err != nil {
		return "", fmt.Errorf("insertando el postulante: %w", err)
	}
	return id, nil
}

// BuscarPorCedula devuelve la cuenta de la cédula.
func (r *PostulanteRepo) BuscarPorCedula(ctx context.Context, cedula string) (domain.Cuenta, error) {
	var c domain.Cuenta
	err := r.db.QueryRow(ctx,
		`SELECT id, cedula, password_hash FROM postulantes WHERE cedula = $1`, cedula,
	).Scan(&c.ID, &c.Cedula, &c.PasswordHash)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Cuenta{}, domain.NotFound("La cédula no tiene cuenta")
	}
	if err != nil {
		return domain.Cuenta{}, fmt.Errorf("buscando el postulante: %w", err)
	}
	return c, nil
}

// TienePostulacion indica si el postulante ya tiene una fila en postulaciones.
func (r *PostulanteRepo) TienePostulacion(ctx context.Context, postulanteID string) (bool, error) {
	var existe bool
	err := r.db.QueryRow(ctx,
		`SELECT EXISTS (SELECT 1 FROM postulaciones WHERE postulante_id = $1)`, postulanteID,
	).Scan(&existe)
	if err != nil {
		return false, fmt.Errorf("consultando la postulación: %w", err)
	}
	return existe, nil
}
