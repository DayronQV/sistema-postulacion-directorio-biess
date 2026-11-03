package redis_test

import (
	"context"
	"os"
	"reflect"
	"sync"
	"testing"
	"time"

	goredis "github.com/redis/go-redis/v9"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/domain"
	redisrepo "github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/repository/redis"
)

var (
	redisCheck    sync.Once
	redisCheckErr error
)

// newCacheClient es newClient con la comprobación de disponibilidad hecha una sola
// vez: sin servidor, solo la primera prueba espera el tiempo de conexión y el resto se
// salta al instante.
func newCacheClient(t *testing.T) *goredis.Client {
	t.Helper()
	redisCheck.Do(func() {
		addr := os.Getenv("REDIS_ADDR")
		if addr == "" {
			addr = "localhost:6379"
		}
		client, err := redisrepo.NewClient(context.Background(), addr)
		if err == nil {
			_ = client.Close()
		}
		redisCheckErr = err
	})
	if redisCheckErr != nil {
		t.Skipf("Redis no disponible: %v", redisCheckErr)
	}
	return newClient(t)
}

func TestCacheCatalogos_GuardarObtenerYTTL(t *testing.T) {
	ctx := context.Background()
	client := newCacheClient(t)
	clave := "test-" + time.Now().Format("150405.000000")
	t.Cleanup(func() { client.Del(ctx, "cache:catalogo:"+clave) })
	c := redisrepo.NewCacheCatalogos(client)

	if _, hit, err := c.Obtener(ctx, clave); err != nil || hit {
		t.Fatalf("Obtener de una clave ausente = hit %v, err %v; se esperaba un fallo limpio", hit, err)
	}

	items := []domain.ItemCatalogo{{ID: 17, Nombre: "Pichincha"}, {ID: 3, Nombre: "Ñuñoa"}}
	if err := c.Guardar(ctx, clave, items, time.Hour); err != nil {
		t.Fatal(err)
	}
	got, hit, err := c.Obtener(ctx, clave)
	if err != nil || !hit || !reflect.DeepEqual(got, items) {
		t.Fatalf("Obtener = %v, hit %v, err %v; se esperaba %v", got, hit, err, items)
	}
	if ttl := client.TTL(ctx, "cache:catalogo:"+clave).Val(); ttl <= 59*time.Minute || ttl > time.Hour {
		t.Fatalf("TTL = %v; se esperaba cerca de 1h", ttl)
	}
}

func TestCacheCatalogos_ListaVaciaSeGuardaComoArreglo(t *testing.T) {
	ctx := context.Background()
	client := newCacheClient(t)
	clave := "test-vacia-" + time.Now().Format("150405.000000")
	t.Cleanup(func() { client.Del(ctx, "cache:catalogo:"+clave) })
	c := redisrepo.NewCacheCatalogos(client)

	if err := c.Guardar(ctx, clave, []domain.ItemCatalogo{}, time.Minute); err != nil {
		t.Fatal(err)
	}
	if raw := client.Get(ctx, "cache:catalogo:"+clave).Val(); raw != "[]" {
		t.Fatalf("valor guardado = %q; se esperaba []", raw)
	}
	got, hit, err := c.Obtener(ctx, clave)
	if err != nil || !hit || len(got) != 0 {
		t.Fatalf("Obtener = %v, hit %v, err %v; se esperaba un acierto con lista vacía", got, hit, err)
	}
}

func TestCacheCatalogos_ValorIlegibleSeTrataComoAusente(t *testing.T) {
	ctx := context.Background()
	client := newCacheClient(t)
	clave := "test-corrupta-" + time.Now().Format("150405.000000")
	t.Cleanup(func() { client.Del(ctx, "cache:catalogo:"+clave) })
	if err := client.Set(ctx, "cache:catalogo:"+clave, "{no es json", time.Minute).Err(); err != nil {
		t.Fatal(err)
	}

	_, hit, err := redisrepo.NewCacheCatalogos(client).Obtener(ctx, clave)
	if err != nil || hit {
		t.Fatalf("Obtener = hit %v, err %v; se esperaba tratarlo como ausente", hit, err)
	}
}

func TestCacheCatalogos_RedisCaido_DevuelveError(t *testing.T) {
	client := newCacheClient(t)
	_ = client.Close() // el cliente cerrado hace que cada comando falle, como un Redis caído
	c := redisrepo.NewCacheCatalogos(client)

	if _, _, err := c.Obtener(context.Background(), "paises"); err == nil {
		t.Fatal("Obtener con Redis caído debe devolver error")
	}
	if err := c.Guardar(context.Background(), "paises", nil, time.Hour); err == nil {
		t.Fatal("Guardar con Redis caído debe devolver error")
	}
}
