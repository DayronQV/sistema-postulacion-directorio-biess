package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/domain"
)

// FamiliaresRepo implementa domain.SeccionPostulacion para postulacion_familiares. Es
// la sección de referencia: las demás (formación, experiencia, ...) siguen el mismo
// molde en su propio archivo y se agregan al cableado sin tocar la transacción.
type FamiliaresRepo struct {
	db *pgxpool.Pool
}

var _ domain.SeccionPostulacion = (*FamiliaresRepo)(nil)

// NewFamiliaresRepo crea el repositorio de familiares.
func NewFamiliaresRepo(db *pgxpool.Pool) *FamiliaresRepo {
	return &FamiliaresRepo{db: db}
}

// Cargar completa f.Familiares con las filas de la postulación. La tabla no guarda el
// orden de captura, así que se devuelven ordenadas por nombre.
func (r *FamiliaresRepo) Cargar(ctx context.Context, postulacionID string, f *domain.Formulario) error {
	rows, err := r.db.Query(ctx, `
SELECT nombres_apellidos, parentesco, institucion, area, cargo
FROM postulacion_familiares
WHERE postulacion_id = $1
ORDER BY nombres_apellidos, id`, postulacionID)
	if err != nil {
		return fmt.Errorf("consultando los familiares: %w", err)
	}
	defer rows.Close()

	f.Familiares = nil
	for rows.Next() {
		var fam domain.Familiar
		if err := rows.Scan(&fam.NombresApellidos, &fam.Parentesco, &fam.Institucion, &fam.Area, &fam.Cargo); err != nil {
			return fmt.Errorf("leyendo un familiar: %w", err)
		}
		f.Familiares = append(f.Familiares, fam)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("leyendo los familiares: %w", err)
	}
	return nil
}

// Reemplazar borra los familiares de la postulación y escribe los de f, dentro de tx.
func (r *FamiliaresRepo) Reemplazar(ctx context.Context, tx domain.Tx, postulacionID string, f domain.Formulario) error {
	ptx, err := pgxTx(tx)
	if err != nil {
		return err
	}
	if _, err := ptx.Exec(ctx, `DELETE FROM postulacion_familiares WHERE postulacion_id = $1`, postulacionID); err != nil {
		return fmt.Errorf("borrando los familiares: %w", err)
	}
	for _, fam := range f.Familiares {
		if _, err := ptx.Exec(ctx, `
INSERT INTO postulacion_familiares (postulacion_id, nombres_apellidos, parentesco, institucion, area, cargo)
VALUES ($1, $2, $3, $4, $5, $6)`,
			postulacionID, fam.NombresApellidos, fam.Parentesco, fam.Institucion, fam.Area, fam.Cargo,
		); err != nil {
			return fmt.Errorf("guardando un familiar: %w", err)
		}
	}
	return nil
}
