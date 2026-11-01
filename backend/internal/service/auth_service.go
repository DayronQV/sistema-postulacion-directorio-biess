package service

import (
	"context"
	"errors"
	"fmt"
	"net/url"
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

// Reglas de la recuperación de contraseña (docs/especificacion.md, EF-11). La
// validación de identidad tiene el mismo bloqueo que el registro, con su propio
// contador.
const (
	MaxIntentosRecuperar     = 3
	VentanaIntentosRecuperar = 10 * time.Minute
	BloqueoRecuperar         = 10 * time.Minute
	TTLTokenRestablecer      = 30 * time.Minute
)

// Mensajes de la recuperación de contraseña.
const (
	MsgRecuperarBloqueado = "recuperación fallida, inténtelo nuevamente en 10 minutos"
	MsgEnlaceInvalido     = "El enlace para restablecer la contraseña no es válido o expiró. Solicite uno nuevo"
)

// RecuperacionService implementa "¿Olvidaste tu contraseña?": valida la identidad,
// envía por correo un enlace de un solo uso y cambia la contraseña con ese enlace.
type RecuperacionService struct {
	registroCivil domain.RegistroCivil
	postulantes   domain.Postulantes
	intentos      domain.Intentos // validación de identidad de este flujo
	intentosLogin domain.Intentos // se desbloquea al cambiar la contraseña
	tokens        domain.TokensRestablecer
	mailer        domain.Mailer
	// urlRestablecer es la página del frontend que recibe el token, por ejemplo
	// "http://localhost:5173/restablecer".
	urlRestablecer string
}

// NewRecuperacionService crea el servicio con sus dependencias inyectadas.
func NewRecuperacionService(
	registroCivil domain.RegistroCivil,
	postulantes domain.Postulantes,
	intentos domain.Intentos,
	intentosLogin domain.Intentos,
	tokens domain.TokensRestablecer,
	mailer domain.Mailer,
	urlRestablecer string,
) *RecuperacionService {
	return &RecuperacionService{
		registroCivil:  registroCivil,
		postulantes:    postulantes,
		intentos:       intentos,
		intentosLogin:  intentosLogin,
		tokens:         tokens,
		mailer:         mailer,
		urlRestablecer: urlRestablecer,
	}
}

// Recuperar valida la identidad contra el Registro Civil y, si la cédula tiene
// cuenta, le envía al correo registrado un enlace para restablecer la contraseña.
//
// Un dato de identidad incorrecto responde igual que en el registro (401 y cuenta
// como fallo; al tercero, bloqueo de 10 minutos). Con la identidad correcta la
// respuesta es la misma tenga o no cuenta la cédula: así no se revela quién está
// registrado.
func (s *RecuperacionService) Recuperar(ctx context.Context, in IdentidadInput) error {
	cedula := strings.TrimSpace(in.Cedula)
	if err := domain.ValidarCedula(cedula); err != nil {
		return err
	}
	fecha, codigo, err := parseDatoIdentidad(in)
	if err != nil {
		return err
	}

	restante, err := s.intentos.Bloqueo(ctx, cedula)
	if err != nil {
		return err
	}
	if restante > 0 {
		return domain.Locked(MsgRecuperarBloqueado, restante)
	}
	ok, err := comprobarIdentidad(ctx, s.registroCivil, cedula, fecha, codigo)
	if err != nil {
		return err
	}
	if !ok {
		return s.fallo(ctx, cedula)
	}
	if err := s.intentos.Reiniciar(ctx, cedula); err != nil {
		return err
	}

	cuenta, err := s.postulantes.BuscarPorCedula(ctx, cedula)
	switch {
	case errors.Is(err, domain.ErrNotFound):
		return nil
	case err != nil:
		return err
	}
	token, err := s.tokens.Emitir(ctx, cedula, TTLTokenRestablecer)
	if err != nil {
		return err
	}
	return s.mailer.Enviar(ctx, correoRestablecer(cuenta.Email, s.enlace(token)))
}

// Restablecer cambia la contraseña con el token del enlace. El token sirve una sola
// vez; al terminar se borran los bloqueos de login de la cédula, para que pueda
// entrar enseguida con la contraseña nueva.
func (s *RecuperacionService) Restablecer(ctx context.Context, token, password string) error {
	if msg := validarPassword(password); msg != "" {
		return domain.Validation("Datos inválidos", map[string]string{"password": msg})
	}
	token = strings.TrimSpace(token)
	if token == "" {
		return domain.Unauthorized(MsgEnlaceInvalido)
	}
	cedula, err := s.tokens.Consumir(ctx, token)
	switch {
	case errors.Is(err, domain.ErrNotFound):
		return domain.Unauthorized(MsgEnlaceInvalido)
	case err != nil:
		return err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hasheando la contraseña: %w", err)
	}
	err = s.postulantes.ActualizarPassword(ctx, cedula, string(hash))
	switch {
	case errors.Is(err, domain.ErrNotFound):
		// La cuenta desapareció después de emitir el enlace: el enlace ya no sirve.
		return domain.Unauthorized(MsgEnlaceInvalido)
	case err != nil:
		return err
	}
	return s.intentosLogin.Desbloquear(ctx, cedula)
}

// enlace arma la URL del frontend con el token en la query.
func (s *RecuperacionService) enlace(token string) string {
	return strings.TrimRight(s.urlRestablecer, "/") + "?token=" + url.QueryEscape(token)
}

// fallo registra un intento fallido y bloquea la cédula al llegar al máximo.
func (s *RecuperacionService) fallo(ctx context.Context, cedula string) error {
	n, err := s.intentos.RegistrarFallo(ctx, cedula, VentanaIntentosRecuperar)
	if err != nil {
		return err
	}
	if n >= MaxIntentosRecuperar {
		if err := s.intentos.Bloquear(ctx, cedula, BloqueoRecuperar); err != nil {
			return err
		}
		return domain.Locked(MsgRecuperarBloqueado, BloqueoRecuperar)
	}
	return domain.Unauthorized(MsgIdentidadIncorrecta)
}

func correoRestablecer(para, enlace string) domain.Mensaje {
	return domain.Mensaje{
		Para:   para,
		Asunto: "Restablecer su contraseña - Concurso Directorio BIESS",
		Texto: "Recibimos una solicitud para restablecer la contraseña de su cuenta.\n\n" +
			"Para crear una nueva contraseña, abra este enlace (vence en 30 minutos y sirve una sola vez):\n" +
			enlace + "\n\n" +
			"Si usted no hizo esta solicitud, ignore este correo: su contraseña no cambia.",
	}
}
