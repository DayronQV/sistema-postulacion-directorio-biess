package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	goredis "github.com/redis/go-redis/v9"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/domain"
)

// prefijoCatalogo antecede la clave lógica del servicio: cache:catalogo:paises,
// cache:catalogo:provincias:{paisId}, cache:catalogo:cantones:{provinciaId}.
const prefijoCatalogo = "cache:catalogo:"

// cacheTimeout acota cada operación de la caché. Si Redis no responde, la lectura
// falla pronto y el servicio sigue con PostgreSQL en lugar de agotar el plazo de la
// petición esperando a Redis.
const cacheTimeout = 500 * time.Millisecond

// CacheCatalogos guarda las listas de catálogos como JSON, con TTL.
type CacheCatalogos struct {
	client *goredis.Client
}

var _ domain.CacheCatalogos = (*CacheCatalogos)(nil)

// NewCacheCatalogos crea la caché de catálogos.
func NewCacheCatalogos(client *goredis.Client) *CacheCatalogos {
	return &CacheCatalogos{client: client}
}

// itemJSON es la forma guardada en Redis; el dominio no lleva etiquetas JSON.
type itemJSON struct {
	ID     int    `json:"id"`
	Nombre string `json:"nombre"`
}

// Obtener devuelve la lista guardada. Una clave ausente es un acierto negativo
// (false, sin error). Un valor que no se puede leer se trata también como ausente:
// el servicio lo reemplaza con el dato de la base al guardar.
func (c *CacheCatalogos) Obtener(ctx context.Context, clave string) ([]domain.ItemCatalogo, bool, error) {
	ctx, cancel := context.WithTimeout(ctx, cacheTimeout)
	defer cancel()

	raw, err := c.client.Get(ctx, prefijoCatalogo+clave).Bytes()
	if errors.Is(err, goredis.Nil) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, fmt.Errorf("leyendo %s de la caché: %w", clave, err)
	}

	var stored []itemJSON
	if err := json.Unmarshal(raw, &stored); err != nil {
		return nil, false, nil
	}
	items := make([]domain.ItemCatalogo, len(stored))
	for i, it := range stored {
		items[i] = domain.ItemCatalogo{ID: it.ID, Nombre: it.Nombre}
	}
	return items, true, nil
}

// Guardar guarda la lista con el TTL indicado.
func (c *CacheCatalogos) Guardar(ctx context.Context, clave string, items []domain.ItemCatalogo, ttl time.Duration) error {
	ctx, cancel := context.WithTimeout(ctx, cacheTimeout)
	defer cancel()

	stored := make([]itemJSON, len(items))
	for i, it := range items {
		stored[i] = itemJSON{ID: it.ID, Nombre: it.Nombre}
	}
	raw, err := json.Marshal(stored)
	if err != nil {
		return fmt.Errorf("serializando %s para la caché: %w", clave, err)
	}
	if err := c.client.Set(ctx, prefijoCatalogo+clave, raw, ttl).Err(); err != nil {
		return fmt.Errorf("guardando %s en la caché: %w", clave, err)
	}
	return nil
}
