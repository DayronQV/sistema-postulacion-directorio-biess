package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/domain"
)

var _ domain.Meritos = (*PostulacionRepo)(nil)

// EstadoPorID devuelve el estado de la postulación.
func (r *PostulacionRepo) EstadoPorID(ctx context.Context, postulacionID string) (string, error) {
	var estado string
	err := r.db.QueryRow(ctx, `SELECT estado FROM postulaciones WHERE id = $1`, postulacionID).Scan(&estado)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", domain.NotFound("La postulación no existe")
	}
	if err != nil {
		return "", fmt.Errorf("consultando el estado de la postulación: %w", err)
	}
	return estado, nil
}

// ActualizarMerito escribe supero_merito solo si la postulación sigue enviada: la condición
// va en el mismo UPDATE, así que un cambio de estado concurrente no se pisa.
func (r *PostulacionRepo) ActualizarMerito(ctx context.Context, postulacionID string, supero bool) (bool, error) {
	tag, err := r.db.Exec(ctx, `
UPDATE postulaciones SET supero_merito = $2, updated_at = now()
WHERE id = $1 AND estado = $3`, postulacionID, supero, domain.EstadoEnviada)
	if err != nil {
		return false, fmt.Errorf("marcando el mérito de la postulación: %w", err)
	}
	return tag.RowsAffected() == 1, nil
}
