package auth_test

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/auth"
	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/domain"
)

const (
	secreto      = "secreto-de-prueba-con-mas-de-32-caracteres"
	otroSecreto  = "otro-secreto-distinto-con-mas-de-32-caracteres"
	duracion     = 30 * time.Minute
	idPostulante = "6b0c1c0e-0000-4000-8000-000000000001"
)

var ahoraBase = time.Date(2026, 11, 1, 10, 0, 0, 0, time.UTC)

// nuevoEmisor crea un emisor con un reloj controlado: el reloj devuelto se puede
// adelantar para probar tokens vencidos.
func nuevoEmisor(t *testing.T) (*auth.JWTIssuer, *time.Time) {
	t.Helper()
	reloj := ahoraBase
	i, err := auth.NewJWTIssuer(secreto, duracion, auth.WithClock(func() time.Time { return reloj }))
	if err != nil {
		t.Fatal(err)
	}
	return i, &reloj
}

func claimsPostulante() domain.Claims {
	return domain.Claims{UsuarioID: idPostulante, Cedula: "1712345675", Rol: domain.RolPostulante}
}

func emitir(t *testing.T, i *auth.JWTIssuer, c domain.Claims) string {
	t.Helper()
	e, err := i.Emitir(c)
	if err != nil {
		t.Fatalf("Emitir: %v", err)
	}
	return e.Token
}

func exigirNoAutorizado(t *testing.T, err error) {
	t.Helper()
	if !errors.Is(err, domain.ErrUnauthorized) {
		t.Fatalf("se esperaba ErrUnauthorized, got %v", err)
	}
}

func TestEmitirYValidarDevuelveLaMismaIdentidad(t *testing.T) {
	i, _ := nuevoEmisor(t)
	casos := []domain.Claims{
		claimsPostulante(),
		{UsuarioID: "admin-1", Usuario: "comision", Rol: domain.RolAdministrador},
		{UsuarioID: "analista-1", Usuario: "talento.humano", Rol: domain.RolAnalista},
	}
	for _, c := range casos {
		t.Run(string(c.Rol), func(t *testing.T) {
			got, err := i.Validar(emitir(t, i, c))
			if err != nil {
				t.Fatalf("Validar: %v", err)
			}
			if got != c {
				t.Fatalf("identidad distinta: se emitió %+v y se validó %+v", c, got)
			}
		})
	}
}

func TestEmitirInformaLaDuracion(t *testing.T) {
	i, _ := nuevoEmisor(t)
	e, err := i.Emitir(claimsPostulante())
	if err != nil {
		t.Fatal(err)
	}
	if e.ExpiraEn != duracion {
		t.Fatalf("ExpiraEn = %v, se esperaba %v", e.ExpiraEn, duracion)
	}
}

func TestTokenVencido(t *testing.T) {
	i, reloj := nuevoEmisor(t)
	token := emitir(t, i, claimsPostulante())

	*reloj = ahoraBase.Add(duracion - time.Second)
	if _, err := i.Validar(token); err != nil {
		t.Fatalf("un segundo antes de vencer el token debe valer: %v", err)
	}

	*reloj = ahoraBase.Add(duracion + time.Second)
	_, err := i.Validar(token)
	exigirNoAutorizado(t, err)
}

func TestFirmaAlterada(t *testing.T) {
	i, _ := nuevoEmisor(t)
	token := emitir(t, i, claimsPostulante())
	partes := strings.Split(token, ".")

	// Cambia un carácter del medio de la firma (el último puede ser relleno y no
	// alterar los bytes).
	firma := []byte(partes[2])
	if firma[10] == 'A' {
		firma[10] = 'B'
	} else {
		firma[10] = 'A'
	}
	alterado := partes[0] + "." + partes[1] + "." + string(firma)

	_, err := i.Validar(alterado)
	exigirNoAutorizado(t, err)
}

func TestContenidoAlteradoConLaFirmaOriginal(t *testing.T) {
	i, _ := nuevoEmisor(t)
	token := emitir(t, i, claimsPostulante())
	partes := strings.Split(token, ".")

	// Un postulante intenta subirse el rol a administrador sin conocer el secreto.
	payload, err := base64.RawURLEncoding.DecodeString(partes[1])
	if err != nil {
		t.Fatal(err)
	}
	var campos map[string]any
	if err := json.Unmarshal(payload, &campos); err != nil {
		t.Fatal(err)
	}
	campos["rol"] = string(domain.RolAdministrador)
	nuevo, _ := json.Marshal(campos)
	alterado := partes[0] + "." + base64.RawURLEncoding.EncodeToString(nuevo) + "." + partes[2]

	_, err = i.Validar(alterado)
	exigirNoAutorizado(t, err)
}

func TestFirmadoConOtroSecreto(t *testing.T) {
	i, _ := nuevoEmisor(t)
	ajeno, err := auth.NewJWTIssuer(otroSecreto, duracion, auth.WithClock(func() time.Time { return ahoraBase }))
	if err != nil {
		t.Fatal(err)
	}
	_, err = i.Validar(emitir(t, ajeno, claimsPostulante()))
	exigirNoAutorizado(t, err)
}

// Tokens armados a mano para comprobar que el validador no acepta atajos.
func TestTokensMalFormadosOMalFirmados(t *testing.T) {
	i, _ := nuevoEmisor(t)
	vigente := jwt.NewNumericDate(ahoraBase.Add(duracion))
	valido := func() jwt.MapClaims {
		return jwt.MapClaims{"sub": idPostulante, "rol": "postulante", "exp": vigente}
	}
	firmar := func(m jwt.SigningMethod, claims jwt.MapClaims, clave any) string {
		s, err := jwt.NewWithClaims(m, claims).SignedString(clave)
		if err != nil {
			t.Fatal(err)
		}
		return s
	}

	sinExp := valido()
	delete(sinExp, "exp")
	sinSub := valido()
	delete(sinSub, "sub")
	rolDesconocido := valido()
	rolDesconocido["rol"] = "superusuario"
	sinRol := valido()
	delete(sinRol, "rol")

	casos := map[string]string{
		"algoritmo none (sin firma)":  firmar(jwt.SigningMethodNone, valido(), jwt.UnsafeAllowNoneSignatureType),
		"otro algoritmo HMAC (HS512)": firmar(jwt.SigningMethodHS512, valido(), []byte(secreto)),
		"sin fecha de expiración":     firmar(jwt.SigningMethodHS256, sinExp, []byte(secreto)),
		"sin sub":                     firmar(jwt.SigningMethodHS256, sinSub, []byte(secreto)),
		"rol desconocido":             firmar(jwt.SigningMethodHS256, rolDesconocido, []byte(secreto)),
		"sin rol":                     firmar(jwt.SigningMethodHS256, sinRol, []byte(secreto)),
		"cadena vacía":                "",
		"texto cualquiera":            "esto-no-es-un-token",
		"tres partes sin sentido":     "a.b.c",
		"solo cabecera y contenido":   "eyJhbGciOiJIUzI1NiJ9.eyJzdWIiOiIxIn0",
	}
	for nombre, token := range casos {
		t.Run(nombre, func(t *testing.T) {
			_, err := i.Validar(token)
			exigirNoAutorizado(t, err)
		})
	}

	// Control: el mismo token bien armado sí vale, para saber que los rechazos de
	// arriba se deben a su defecto y no a un error del armado.
	if _, err := i.Validar(firmar(jwt.SigningMethodHS256, valido(), []byte(secreto))); err != nil {
		t.Fatalf("el token de control debe ser válido: %v", err)
	}
}

func TestElErrorNoDetallaLaCausa(t *testing.T) {
	i, reloj := nuevoEmisor(t)
	token := emitir(t, i, claimsPostulante())
	*reloj = ahoraBase.Add(2 * duracion)

	_, venc := i.Validar(token)
	_, basura := i.Validar("basura")
	var a, b *domain.Error
	if !errors.As(venc, &a) || !errors.As(basura, &b) {
		t.Fatal("se esperaban errores de dominio")
	}
	if a.Message != b.Message {
		t.Fatalf("el mensaje distingue la causa: %q frente a %q", a.Message, b.Message)
	}
}

func TestEmitirRechazaClaimsInvalidos(t *testing.T) {
	i, _ := nuevoEmisor(t)
	casos := map[string]domain.Claims{
		"sin id de usuario": {Cedula: "1712345675", Rol: domain.RolPostulante},
		"sin rol":           {UsuarioID: idPostulante},
		"rol desconocido":   {UsuarioID: idPostulante, Rol: "superusuario"},
	}
	for nombre, c := range casos {
		t.Run(nombre, func(t *testing.T) {
			if _, err := i.Emitir(c); err == nil {
				t.Fatal("se esperaba un error")
			}
		})
	}
}

func TestNewJWTIssuerValidaLosParametros(t *testing.T) {
	if _, err := auth.NewJWTIssuer(strings.Repeat("a", domain.MinJWTSecretLength-1), duracion); err == nil {
		t.Error("un secreto de 31 caracteres debe rechazarse")
	}
	if _, err := auth.NewJWTIssuer(strings.Repeat("a", domain.MinJWTSecretLength), duracion); err != nil {
		t.Errorf("un secreto de 32 caracteres debe aceptarse: %v", err)
	}
	for _, ttl := range []time.Duration{0, -time.Minute} {
		if _, err := auth.NewJWTIssuer(secreto, ttl); err == nil {
			t.Errorf("la duración %v debe rechazarse", ttl)
		}
	}
}
