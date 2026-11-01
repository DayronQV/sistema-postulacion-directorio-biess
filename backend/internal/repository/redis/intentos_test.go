package redis_test

import (
	"context"
	"os"
	"testing"
	"time"

	goredis "github.com/redis/go-redis/v9"

	redisrepo "github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/repository/redis"
)

// newClient conecta al Redis de docker compose. Si no está levantado, la prueba se
// salta: las reglas del bloqueo ya se prueban en el servicio con repositorios falsos.
func newClient(t *testing.T) *goredis.Client {
	t.Helper()
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		addr = "localhost:6379"
	}
	client, err := redisrepo.NewClient(context.Background(), addr)
	if err != nil {
		t.Skipf("Redis no disponible: %v", err)
	}
	t.Cleanup(func() { _ = client.Close() })
	return client
}

func TestIntentos_ContadorBloqueoYTTL(t *testing.T) {
	ctx := context.Background()
	client := newClient(t)
	const cedula = "0000000000"
	flujo := "test-" + time.Now().Format("150405.000000")
	t.Cleanup(func() {
		client.Del(ctx, "auth:fail:"+flujo+":"+cedula, "auth:lock:"+flujo+":"+cedula)
	})
	r := redisrepo.NewIntentos(client, flujo)

	for want := 1; want <= 2; want++ {
		n, err := r.RegistrarFallo(ctx, cedula, time.Minute)
		if err != nil || n != want {
			t.Fatalf("RegistrarFallo = %d, %v; se esperaba %d", n, err, want)
		}
	}
	if ttl := client.TTL(ctx, "auth:fail:"+flujo+":"+cedula).Val(); ttl <= 0 || ttl > time.Minute {
		t.Fatalf("el contador no tiene TTL: %v", ttl)
	}

	if d, _ := r.Bloqueo(ctx, cedula); d != 0 {
		t.Fatalf("no debería estar bloqueada: %v", d)
	}
	if err := r.Bloquear(ctx, cedula, 10*time.Minute); err != nil {
		t.Fatal(err)
	}
	d, err := r.Bloqueo(ctx, cedula)
	if err != nil || d <= 9*time.Minute || d > 10*time.Minute {
		t.Fatalf("Bloqueo = %v, %v", d, err)
	}
	if client.Exists(ctx, "auth:fail:"+flujo+":"+cedula).Val() != 0 {
		t.Fatalf("Bloquear debe borrar el contador")
	}

	if _, err := r.RegistrarFallo(ctx, cedula, time.Minute); err != nil {
		t.Fatal(err)
	}
	if err := r.Reiniciar(ctx, cedula); err != nil {
		t.Fatal(err)
	}
	if client.Exists(ctx, "auth:fail:"+flujo+":"+cedula).Val() != 0 {
		t.Fatalf("Reiniciar debe borrar el contador")
	}
}

func TestTicketsRegistro_Emitir(t *testing.T) {
	ctx := context.Background()
	client := newClient(t)
	id, err := redisrepo.NewTicketsRegistro(client).Emitir(ctx, "1712345675", time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { client.Del(ctx, "registro:ticket:"+id) })
	if len(id) != 32 {
		t.Fatalf("id = %q, se esperaban 32 caracteres hex", id)
	}
	if v := client.Get(ctx, "registro:ticket:"+id).Val(); v != "1712345675" {
		t.Fatalf("valor = %q", v)
	}
	if ttl := client.TTL(ctx, "registro:ticket:"+id).Val(); ttl <= 0 {
		t.Fatalf("el ticket no tiene TTL: %v", ttl)
	}
}
