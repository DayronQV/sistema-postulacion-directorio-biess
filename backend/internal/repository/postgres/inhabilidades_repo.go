package postgres

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/domain"
)

// InhabilidadesRepo implementa domain.SeccionPostulacion para la tabla
// postulacion_inhabilidades (una fila por pregunta respondida) y guarda en
// postulaciones.elegible el resultado que calculó el servicio.
type InhabilidadesRepo struct {
	db *pgxpool.Pool
}

var _ domain.SeccionPostulacion = (*InhabilidadesRepo)(nil)

// NewInhabilidadesRepo crea el repositorio de inhabilidades.
func NewInhabilidadesRepo(db *pgxpool.Pool) *InhabilidadesRepo {
	return &InhabilidadesRepo{db: db}
}

// detalleRemovido es el JSON de postulacion_inhabilidades.detalle de la pregunta 7.
type detalleRemovido struct {
	Institucion string `json:"institucion"`
	Cargo       string `json:"cargo"`
	FechaFin    string `json:"fecha_fin"`
	Motivo      string `json:"motivo"`
}

// Cargar completa f.Inhabilidades y f.Elegible con lo guardado de la postulación.
func (r *InhabilidadesRepo) Cargar(ctx context.Context, postulacionID string, f *domain.Formulario) error {
	if err := r.db.QueryRow(ctx,
		`SELECT elegible FROM postulaciones WHERE id = $1`, postulacionID,
	).Scan(&f.Elegible); err != nil {
		return fmt.Errorf("consultando la elegibilidad: %w", err)
	}
	rows, err := r.db.Query(ctx, `
SELECT pregunta, respuesta, detalle
FROM postulacion_inhabilidades
WHERE postulacion_id = $1
ORDER BY pregunta`, postulacionID)
	if err != nil {
		return fmt.Errorf("consultando las inhabilidades: %w", err)
	}
	defer rows.Close()
	f.Inhabilidades = domain.Inhabilidades{Respuestas: map[int]bool{}}
	for rows.Next() {
		var (
			pregunta  int
			respuesta bool
			detalle   []byte
		)
		if err := rows.Scan(&pregunta, &respuesta, &detalle); err != nil {
			return fmt.Errorf("leyendo una inhabilidad: %w", err)
		}
		f.Inhabilidades.Respuestas[pregunta] = respuesta
		if pregunta == domain.PreguntaRemovido && len(detalle) > 0 {
			var d detalleRemovido
			if err := json.Unmarshal(detalle, &d); err != nil {
				return fmt.Errorf("leyendo el detalle de la pregunta %d: %w", pregunta, err)
			}
			f.Inhabilidades.Removido = domain.RemovidoDetalle(d)
		}
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("leyendo las inhabilidades: %w", err)
	}
	return nil
}

// Reemplazar borra las respuestas de la postulación, escribe las de f y actualiza
// elegible, todo dentro de tx.
func (r *InhabilidadesRepo) Reemplazar(ctx context.Context, tx domain.Tx, postulacionID string, f domain.Formulario) error {
	ptx, err := pgxTx(tx)
	if err != nil {
		return err
	}
	if _, err := ptx.Exec(ctx, `DELETE FROM postulacion_inhabilidades WHERE postulacion_id = $1`, postulacionID); err != nil {
		return fmt.Errorf("borrando las inhabilidades: %w", err)
	}
	for pregunta := 1; pregunta <= domain.TotalPreguntasInhabilidades; pregunta++ {
		respuesta, respondida := f.Inhabilidades.Respuestas[pregunta]
		if !respondida {
			continue
		}
		var detalle []byte
		if pregunta == domain.PreguntaRemovido && respuesta {
			if detalle, err = json.Marshal(detalleRemovido(f.Inhabilidades.Removido)); err != nil {
				return fmt.Errorf("codificando el detalle de la pregunta %d: %w", pregunta, err)
			}
		}
		if _, err := ptx.Exec(ctx, `
INSERT INTO postulacion_inhabilidades (postulacion_id, pregunta, respuesta, detalle)
VALUES ($1, $2, $3, $4::jsonb)`, postulacionID, pregunta, respuesta, detalle); err != nil {
			return fmt.Errorf("guardando la pregunta %d de inhabilidades: %w", pregunta, err)
		}
	}
	if _, err := ptx.Exec(ctx, `UPDATE postulaciones SET elegible = $2 WHERE id = $1`, postulacionID, f.Elegible); err != nil {
		return fmt.Errorf("guardando la elegibilidad: %w", err)
	}
	return nil
}
