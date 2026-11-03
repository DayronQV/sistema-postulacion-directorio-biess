package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/domain"
)

type DocumentoRepo struct {
	db *pgxpool.Pool
}

var _ domain.Documentos = (*DocumentoRepo)(nil)
var _ domain.DocumentosAdmin = (*DocumentoRepo)(nil)

func NewDocumentoRepo(db *pgxpool.Pool) *DocumentoRepo {
	return &DocumentoRepo{db: db}
}

// Guardar bloquea la postulación mientras calcula el total para que dos subidas
// simultáneas no puedan superar juntas el máximo de 50 MB.
func (r *DocumentoRepo) Guardar(ctx context.Context, postulanteID string, d domain.Documento, limiteTotal int64) (domain.Documento, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return domain.Documento{}, fmt.Errorf("iniciando la subida del documento: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var postulacionID string
	if err = tx.QueryRow(ctx, `SELECT id FROM postulaciones WHERE postulante_id = $1 FOR UPDATE`, postulanteID).Scan(&postulacionID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Documento{}, domain.NotFound("Postulación no encontrada")
		}
		return domain.Documento{}, fmt.Errorf("buscando la postulación del documento: %w", err)
	}
	var total int64
	if err = tx.QueryRow(ctx, `SELECT COALESCE(SUM(tamano_bytes), 0) FROM documentos WHERE postulacion_id = $1`, postulacionID).Scan(&total); err != nil {
		return domain.Documento{}, fmt.Errorf("calculando el tamaño de los documentos: %w", err)
	}
	if total+d.Tamano > limiteTotal {
		return domain.Documento{}, domain.Unprocessable("Los documentos de la postulación superan el máximo total de 50 MB", map[string]string{"archivo": "Se superaría el límite total de 50 MB"})
	}
	if err = tx.QueryRow(ctx, `
INSERT INTO documentos (postulacion_id, tipo, nombre, tamano_bytes, contenido)
VALUES ($1, $2, $3, $4, $5)
RETURNING id::text`, postulacionID, d.Tipo, d.Nombre, d.Tamano, d.Contenido).Scan(&d.ID); err != nil {
		return domain.Documento{}, fmt.Errorf("guardando el documento: %w", err)
	}
	if err = tx.Commit(ctx); err != nil {
		return domain.Documento{}, fmt.Errorf("confirmando la subida del documento: %w", err)
	}
	return d, nil
}

func (r *DocumentoRepo) ObtenerPropio(ctx context.Context, postulanteID, documentoID string) (domain.Documento, error) {
	var d domain.Documento
	err := r.db.QueryRow(ctx, `
SELECT d.id::text, d.tipo, d.nombre, d.tamano_bytes, d.contenido
FROM documentos d
JOIN postulaciones p ON p.id = d.postulacion_id
WHERE p.postulante_id = $1 AND d.id::text = $2`, postulanteID, documentoID).Scan(
		&d.ID, &d.Tipo, &d.Nombre, &d.Tamano, &d.Contenido,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Documento{}, domain.NotFound("Documento no encontrado")
	}
	if err != nil {
		return domain.Documento{}, fmt.Errorf("descargando el documento: %w", err)
	}
	return d, nil
}

// ObtenerDePostulacion descarga un documento solo si pertenece al expediente indicado.
// Un id ajeno responde igual que uno inexistente para no exponer otros expedientes.
func (r *DocumentoRepo) ObtenerDePostulacion(ctx context.Context, postulacionID, documentoID string) (domain.Documento, error) {
	var d domain.Documento
	err := r.db.QueryRow(ctx, `
SELECT id::text, tipo, nombre, tamano_bytes, contenido
FROM documentos
WHERE postulacion_id = $1 AND id::text = $2`, postulacionID, documentoID).Scan(
		&d.ID, &d.Tipo, &d.Nombre, &d.Tamano, &d.Contenido,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Documento{}, domain.NotFound("Documento no encontrado")
	}
	if err != nil {
		return domain.Documento{}, fmt.Errorf("descargando el documento de la postulación: %w", err)
	}
	return d, nil
}

func (r *DocumentoRepo) EliminarPropioNoReferenciado(ctx context.Context, postulanteID, documentoID string) error {
	var eliminado string
	err := r.db.QueryRow(ctx, `
DELETE FROM documentos d
USING postulaciones p
WHERE d.postulacion_id = p.id
  AND p.postulante_id = $1
  AND d.id::text = $2
  AND NOT EXISTS (SELECT 1 FROM postulaciones x WHERE x.respaldo_organizacion_doc_id = d.id)
  AND NOT EXISTS (SELECT 1 FROM formacion_academica x WHERE x.documento_id = d.id)
  AND NOT EXISTS (SELECT 1 FROM experiencia_profesional x WHERE x.documento_id = d.id)
  AND NOT EXISTS (SELECT 1 FROM publicaciones x WHERE x.documento_id = d.id)
  AND NOT EXISTS (SELECT 1 FROM capacitaciones x WHERE x.documento_id = d.id)
RETURNING d.id::text`, postulanteID, documentoID).Scan(&eliminado)
	if err == nil {
		return nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("eliminando el documento: %w", err)
	}
	var propio bool
	if err = r.db.QueryRow(ctx, `
SELECT EXISTS (
  SELECT 1 FROM documentos d JOIN postulaciones p ON p.id = d.postulacion_id
  WHERE p.postulante_id = $1 AND d.id::text = $2)`, postulanteID, documentoID).Scan(&propio); err != nil {
		return fmt.Errorf("comprobando el documento: %w", err)
	}
	if propio {
		return domain.Conflict("El documento está siendo usado por la postulación", nil)
	}
	return domain.NotFound("Documento no encontrado")
}
