package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/domain"
)

// FormacionRepo implementa domain.SeccionPostulacion para la tabla formacion_academica.
type FormacionRepo struct {
	db *pgxpool.Pool
}

var _ domain.SeccionPostulacion = (*FormacionRepo)(nil)

// NewFormacionRepo crea el repositorio de formación académica.
func NewFormacionRepo(db *pgxpool.Pool) *FormacionRepo {
	return &FormacionRepo{db: db}
}

// Cargar completa f.Formacion con las filas de la postulación en el orden de captura
// (ver idOrdenado). Una fila sin respaldo devuelve DocumentoID vacío.
func (r *FormacionRepo) Cargar(ctx context.Context, postulacionID string, f *domain.Formulario) error {
	rows, err := r.db.Query(ctx, `
SELECT nivel, institucion, tipo, titulo, area, senescyt_numero,
       to_char(senescyt_fecha, 'YYYY-MM-DD'), COALESCE(documento_id::text, '')
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

// Reemplazar borra la formación de la postulación y escribe la de f, dentro de tx. Antes
// comprueba que cada respaldo indicado sea de esta postulación y de tipo formacion.
func (r *FormacionRepo) Reemplazar(ctx context.Context, tx domain.Tx, postulacionID string, f domain.Formulario) error {
	ptx, err := pgxTx(tx)
	if err != nil {
		return err
	}
	for i, fila := range f.Formacion {
		if fila.DocumentoID == "" {
			continue
		}
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
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9::date, NULLIF($10, '')::uuid)`,
			id, postulacionID, fila.Nivel, fila.Institucion, fila.Tipo, fila.Titulo, fila.Area,
			fila.SenescytNumero, fila.SenescytFecha, fila.DocumentoID); err != nil {
			return fmt.Errorf("guardando la formación académica: %w", err)
		}
	}
	return nil
}
