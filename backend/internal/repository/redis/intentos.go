package redis

import (
	"context"
	"fmt"
	"time"

	goredis "github.com/redis/go-redis/v9"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/domain"
)

// Intentos guarda en Redis el contador de fallos y el bloqueo de un flujo de
// autenticación. Claves (docs/arquitectura.md):
//
//	auth:fail:{flujo}:{cedula}  contador de fallos, expira con la ventana
//	auth:lock:{flujo}:{cedula}  bloqueo activo, expira con la duración del bloqueo
type Intentos struct {
	client *goredis.Client
	flujo  string
}

var _ domain.Intentos = (*Intentos)(nil)

// NewIntentos crea el repositorio para un flujo ("registro", "login", ...).
func NewIntentos(client *goredis.Client, flujo string) *Intentos {
	return &Intentos{client: client, flujo: flujo}
}

func (r *Intentos) failKey(cedula string) string { return "auth:fail:" + r.flujo + ":" + cedula }
func (r *Intentos) lockKey(cedula string) string { return "auth:lock:" + r.flujo + ":" + cedula }

// Bloqueo devuelve el TTL restante del bloqueo, o 0 si no existe.
func (r *Intentos) Bloqueo(ctx context.Context, cedula string) (time.Duration, error) {
	ttl, err := r.client.PTTL(ctx, r.lockKey(cedula)).Result()
	if err != nil {
		return 0, fmt.Errorf("consultando el bloqueo de %s: %w", r.flujo, err)
	}
	// PTTL devuelve -2 si la clave no existe y -1 si no tiene TTL. Un bloqueo sin TTL
	// no debería existir; se trata como inexistente para no bloquear para siempre.
	if ttl <= 0 {
		return 0, nil
	}
	return ttl, nil
}

// RegistrarFallo incrementa el contador y le pone el TTL solo en el primer fallo,
// para que la ventana no se alargue con cada intento.
func (r *Intentos) RegistrarFallo(ctx context.Context, cedula string, ventana time.Duration) (int, error) {
	key := r.failKey(cedula)
	var incr *goredis.IntCmd
	_, err := r.client.TxPipelined(ctx, func(p goredis.Pipeliner) error {
		incr = p.Incr(ctx, key)
		p.ExpireNX(ctx, key, ventana)
		return nil
	})
	if err != nil {
		return 0, fmt.Errorf("registrando el fallo de %s: %w", r.flujo, err)
	}
	return int(incr.Val()), nil
}

// Bloquear crea el bloqueo y borra el contador en la misma transacción: al terminar
// el bloqueo, el postulante vuelve a tener todos sus intentos.
func (r *Intentos) Bloquear(ctx context.Context, cedula string, duracion time.Duration) error {
	_, err := r.client.TxPipelined(ctx, func(p goredis.Pipeliner) error {
		p.Set(ctx, r.lockKey(cedula), "1", duracion)
		p.Del(ctx, r.failKey(cedula))
		return nil
	})
	if err != nil {
		return fmt.Errorf("bloqueando %s: %w", r.flujo, err)
	}
	return nil
}

// Reiniciar borra el contador de fallos.
func (r *Intentos) Reiniciar(ctx context.Context, cedula string) error {
	if err := r.client.Del(ctx, r.failKey(cedula)).Err(); err != nil {
		return fmt.Errorf("reiniciando los intentos de %s: %w", r.flujo, err)
	}
	return nil
}

// Desbloquear borra el bloqueo y el contador.
func (r *Intentos) Desbloquear(ctx context.Context, cedula string) error {
	if err := r.client.Del(ctx, r.lockKey(cedula), r.failKey(cedula)).Err(); err != nil {
		return fmt.Errorf("desbloqueando %s: %w", r.flujo, err)
	}
	return nil
}
