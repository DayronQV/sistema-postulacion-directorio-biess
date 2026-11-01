package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/domain"
)

type ReferenciasRepo struct {
	db *pgxpool.Pool
}

var _ domain.SeccionPostulacion = (*ReferenciasRepo)(nil)

func NewReferenciasRepo(db *pgxpool.Pool) *ReferenciasRepo {
	return &ReferenciasRepo{db: db}
}

func (r *ReferenciasRepo) Cargar(ctx context.Context, postulacionID string, f *domain.Formulario) error {
	rows, err := r.db.Query(ctx, `
SELECT nombres_apellidos, relacion, telefono, correo
FROM referencias_personales
WHERE postulacion_id = $1
ORDER BY nombres_apellidos, id`, postulacionID)
	if err != nil {
		return fmt.Errorf("consultando las referencias personales: %w", err)
	}
	defer rows.Close()
	f.Referencias = nil
	for rows.Next() {
		var referencia domain.ReferenciaPersonal
		if err := rows.Scan(&referencia.NombresApellidos, &referencia.Relacion, &referencia.Telefono, &referencia.Correo); err != nil {
			return fmt.Errorf("leyendo una referencia personal: %w", err)
		}
		f.Referencias = append(f.Referencias, referencia)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("leyendo las referencias personales: %w", err)
	}
	return nil
}

func (r *ReferenciasRepo) Reemplazar(ctx context.Context, tx domain.Tx, postulacionID string, f domain.Formulario) error {
	ptx, err := pgxTx(tx)
	if err != nil {
		return err
	}
	if _, err := ptx.Exec(ctx, `DELETE FROM referencias_personales WHERE postulacion_id = $1`, postulacionID); err != nil {
		return fmt.Errorf("borrando las referencias personales: %w", err)
	}
	for _, referencia := range f.Referencias {
		if _, err := ptx.Exec(ctx, `
INSERT INTO referencias_personales (postulacion_id, nombres_apellidos, relacion, telefono, correo)
VALUES ($1, $2, $3, $4, $5)`, postulacionID, referencia.NombresApellidos,
			referencia.Relacion, referencia.Telefono, referencia.Correo); err != nil {
			return fmt.Errorf("guardando una referencia personal: %w", err)
		}
	}
	return nil
}
