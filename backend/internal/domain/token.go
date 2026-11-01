package domain

import "time"

// MinJWTSecretLength es la longitud mínima, en caracteres, de JWT_SECRET. Un secreto
// corto se adivina por fuerza bruta y permitiría falsificar tokens.
const MinJWTSecretLength = 32

// Rol es el perfil con el que un usuario entra al sistema.
type Rol string

const (
	RolPostulante    Rol = "postulante"
	RolAnalista      Rol = "analista"
	RolAdministrador Rol = "administrador"
)

// Valido indica si es uno de los tres roles del sistema.
func (r Rol) Valido() bool {
	switch r {
	case RolPostulante, RolAnalista, RolAdministrador:
		return true
	}
	return false
}

// Claims es la identidad que viaja dentro del token.
type Claims struct {
	UsuarioID string // id del usuario (claim "sub")
	Cedula    string // solo para postulantes
	Usuario   string // solo para analistas y administradores
	Rol       Rol
}

// TokenEmitido es un token recién firmado y lo que dura.
type TokenEmitido struct {
	Token    string
	ExpiraEn time.Duration
}

// TokenIssuer es el puerto que emite y valida los tokens de sesión. La implementación
// (JWT firmado con HS256) vive en internal/auth; los middlewares y los servicios de
// login dependen solo de esta interfaz.
type TokenIssuer interface {
	// Emitir firma un token para la identidad, con la duración configurada
	// (JWT_TTL_MINUTES). Devuelve un error si los claims no son válidos.
	Emitir(claims Claims) (TokenEmitido, error)

	// Validar comprueba la firma y la expiración del token y devuelve su identidad.
	// Cualquier problema (formato, firma, expirado, rol desconocido) devuelve un error
	// que cumple errors.Is(err, ErrUnauthorized), sin detallar la causa al cliente.
	Validar(token string) (Claims, error)
}
