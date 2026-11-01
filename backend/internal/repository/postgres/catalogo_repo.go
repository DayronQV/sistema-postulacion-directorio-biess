package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/domain"
)

// paisPrioritario es el país que encabeza la lista de países.
const paisPrioritario = "Ecuador"

// CatalogoRepo implementa domain.Catalogos sobre las tablas paises, provincias y
// cantones.
type CatalogoRepo struct {
	db *pgxpool.Pool
}

var _ domain.Catalogos = (*CatalogoRepo)(nil)

// NewCatalogoRepo crea el repositorio de catálogos.
func NewCatalogoRepo(db *pgxpool.Pool) *CatalogoRepo {
	return &CatalogoRepo{db: db}
}

// Paises lista los países con Ecuador primero y el resto por nombre.
func (r *CatalogoRepo) Paises(ctx context.Context) ([]domain.ItemCatalogo, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, nombre FROM paises ORDER BY (nombre <> $1), nombre`, paisPrioritario)
	if err != nil {
		return nil, fmt.Errorf("consultando los países: %w", err)
	}
	defer rows.Close()

	items := []domain.ItemCatalogo{}
	for rows.Next() {
		var it domain.ItemCatalogo
		if err := rows.Scan(&it.ID, &it.Nombre); err != nil {
			return nil, fmt.Errorf("leyendo un país: %w", err)
		}
		items = append(items, it)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("leyendo los países: %w", err)
	}
	return items, nil
}

// Provincias lista las provincias del país por nombre. Devuelve domain.NotFound si el
// país no existe.
func (r *CatalogoRepo) Provincias(ctx context.Context, paisID int) ([]domain.ItemCatalogo, error) {
	return r.hijos(ctx, "provincias del país",
		`SELECT h.id, h.nombre
		   FROM paises p
		   LEFT JOIN provincias h ON h.pais_id = p.id
		  WHERE p.id = $1
		  ORDER BY h.nombre`,
		paisID, "El país no existe")
}

// Cantones lista los cantones de la provincia por nombre. Devuelve domain.NotFound si
// la provincia no existe.
func (r *CatalogoRepo) Cantones(ctx context.Context, provinciaID int) ([]domain.ItemCatalogo, error) {
	return r.hijos(ctx, "cantones de la provincia",
		`SELECT h.id, h.nombre
		   FROM provincias p
		   LEFT JOIN cantones h ON h.provincia_id = p.id
		  WHERE p.id = $1
		  ORDER BY h.nombre`,
		provinciaID, "La provincia no existe")
}

// hijos ejecuta una consulta de la forma "padre LEFT JOIN hijos": devuelve una fila
// por hijo, o una sola con hijo NULL si el padre existe sin hijos, o ninguna si el
// padre no existe. Así una sola consulta distingue lista vacía de padre inexistente.
func (r *CatalogoRepo) hijos(ctx context.Context, que, query string, padreID int, msgNoExiste string) ([]domain.ItemCatalogo, error) {
	rows, err := r.db.Query(ctx, query, padreID)
	if err != nil {
		return nil, fmt.Errorf("consultando las %s: %w", que, err)
	}
	defer rows.Close()

	existe := false
	items := []domain.ItemCatalogo{}
	for rows.Next() {
		existe = true
		var (
			id     *int
			nombre *string
		)
		if err := rows.Scan(&id, &nombre); err != nil {
			return nil, fmt.Errorf("leyendo las %s: %w", que, err)
		}
		if id != nil && nombre != nil {
			items = append(items, domain.ItemCatalogo{ID: *id, Nombre: *nombre})
		}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("leyendo las %s: %w", que, err)
	}
	if !existe {
		return nil, domain.NotFound(msgNoExiste)
	}
	return items, nil
}
