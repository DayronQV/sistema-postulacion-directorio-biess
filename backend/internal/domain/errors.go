// Package domain contiene las entidades, los errores y los puertos (interfaces)
// del sistema. No depende de HTTP, SQL ni de ninguna librería externa.
package domain

import (
	"errors"
	"time"
)

// Errores base del dominio. Los servicios devuelven errores construidos con los
// helpers de abajo (Validation, NotFound, ...) y la capa HTTP los traduce al formato
// de error de docs/api.md. Se comparan con errors.Is.
var (
	ErrValidation    = errors.New("validation error")     // 400 VALIDATION_ERROR
	ErrUnprocessable = errors.New("unprocessable entity") // 422 VALIDATION_ERROR (regla de negocio)
	ErrUnauthorized  = errors.New("unauthorized")         // 401 UNAUTHORIZED
	ErrForbidden     = errors.New("forbidden")            // 403 FORBIDDEN
	ErrNotFound      = errors.New("not found")            // 404 NOT_FOUND
	ErrConflict      = errors.New("conflict")             // 409 CONFLICT
	ErrLocked        = errors.New("locked")               // 429 LOCKED
)

// Error es un error de dominio con mensaje para el usuario. Envuelve a uno de los
// errores base para que errors.Is(err, domain.ErrNotFound) funcione.
type Error struct {
	Base       error             // uno de los errores base
	Message    string            // mensaje en español para mostrar al usuario
	Fields     map[string]string // errores por campo (opcional)
	RetryAfter time.Duration     // tiempo de espera, solo para ErrLocked (opcional)
}

func (e *Error) Error() string { return e.Message }

// Unwrap permite errors.Is con el error base.
func (e *Error) Unwrap() error { return e.Base }

// Validation indica datos inválidos (400). fields puede ser nil.
func Validation(message string, fields map[string]string) error {
	return &Error{Base: ErrValidation, Message: message, Fields: fields}
}

// Unprocessable indica una regla de negocio incumplida (422), por ejemplo
// "nacionalidad no ecuatoriana".
func Unprocessable(message string, fields map[string]string) error {
	return &Error{Base: ErrUnprocessable, Message: message, Fields: fields}
}

// Unauthorized indica credenciales o sesión inválidas (401).
func Unauthorized(message string) error {
	return &Error{Base: ErrUnauthorized, Message: message}
}

// Forbidden indica que el usuario no tiene permiso (403).
func Forbidden(message string) error {
	return &Error{Base: ErrForbidden, Message: message}
}

// NotFound indica que el recurso no existe (404).
func NotFound(message string) error {
	return &Error{Base: ErrNotFound, Message: message}
}

// Conflict indica que el recurso ya existe o que el estado no lo permite (409).
func Conflict(message string, fields map[string]string) error {
	return &Error{Base: ErrConflict, Message: message, Fields: fields}
}

// Locked indica bloqueo temporal por intentos fallidos (429). retryAfter es el
// tiempo restante del bloqueo.
func Locked(message string, retryAfter time.Duration) error {
	return &Error{Base: ErrLocked, Message: message, RetryAfter: retryAfter}
}
