package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/domain"
)

// FasesRepo implementa domain.Fases sobre la tabla fases.
type FasesRepo struct {
	db *pgxpool.Pool
}

var _ domain.Fases = (*FasesRepo)(nil)

// NewFasesRepo crea el repositorio de fases.
func NewFasesRepo(db *pgxpool.Pool) *FasesRepo {
	return &FasesRepo{db: db}
}

const selectFase = `
SELECT id, nombre, orden, publicada, habilita_formulario, fecha_inicio, fecha_fin
FROM fases`

func scanFase(row pgx.Row) (domain.Fase, error) {
	var f domain.Fase
	err := row.Scan(&f.ID, &f.Nombre, &f.Orden, &f.Publicada, &f.HabilitaFormulario, &f.FechaInicio, &f.FechaFin)
	return f, err
}

// Listar devuelve las fases ordenadas.
func (r *FasesRepo) Listar(ctx context.Context) ([]domain.Fase, error) {
	rows, err := r.db.Query(ctx, selectFase+` ORDER BY orden`)
	if err != nil {
		return nil, fmt.Errorf("consultando las fases: %w", err)
	}
	defer rows.Close()
	fases := []domain.Fase{}
	for rows.Next() {
		f, err := scanFase(rows)
		if err != nil {
			return nil, fmt.Errorf("leyendo una fase: %w", err)
		}
		fases = append(fases, f)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("leyendo las fases: %w", err)
	}
	return fases, nil
}

// Obtener devuelve la fase por id.
func (r *FasesRepo) Obtener(ctx context.Context, id int) (domain.Fase, error) {
	return unaFase(r.db.QueryRow(ctx, selectFase+` WHERE id = $1`, id))
}

// PorNombre devuelve la fase por nombre.
func (r *FasesRepo) PorNombre(ctx context.Context, nombre string) (domain.Fase, error) {
	return unaFase(r.db.QueryRow(ctx, selectFase+` WHERE nombre = $1`, nombre))
}

func unaFase(row pgx.Row) (domain.Fase, error) {
	f, err := scanFase(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Fase{}, domain.NotFound("La fase no existe")
	}
	if err != nil {
		return domain.Fase{}, fmt.Errorf("consultando la fase: %w", err)
	}
	return f, nil
}

// Actualizar guarda el estado y las fechas de la fase.
func (r *FasesRepo) Actualizar(ctx context.Context, f domain.Fase) error {
	tag, err := r.db.Exec(ctx, `
UPDATE fases SET publicada = $2, habilita_formulario = $3, fecha_inicio = $4, fecha_fin = $5
WHERE id = $1`, f.ID, f.Publicada, f.HabilitaFormulario, f.FechaInicio, f.FechaFin)
	if err != nil {
		return fmt.Errorf("actualizando la fase: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.NotFound("La fase no existe")
	}
	return nil
}
