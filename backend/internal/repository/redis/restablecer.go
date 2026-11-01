package redis

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	goredis "github.com/redis/go-redis/v9"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/domain"
)

// TokensRestablecer guarda los tokens del enlace para restablecer la contraseña en
// auth:reset:{token} → cédula, con TTL.
type TokensRestablecer struct {
	client *goredis.Client
}

var _ domain.TokensRestablecer = (*TokensRestablecer)(nil)

// NewTokensRestablecer crea el repositorio de tokens de restablecimiento.
func NewTokensRestablecer(client *goredis.Client) *TokensRestablecer {
	return &TokensRestablecer{client: client}
}

func resetKey(token string) string { return "auth:reset:" + token }

// Emitir genera un token aleatorio de 256 bits y lo guarda con la cédula.
func (r *TokensRestablecer) Emitir(ctx context.Context, cedula string, ttl time.Duration) (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generando el token de restablecimiento: %w", err)
	}
	token := hex.EncodeToString(b)
	if err := r.client.Set(ctx, resetKey(token), cedula, ttl).Err(); err != nil {
		return "", fmt.Errorf("guardando el token de restablecimiento: %w", err)
	}
	return token, nil
}

// Consumir lee y borra el token con GETDEL: dos peticiones simultáneas con el mismo
// enlace no pueden usarlo las dos.
func (r *TokensRestablecer) Consumir(ctx context.Context, token string) (string, error) {
	cedula, err := r.client.GetDel(ctx, resetKey(token)).Result()
	if errors.Is(err, goredis.Nil) {
		return "", domain.NotFound("El token no existe o expiró")
	}
	if err != nil {
		return "", fmt.Errorf("leyendo el token de restablecimiento: %w", err)
	}
	return cedula, nil
}
