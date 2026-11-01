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
	// Cedula devuelve la cédula del ticket, o un error que cumple
	// errors.Is(err, ErrNotFound) si no existe o ya expiró.
	Cedula(ctx context.Context, id string) (string, error)
	// Eliminar borra el ticket para que no se pueda reutilizar.
	Eliminar(ctx context.Context, id string) error
}

// NuevoPostulante son los datos de una cuenta nueva. La contraseña llega ya
// hasheada: el repositorio nunca ve la contraseña en texto plano.
type NuevoPostulante struct {
	Cedula       string
	Email        string
	PasswordHash string
}

// Mensajes de una cuenta que ya existe.
const (
	MsgCedulaRegistrada = "La cédula ya se encuentra registrada"
	MsgEmailRegistrado  = "El correo electrónico ya se encuentra registrado"
)

// ConflictoCedula es el 409 de una cédula que ya tiene cuenta.
func ConflictoCedula() error {
	return Conflict(MsgCedulaRegistrada, map[string]string{"cedula": MsgCedulaRegistrada})
}

// ConflictoEmail es el 409 de un correo que ya pertenece a una cuenta.
func ConflictoEmail() error {
	return Conflict(MsgEmailRegistrado, map[string]string{"email": MsgEmailRegistrado})
}

// Postulantes es el puerto de la tabla postulantes.
type Postulantes interface {
	ExisteCedula(ctx context.Context, cedula string) (bool, error)
	ExisteEmail(ctx context.Context, email string) (bool, error)
	// Crear guarda la cuenta y devuelve su id. Si la cédula o el correo ya existen
	// devuelve un error ErrConflict con el campo correspondiente en Fields.
	Crear(ctx context.Context, p NuevoPostulante) (string, error)
}
