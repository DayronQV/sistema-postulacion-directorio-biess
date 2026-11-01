package domain

import (
	"context"
	"time"
)

// Intentos es el puerto del contador de intentos fallidos y del bloqueo temporal de
// una cédula. Cada flujo (registro, login, ...) usa su propia instancia para que sus
// contadores no se mezclen. La implementación vive en internal/repository/redis.
type Intentos interface {
	// Bloqueo devuelve el tiempo que le queda al bloqueo de la cédula, o 0 si no
	// está bloqueada.
	Bloqueo(ctx context.Context, cedula string) (time.Duration, error)
	// RegistrarFallo suma un intento fallido y devuelve el total acumulado. El
	// contador expira solo pasada la ventana, contada desde el primer fallo.
	RegistrarFallo(ctx context.Context, cedula string, ventana time.Duration) (int, error)
	// Bloquear bloquea la cédula durante duracion y reinicia su contador.
	Bloquear(ctx context.Context, cedula string, duracion time.Duration) error
	// Reiniciar borra el contador de fallos de la cédula.
	Reiniciar(ctx context.Context, cedula string) error
}

// TicketsRegistro es el puerto de los tickets de identidad validada: prueban que la
// cédula superó el primer paso del registro y que puede elegir correo y contraseña.
type TicketsRegistro interface {
	// Emitir guarda un ticket nuevo para la cédula, válido durante ttl, y devuelve
	// su identificador.
	Emitir(ctx context.Context, cedula string, ttl time.Duration) (string, error)
}

// Postulantes es el puerto de la tabla postulantes. Por ahora solo consulta si la
// cédula ya tiene cuenta; el registro de la cuenta se agrega en su propia issue.
type Postulantes interface {
	ExisteCedula(ctx context.Context, cedula string) (bool, error)
}
