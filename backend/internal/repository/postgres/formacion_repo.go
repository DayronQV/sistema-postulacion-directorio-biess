package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/domain"
)

type FormacionRepo struct {
	db *pgxpool.Pool
}

var _ domain.SeccionPostulacion = (*FormacionRepo)(nil)

func NewFormacionRepo(db *pgxpool.Pool) *FormacionRepo {
	return &FormacionRepo{db: db}
}

func (r *FormacionRepo) Cargar(ctx context.Context, postulacionID string, f *domain.Formulario) error {
	rows, err := r.db.Query(ctx, `
SELECT nivel, institucion, tipo, titulo, area, senescyt_numero,
       to_char(senescyt_fecha, 'YYYY-MM-DD'), documento_id::text
FROM formacion_academica
WHERE postulacion_id = $1
ORDER BY id`, postulacionID)
	if err != nil {
		return fmt.Errorf("consultando la formación académica: %w", err)
	}
	defer rows.Close()
	f.Formacion = nil
	for rows.Next() {
		var fila domain.FormacionAcademica
		if err := rows.Scan(&fila.Nivel, &fila.Institucion, &fila.Tipo, &fila.Titulo, &fila.Area,
			&fila.SenescytNumero, &fila.SenescytFecha, &fila.DocumentoID); err != nil {
			return fmt.Errorf("leyendo una formación académica: %w", err)
		}
		f.Formacion = append(f.Formacion, fila)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("leyendo la formación académica: %w", err)
	}
	return nil
}

func (r *FormacionRepo) Reemplazar(ctx context.Context, tx domain.Tx, postulacionID string, f domain.Formulario) error {
	ptx, err := pgxTx(tx)
	if err != nil {
		return err
	}
	for i, fila := range f.Formacion {
		var propio bool
		if err := ptx.QueryRow(ctx, `
SELECT EXISTS (
  SELECT 1 FROM documentos
  WHERE id::text = $1 AND postulacion_id = $2 AND tipo = 'formacion')`,
			fila.DocumentoID, postulacionID).Scan(&propio); err != nil {
			return fmt.Errorf("comprobando el documento de formación: %w", err)
		}
		if !propio {
			campo := fmt.Sprintf("formacion[%d].documento_id", i)
			return domain.Validation("Datos inválidos", map[string]string{campo: "Referencia no válida"})
		}
	}
	if _, err := ptx.Exec(ctx, `DELETE FROM formacion_academica WHERE postulacion_id = $1`, postulacionID); err != nil {
		return fmt.Errorf("borrando la formación académica: %w", err)
	}
	ahora := time.Now()
	for i, fila := range f.Formacion {
		id, err := idOrdenado(ahora, i)
		if err != nil {
			return err
		}
		if _, err := ptx.Exec(ctx, `
INSERT INTO formacion_academica
  (id, postulacion_id, nivel, institucion, tipo, titulo, area, senescyt_numero, senescyt_fecha, documento_id)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9::date, $10::uuid)`,
			id, postulacionID, fila.Nivel, fila.Institucion, fila.Tipo, fila.Titulo, fila.Area,
			fila.SenescytNumero, fila.SenescytFecha, fila.DocumentoID); err != nil {
			return fmt.Errorf("guardando la formación académica: %w", err)
		}
	}
	return nil
}
