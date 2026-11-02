package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CapacitacionesRepo struct{ db *pgxpool.Pool }

var _ domain.SeccionPostulacion = (*CapacitacionesRepo)(nil)

func NewCapacitacionesRepo(db *pgxpool.Pool) *CapacitacionesRepo { return &CapacitacionesRepo{db: db} }

func (r *CapacitacionesRepo) Cargar(ctx context.Context, postulacionID string, f *domain.Formulario) error {
	rows, err := r.db.Query(ctx, `SELECT tipo, nombre, institucion, to_char(fecha_inicio, 'YYYY-MM-DD'), to_char(fecha_fin, 'YYYY-MM-DD'), horas, COALESCE(documento_id::text, '') FROM capacitaciones WHERE postulacion_id = $1 ORDER BY id`, postulacionID)
	if err != nil {
		return fmt.Errorf("consultando las capacitaciones: %w", err)
	}
	defer rows.Close()
	f.Capacitaciones = nil
	for rows.Next() {
		var fila domain.Capacitacion
		if err := rows.Scan(&fila.Tipo, &fila.Nombre, &fila.Institucion, &fila.FechaInicio, &fila.FechaFin, &fila.Horas, &fila.DocumentoID); err != nil {
			return fmt.Errorf("leyendo una capacitación: %w", err)
		}
		f.Capacitaciones = append(f.Capacitaciones, fila)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("leyendo las capacitaciones: %w", err)
	}
	return nil
}

func (r *CapacitacionesRepo) Reemplazar(ctx context.Context, tx domain.Tx, postulacionID string, f domain.Formulario) error {
	ptx, err := pgxTx(tx)
	if err != nil {
		return err
	}
	for i, fila := range f.Capacitaciones {
		if fila.DocumentoID == "" {
			continue
		}
		var propio bool
		if err := ptx.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM documentos WHERE id::text = $1 AND postulacion_id = $2 AND tipo = 'capacitacion')`, fila.DocumentoID, postulacionID).Scan(&propio); err != nil {
			return fmt.Errorf("comprobando el documento de capacitación: %w", err)
		}
		if !propio {
			return domain.Validation("Datos inválidos", map[string]string{fmt.Sprintf("capacitaciones[%d].documento_id", i): "Referencia no válida"})
		}
	}
	if _, err := ptx.Exec(ctx, `DELETE FROM capacitaciones WHERE postulacion_id = $1`, postulacionID); err != nil {
		return fmt.Errorf("borrando las capacitaciones: %w", err)
	}
	ahora := time.Now()
	for i, fila := range f.Capacitaciones {
		id, err := idOrdenado(ahora, i)
		if err != nil {
			return err
		}
		if _, err := ptx.Exec(ctx, `INSERT INTO capacitaciones (id, postulacion_id, tipo, nombre, institucion, fecha_inicio, fecha_fin, horas, documento_id) VALUES ($1, $2, $3, $4, $5, $6::date, $7::date, $8, NULLIF($9, '')::uuid)`, id, postulacionID, fila.Tipo, fila.Nombre, fila.Institucion, fila.FechaInicio, fila.FechaFin, fila.Horas, fila.DocumentoID); err != nil {
			return fmt.Errorf("guardando una capacitación: %w", err)
		}
	}
	return nil
}
