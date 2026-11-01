package domain

import (
	"context"
	"time"
)

// ItemCatalogo es un elemento de los catálogos geográficos (país, provincia o cantón).
type ItemCatalogo struct {
	ID     int
	Nombre string
}

// Catalogos es el puerto de lectura de la división política. La implementación vive
// en internal/repository/postgres.
//
// Provincias y Cantones devuelven domain.NotFound si el padre (país o provincia) no
// existe; si existe pero no tiene hijos devuelven una lista vacía.
type Catalogos interface {
	// Paises lista todos los países, con Ecuador primero y el resto por nombre.
	Paises(ctx context.Context) ([]ItemCatalogo, error)
	// Provincias lista las provincias del país, por nombre.
	Provincias(ctx context.Context, paisID int) ([]ItemCatalogo, error)
	// Cantones lista los cantones de la provincia, por nombre.
	Cantones(ctx context.Context, provinciaID int) ([]ItemCatalogo, error)
}

// CacheCatalogos es el puerto de la caché de catálogos. La clave es lógica
// ("paises", "provincias:1", ...): el prefijo del almacén lo pone la implementación,
// que vive en internal/repository/redis.
type CacheCatalogos interface {
	// Obtener devuelve la lista guardada y true, o false si la clave no está en la
	// caché. Un error significa que la caché no pudo consultarse.
	Obtener(ctx context.Context, clave string) ([]ItemCatalogo, bool, error)
	// Guardar guarda la lista durante ttl.
	Guardar(ctx context.Context, clave string, items []ItemCatalogo, ttl time.Duration) error
}
