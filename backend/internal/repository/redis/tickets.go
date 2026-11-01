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

// TicketsRegistro guarda los tickets de identidad validada en
// registro:ticket:{id} → cédula, con TTL.
type TicketsRegistro struct {
	client *goredis.Client
}

var _ domain.TicketsRegistro = (*TicketsRegistro)(nil)

// NewTicketsRegistro crea el repositorio de tickets del registro.
func NewTicketsRegistro(client *goredis.Client) *TicketsRegistro {
	return &TicketsRegistro{client: client}
}

func ticketKey(id string) string { return "registro:ticket:" + id }

// Emitir genera un identificador aleatorio de 128 bits y lo guarda con la cédula.
func (r *TicketsRegistro) Emitir(ctx context.Context, cedula string, ttl time.Duration) (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generando el ticket: %w", err)
	}
	id := hex.EncodeToString(b)
	if err := r.client.Set(ctx, ticketKey(id), cedula, ttl).Err(); err != nil {
		return "", fmt.Errorf("guardando el ticket: %w", err)
	}
	return id, nil
}

// Cedula lee el ticket sin borrarlo: si el registro falla por un dato inválido, el
// postulante puede corregirlo sin repetir la validación de identidad.
func (r *TicketsRegistro) Cedula(ctx context.Context, id string) (string, error) {
	cedula, err := r.client.Get(ctx, ticketKey(id)).Result()
	if errors.Is(err, goredis.Nil) {
		return "", domain.NotFound("El ticket no existe o expiró")
	}
	if err != nil {
		return "", fmt.Errorf("leyendo el ticket: %w", err)
	}
	return cedula, nil
}

// Eliminar borra el ticket.
func (r *TicketsRegistro) Eliminar(ctx context.Context, id string) error {
	if err := r.client.Del(ctx, ticketKey(id)).Err(); err != nil {
		return fmt.Errorf("borrando el ticket: %w", err)
	}
	return nil
}
