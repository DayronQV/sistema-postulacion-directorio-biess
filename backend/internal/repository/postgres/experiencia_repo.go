package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/domain"
)

// ExperienciaRepo implementa la sección experiencia_profesional dentro de la unidad de trabajo.
type ExperienciaRepo struct {
	db *pgxpool.Pool
}

var _ domain.SeccionPostulacion = (*ExperienciaRepo)(nil)

// NewExperienciaRepo crea el repositorio de experiencia profesional.
func NewExperienciaRepo(db *pgxpool.Pool) *ExperienciaRepo {
	return &ExperienciaRepo{db: db}
}

// Cargar completa f.Experiencia con las filas de la postulación en el orden de captura
// (ver idOrdenado). Solo lee: la duración de cada experiencia la calcula el servicio.
func (r *ExperienciaRepo) Cargar(ctx context.Context, postulacionID string, f *domain.Formulario) error {
	rows, err := r.db.Query(ctx, `
SELECT tipo, cargo, institucion, to_char(fecha_inicio, 'YYYY-MM-DD'),
       to_char(fecha_fin, 'YYYY-MM-DD'), actividades, descripcion, documento_id::text
FROM experiencia_profesional
WHERE postulacion_id = $1
ORDER BY id`, postulacionID)
	if err != nil {
		return fmt.Errorf("consultando la experiencia profesional: %w", err)
	}
	defer rows.Close()

	f.Experiencia = nil
	for rows.Next() {
		var fila domain.ExperienciaProfesional
		if err := rows.Scan(&fila.Tipo, &fila.Cargo, &fila.Institucion, &fila.FechaInicio,
			&fila.FechaFin, &fila.Actividades, &fila.Descripcion, &fila.DocumentoID); err != nil {
			return fmt.Errorf("leyendo una experiencia profesional: %w", err)
		}
		f.Experiencia = append(f.Experiencia, fila)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("leyendo la experiencia profesional: %w", err)
	}
	return nil
}

// Reemplazar borra la experiencia de la postulación y escribe la de f, dentro de tx. Antes
// comprueba que cada respaldo sea de esta postulación y de tipo experiencia.

func (r *ExperienciaRepo) Reemplazar(ctx context.Context, tx domain.Tx, postulacionID string, f domain.Formulario) error {
	ptx, err := pgxTx(tx)
	if err != nil {
		return err
	}
	for i, fila := range f.Experiencia {
		var propio bool
		if err := ptx.QueryRow(ctx, `
SELECT EXISTS (
  SELECT 1 FROM documentos
  WHERE id::text = $1 AND postulacion_id = $2 AND tipo = 'experiencia')`,
			fila.DocumentoID, postulacionID).Scan(&propio); err != nil {
			return fmt.Errorf("comprobando el documento de experiencia: %w", err)
		}
		if !propio {
			campo := fmt.Sprintf("experiencia[%d].documento_id", i)
			return domain.Validation("Datos inválidos", map[string]string{campo: "Referencia no válida"})
		}
	}
	if _, err := ptx.Exec(ctx, `DELETE FROM experiencia_profesional WHERE postulacion_id = $1`, postulacionID); err != nil {
		return fmt.Errorf("borrando la experiencia profesional: %w", err)
	}
	ahora := time.Now()
	for i, fila := range f.Experiencia {
		id, err := idOrdenado(ahora, i)
		if err != nil {
			return err
		}
		if _, err := ptx.Exec(ctx, `
INSERT INTO experiencia_profesional
  (id, postulacion_id, tipo, cargo, institucion, fecha_inicio, fecha_fin, actividades, descripcion, documento_id)
VALUES ($1, $2, $3, $4, $5, $6::date, $7::date, $8, $9, $10::uuid)`,
			id, postulacionID, fila.Tipo, fila.Cargo, fila.Institucion, fila.FechaInicio,
			fila.FechaFin, fila.Actividades, fila.Descripcion, fila.DocumentoID); err != nil {
			return fmt.Errorf("guardando una experiencia profesional: %w", err)
		}
	}
	return nil
}
