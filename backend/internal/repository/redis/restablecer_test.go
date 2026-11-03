package redis_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/domain"
	redisrepo "github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/repository/redis"
)

func TestTokensRestablecer_UnSoloUso(t *testing.T) {
	ctx := context.Background()
	client := newClient(t)
	r := redisrepo.NewTokensRestablecer(client)

	token, err := r.Emitir(ctx, "1802468130", 30*time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { client.Del(ctx, "auth:reset:"+token) })
	if len(token) != 64 {
		t.Fatalf("token = %q, se esperaban 64 caracteres hex", token)
	}
	if ttl := client.TTL(ctx, "auth:reset:"+token).Val(); ttl <= 29*time.Minute || ttl > 30*time.Minute {
		t.Fatalf("TTL = %v, se esperaban 30 minutos", ttl)
	}

	cedula, err := r.Consumir(ctx, token)
	if err != nil || cedula != "1802468130" {
		t.Fatalf("Consumir = %q, %v", cedula, err)
	}
	if _, err := r.Consumir(ctx, token); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("el segundo uso debe dar ErrNotFound, llegó %v", err)
	}
}

func TestIntentos_Desbloquear(t *testing.T) {
	ctx := context.Background()
	client := newClient(t)
	const cedula = "0000000001"
	flujo := "test-desbloquear-" + time.Now().Format("150405.000000")
	r := redisrepo.NewIntentos(client, flujo)
	t.Cleanup(func() {
		client.Del(ctx, "auth:fail:"+flujo+":"+cedula, "auth:lock:"+flujo+":"+cedula)
	})

	if _, err := r.RegistrarFallo(ctx, cedula, time.Minute); err != nil {
		t.Fatal(err)
	}
	if err := r.Bloquear(ctx, cedula, time.Minute); err != nil {
		t.Fatal(err)
	}
	if _, err := r.RegistrarFallo(ctx, cedula, time.Minute); err != nil {
		t.Fatal(err)
	}
	if err := r.Desbloquear(ctx, cedula); err != nil {
		t.Fatal(err)
	}
	if d, _ := r.Bloqueo(ctx, cedula); d != 0 {
		t.Fatalf("sigue bloqueada: %v", d)
	}
	if client.Exists(ctx, "auth:fail:"+flujo+":"+cedula).Val() != 0 {
		t.Fatalf("Desbloquear debe borrar el contador")
	}
}
