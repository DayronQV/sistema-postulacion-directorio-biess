package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/domain"
)

// Reglas del inicio de sesión del postulante (docs/especificacion.md, EF-11).
const (
	MaxIntentosLogin     = 3
	VentanaIntentosLogin = 10 * time.Minute
	BloqueoLogin         = 10 * time.Minute
)

// Mensajes del inicio de sesión. El de credenciales es literal del requerimiento.
const (
	MsgCredencialesIncorrectas = "sus credenciales de ingreso son incorrectas"
	MsgLoginBloqueado          = "Cuenta bloqueada por intentos fallidos, inténtelo nuevamente en 10 minutos"
)

// LoginInput son las credenciales del postulante.
type LoginInput struct {
	Cedula   string
	Password string
}

// LoginResult es la sesión emitida.
type LoginResult struct {
	Token         string
	ExpiraEn      time.Duration
	PrimerIngreso bool // aún no tiene postulación: debe ir al formulario
}

// AuthService implementa el inicio de sesión del postulante.
type AuthService struct {
	postulantes domain.Postulantes
	intentos    domain.Intentos
	tokens      domain.TokenIssuer
	// hashFicticio se compara cuando la cédula no tiene cuenta, para que esa respuesta
	// tarde lo mismo que una contraseña incorrecta y no delate qué cédulas existen.
	hashFicticio []byte
}

// NewAuthService crea el servicio con sus dependencias inyectadas.
func NewAuthService(postulantes domain.Postulantes, intentos domain.Intentos, tokens domain.TokenIssuer) *AuthService {
	hash, err := bcrypt.GenerateFromPassword([]byte("contraseña-ficticia-sin-cuenta"), bcrypt.DefaultCost)
	if err != nil {
		// Solo falla si el costo no es válido, y DefaultCost lo es.
		panic(fmt.Sprintf("generando el hash ficticio: %v", err))
	}
	return &AuthService{postulantes: postulantes, intentos: intentos, tokens: tokens, hashFicticio: hash}
}

// Login valida cédula y contraseña y emite el token de sesión con rol postulante.
//
// Al tercer fallo consecutivo la cédula queda bloqueada 10 minutos y, mientras dure,
// cualquier intento responde LOCKED aunque la contraseña sea correcta. Una cédula sin
// cuenta responde igual que una contraseña incorrecta y también cuenta como fallo.
// Un login correcto reinicia el contador.
func (s *AuthService) Login(ctx context.Context, in LoginInput) (LoginResult, error) {
	cedula := strings.TrimSpace(in.Cedula)
	if err := domain.ValidarCedula(cedula); err != nil {
		return LoginResult{}, err
	}

	restante, err := s.intentos.Bloqueo(ctx, cedula)
	if err != nil {
		return LoginResult{}, err
	}
	if restante > 0 {
		return LoginResult{}, domain.Locked(MsgLoginBloqueado, restante)
	}

	cuenta, err := s.postulantes.BuscarPorCedula(ctx, cedula)
	switch {
	case errors.Is(err, domain.ErrNotFound):
		_ = bcrypt.CompareHashAndPassword(s.hashFicticio, []byte(in.Password))
		return LoginResult{}, s.fallo(ctx, cedula)
	case err != nil:
		return LoginResult{}, err
	}
	if bcrypt.CompareHashAndPassword([]byte(cuenta.PasswordHash), []byte(in.Password)) != nil {
		return LoginResult{}, s.fallo(ctx, cedula)
	}

	if err := s.intentos.Reiniciar(ctx, cedula); err != nil {
		return LoginResult{}, err
	}
	tiene, err := s.postulantes.TienePostulacion(ctx, cuenta.ID)
	if err != nil {
		return LoginResult{}, err
	}
	emitido, err := s.tokens.Emitir(domain.Claims{
		UsuarioID: cuenta.ID,
		Cedula:    cuenta.Cedula,
		Rol:       domain.RolPostulante,
	})
	if err != nil {
		return LoginResult{}, err
	}
	return LoginResult{Token: emitido.Token, ExpiraEn: emitido.ExpiraEn, PrimerIngreso: !tiene}, nil
}

// fallo registra un intento fallido y bloquea la cédula al llegar al máximo.
func (s *AuthService) fallo(ctx context.Context, cedula string) error {
	n, err := s.intentos.RegistrarFallo(ctx, cedula, VentanaIntentosLogin)
	if err != nil {
		return err
	}
	if n >= MaxIntentosLogin {
		if err := s.intentos.Bloquear(ctx, cedula, BloqueoLogin); err != nil {
			return err
		}
		return domain.Locked(MsgLoginBloqueado, BloqueoLogin)
	}
	return domain.Unauthorized(MsgCredencialesIncorrectas)
}
