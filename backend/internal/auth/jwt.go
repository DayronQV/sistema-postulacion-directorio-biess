// Package auth implementa el emisor de tokens de sesión (domain.TokenIssuer) con JWT.
package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/domain"
)

// mensajeSesionInvalida es lo único que se le dice al cliente cuando un token no sirve:
// no se detalla si falló la firma, el formato o la expiración.
const mensajeSesionInvalida = "Su sesión no es válida o expiró. Inicie sesión nuevamente"

// JWTIssuer firma y valida tokens JWT con HS256.
type JWTIssuer struct {
	secret []byte
	ttl    time.Duration
	now    func() time.Time
}

var _ domain.TokenIssuer = (*JWTIssuer)(nil)

// Option ajusta un JWTIssuer al crearlo.
type Option func(*JWTIssuer)

// WithClock cambia el reloj del emisor. Sirve para probar tokens vencidos sin esperar.
func WithClock(now func() time.Time) Option {
	return func(i *JWTIssuer) { i.now = now }
}

// NewJWTIssuer crea el emisor. El secreto debe tener al menos
// domain.MinJWTSecretLength caracteres y ttl debe ser positivo.
func NewJWTIssuer(secret string, ttl time.Duration, opts ...Option) (*JWTIssuer, error) {
	if len(secret) < domain.MinJWTSecretLength {
		return nil, fmt.Errorf("JWT_SECRET debe tener al menos %d caracteres", domain.MinJWTSecretLength)
	}
	if ttl <= 0 {
		return nil, errors.New("la duración del token debe ser positiva")
	}
	i := &JWTIssuer{secret: []byte(secret), ttl: ttl, now: time.Now}
	for _, opt := range opts {
		opt(i)
	}
	return i, nil
}

// tokenClaims es el contenido del JWT: los claims registrados (sub, iat, exp) más los
// propios del sistema.
type tokenClaims struct {
	Cedula  string     `json:"cedula,omitempty"`
	Usuario string     `json:"usuario,omitempty"`
	Rol     domain.Rol `json:"rol"`
	jwt.RegisteredClaims
}

// Emitir firma un token que vence ttl después de ahora.
func (i *JWTIssuer) Emitir(c domain.Claims) (domain.TokenEmitido, error) {
	if c.UsuarioID == "" {
		return domain.TokenEmitido{}, errors.New("el token necesita el id del usuario")
	}
	if !c.Rol.Valido() {
		return domain.TokenEmitido{}, fmt.Errorf("rol no válido: %q", c.Rol)
	}

	ahora := i.now()
	claims := tokenClaims{
		Cedula:  c.Cedula,
		Usuario: c.Usuario,
		Rol:     c.Rol,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   c.UsuarioID,
			IssuedAt:  jwt.NewNumericDate(ahora),
			ExpiresAt: jwt.NewNumericDate(ahora.Add(i.ttl)),
		},
	}
	firmado, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(i.secret)
	if err != nil {
		return domain.TokenEmitido{}, fmt.Errorf("firmando el token: %w", err)
	}
	return domain.TokenEmitido{Token: firmado, ExpiraEn: i.ttl}, nil
}

// Validar comprueba el token y devuelve su identidad. Solo acepta HS256 y exige que
// tenga fecha de expiración, así que un token "sin firma" (alg none), firmado con otro
// algoritmo o sin vencimiento se rechaza.
func (i *JWTIssuer) Validar(token string) (domain.Claims, error) {
	parser := jwt.NewParser(
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithExpirationRequired(),
		jwt.WithTimeFunc(i.now),
	)

	var tc tokenClaims
	_, err := parser.ParseWithClaims(token, &tc, func(*jwt.Token) (any, error) {
		return i.secret, nil
	})
	if err != nil || tc.Subject == "" || !tc.Rol.Valido() {
		return domain.Claims{}, domain.Unauthorized(mensajeSesionInvalida)
	}

	return domain.Claims{
		UsuarioID: tc.Subject,
		Cedula:    tc.Cedula,
		Usuario:   tc.Usuario,
		Rol:       tc.Rol,
	}, nil
}
