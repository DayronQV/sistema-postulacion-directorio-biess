package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/domain"
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
