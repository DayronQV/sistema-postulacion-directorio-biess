package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/domain"
)

type PublicacionesRepo struct{ db *pgxpool.Pool }

var _ domain.SeccionPostulacion = (*PublicacionesRepo)(nil)

func NewPublicacionesRepo(db *pgxpool.Pool) *PublicacionesRepo { return &PublicacionesRepo{db: db} }

func (r *PublicacionesRepo) Cargar(ctx context.Context, postulacionID string, f *domain.Formulario) error {
	rows, err := r.db.Query(ctx, `
SELECT titulo, tipo, medio, to_char(fecha, 'YYYY-MM-DD'), relacionada, COALESCE(documento_id::text, '')
FROM publicaciones WHERE postulacion_id = $1 ORDER BY id`, postulacionID)
	if err != nil {
		return fmt.Errorf("consultando las publicaciones: %w", err)
	}
	defer rows.Close()
	f.Publicaciones = nil
	for rows.Next() {
		var fila domain.PublicacionAcademica
		if err := rows.Scan(&fila.Titulo, &fila.Tipo, &fila.Medio, &fila.Fecha, &fila.Relacionada, &fila.DocumentoID); err != nil {
			return fmt.Errorf("leyendo una publicación: %w", err)
		}
		f.Publicaciones = append(f.Publicaciones, fila)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("leyendo las publicaciones: %w", err)
	}
	return nil
}

func (r *PublicacionesRepo) Reemplazar(ctx context.Context, tx domain.Tx, postulacionID string, f domain.Formulario) error {
	ptx, err := pgxTx(tx)
	if err != nil {
		return err
	}
	for i, fila := range f.Publicaciones {
		var propio bool
		if err := ptx.QueryRow(ctx, `SELECT EXISTS (
  SELECT 1 FROM documentos WHERE id::text = $1 AND postulacion_id = $2 AND tipo = 'publicacion')`,
			fila.DocumentoID, postulacionID).Scan(&propio); err != nil {
			return fmt.Errorf("comprobando el documento de publicación: %w", err)
		}
		if !propio {
			return domain.Validation("Datos inválidos", map[string]string{fmt.Sprintf("publicaciones[%d].documento_id", i): "Referencia no válida"})
		}
	}
	if _, err := ptx.Exec(ctx, `DELETE FROM publicaciones WHERE postulacion_id = $1`, postulacionID); err != nil {
		return fmt.Errorf("borrando las publicaciones: %w", err)
	}
	if !f.TienePublicaciones {
		return nil
	}
	ahora := time.Now()
	for i, fila := range f.Publicaciones {
		id, err := idOrdenado(ahora, i)
		if err != nil {
			return err
		}
		if _, err := ptx.Exec(ctx, `
INSERT INTO publicaciones (id, postulacion_id, titulo, tipo, medio, fecha, relacionada, documento_id)
VALUES ($1, $2, $3, $4, $5, $6::date, $7, $8::uuid)`,
			id, postulacionID, fila.Titulo, fila.Tipo, fila.Medio, fila.Fecha, fila.Relacionada, fila.DocumentoID); err != nil {
			return fmt.Errorf("guardando una publicación: %w", err)
		}
	}
	return nil
}
