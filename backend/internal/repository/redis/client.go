// Package redis implementa los puertos que usan Redis: intentos y bloqueo de
// autenticación, tickets temporales y caché de catálogos.
package redis

import (
	"context"
	"fmt"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

const connectTimeout = 5 * time.Second

// NewClient crea el cliente de Redis y comprueba que el servidor responda.
func NewClient(ctx context.Context, addr string) (*goredis.Client, error) {
	client := goredis.NewClient(&goredis.Options{Addr: addr})

	ctx, cancel := context.WithTimeout(ctx, connectTimeout)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("Redis no responde en %s (¿está levantado con docker compose up -d?): %w", addr, err)
	}
	return client, nil
}
