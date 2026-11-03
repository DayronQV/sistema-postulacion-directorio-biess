package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/domain"
)

var _ domain.PostulacionesAdmin = (*PostulacionRepo)(nil)
var _ domain.MatrizPostulaciones = (*PostulacionRepo)(nil)

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

// ListarMatriz reúne en una sola consulta los indicadores del cuadro comparativo.
func (r *PostulacionRepo) ListarMatriz(ctx context.Context, f domain.FiltroMatrizAdmin) ([]domain.FilaMatrizAdmin, error) {
	rows, err := r.db.Query(ctx, `
SELECT p.codigo, u.cedula, p.nombres, p.apellidos, COALESCE(p.cargo, ''),
       p.estado, p.elegible,
       COALESCE((
           SELECT fa.nivel
           FROM formacion_academica fa
           WHERE fa.postulacion_id = p.id
           ORDER BY CASE fa.nivel
               WHEN 'phd' THEN 4
               WHEN 'maestria' THEN 3
               WHEN 'especialidad' THEN 2
               WHEN 'tercer_nivel' THEN 1
               ELSE 0
           END DESC, fa.id
           LIMIT 1
       ), ''),
       COALESCE((
           SELECT SUM(
               EXTRACT(YEAR FROM age(ep.fecha_fin, ep.fecha_inicio)) * 360 +
               EXTRACT(MONTH FROM age(ep.fecha_fin, ep.fecha_inicio)) * 30 +
               EXTRACT(DAY FROM age(ep.fecha_fin, ep.fecha_inicio))
           )::double precision / 360.0
           FROM experiencia_profesional ep
           WHERE ep.postulacion_id = p.id
       ), 0),
       (SELECT COUNT(*) FROM capacitaciones c WHERE c.postulacion_id = p.id),
       (SELECT COUNT(*) FROM publicaciones pub WHERE pub.postulacion_id = p.id)
FROM postulaciones p
JOIN postulantes u ON u.id = p.postulante_id
WHERE ($1 = '' OR p.cargo = $1)
  AND ($2 = '' OR p.estado = $2)
  AND ($3::boolean IS NULL OR p.elegible = $3)
ORDER BY p.created_at DESC, p.id DESC`, f.Cargo, f.Estado, f.Elegible)
	if err != nil {
		return nil, fmt.Errorf("consultando la matriz de postulantes: %w", err)
	}
	defer rows.Close()

	items := make([]domain.FilaMatrizAdmin, 0)
	for rows.Next() {
		var item domain.FilaMatrizAdmin
		if err := rows.Scan(
			&item.Codigo, &item.Cedula, &item.Nombres, &item.Apellidos, &item.Cargo,
			&item.Estado, &item.Elegible, &item.NivelAcademico, &item.AniosExperiencia,
			&item.Capacitaciones, &item.Publicaciones,
		); err != nil {
			return nil, fmt.Errorf("leyendo la matriz de postulantes: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("leyendo la matriz de postulantes: %w", err)
	}
	return items, nil
}
