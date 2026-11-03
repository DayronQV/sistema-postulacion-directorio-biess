package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/domain"
)

var _ domain.PostulacionesAdmin = (*PostulacionRepo)(nil)

const filtroPostulacionesAdmin = `
FROM postulaciones p
JOIN postulantes u ON u.id = p.postulante_id
WHERE ($1 = '' OR p.cargo = $1)
  AND ($2 = '' OR p.estado = $2)
  AND ($3::boolean IS NULL OR p.elegible = $3)`

// ListarAdmin devuelve una página estable y el total que respeta los mismos filtros.
func (r *PostulacionRepo) ListarAdmin(ctx context.Context, f domain.FiltroPostulacionesAdmin) ([]domain.ResumenPostulacionAdmin, int, error) {
	var total int
	if err := r.db.QueryRow(ctx, `SELECT COUNT(*) `+filtroPostulacionesAdmin,
		f.Cargo, f.Estado, f.Elegible).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("contando la nómina de postulaciones: %w", err)
	}
	rows, err := r.db.Query(ctx, `
SELECT p.id::text, p.codigo, u.cedula, p.nombres, p.apellidos,
       COALESCE(p.cargo, ''), p.estado, p.elegible
`+filtroPostulacionesAdmin+`
ORDER BY p.created_at DESC, p.id DESC
LIMIT $4 OFFSET $5`, f.Cargo, f.Estado, f.Elegible, f.PorPagina, (f.Pagina-1)*f.PorPagina)
	if err != nil {
		return nil, 0, fmt.Errorf("consultando la nómina de postulaciones: %w", err)
	}
	defer rows.Close()
	items := make([]domain.ResumenPostulacionAdmin, 0)
	for rows.Next() {
		var item domain.ResumenPostulacionAdmin
		if err := rows.Scan(&item.ID, &item.Codigo, &item.Cedula, &item.Nombres, &item.Apellidos,
			&item.Cargo, &item.Estado, &item.Elegible); err != nil {
			return nil, 0, fmt.Errorf("leyendo la nómina de postulaciones: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("leyendo la nómina de postulaciones: %w", err)
	}
	return items, total, nil
}

// ObtenerPorIDAdmin obtiene el formulario base por su identificador; el servicio carga
// después las secciones con filas mediante sus puertos existentes.
func (r *PostulacionRepo) ObtenerPorIDAdmin(ctx context.Context, postulacionID string) (domain.Postulacion, error) {
	p, err := scanPostulacion(r.db.QueryRow(ctx, selectPostulacion+` WHERE id = $1`, postulacionID))
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Postulacion{}, domain.NotFound("La postulación no existe")
	}
	if err != nil {
		return domain.Postulacion{}, fmt.Errorf("consultando el detalle de la postulación: %w", err)
	}
	return p, nil
}
