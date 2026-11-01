package service_test

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/domain"
	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/integration/registrocivil"
	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/service"
)

// fakeTokensRestablecer imita a Redis: Consumir borra el token.
type fakeTokensRestablecer struct {
	tokens map[string]string
	ttls   map[string]time.Duration // ttl con el que se emitió cada token
	n      int
}

func (f *fakeTokensRestablecer) Emitir(_ context.Context, cedula string, ttl time.Duration) (string, error) {
	f.n++
	token := fmt.Sprintf("tok%d", f.n)
	f.tokens[token] = cedula
	f.ttls[token] = ttl
	return token, nil
}

func (f *fakeTokensRestablecer) Consumir(_ context.Context, token string) (string, error) {
	cedula, ok := f.tokens[token]
	if !ok {
		return "", domain.NotFound("El token no existe o expiró")
	}
	delete(f.tokens, token)
	return cedula, nil
}

type fakeMailer struct {
	enviados []domain.Mensaje
}

func (f *fakeMailer) Enviar(_ context.Context, m domain.Mensaje) error {
	if err := m.Validar(); err != nil {
		return err
	}
	f.enviados = append(f.enviados, m)
	return nil
}

const (
	cedulaRecuperacion = registrocivil.CedulaRecuperacion // 22/03/2016, código V3141V5926
	emailRecuperacion  = "gabriela@correo.com"
)

type recuperacionFixture struct {
	svc           *service.RecuperacionService
	postulantes   *fakePostulantes
	intentos      *fakeIntentos
	intentosLogin *fakeIntentos
	tokens        *fakeTokensRestablecer
	mailer        *fakeMailer
}

func newRecuperacionFixture() recuperacionFixture {
	f := recuperacionFixture{
		postulantes: &fakePostulantes{cuentas: map[string]domain.Cuenta{
			cedulaRecuperacion: {ID: "id-r", Cedula: cedulaRecuperacion, Email: emailRecuperacion, PasswordHash: "viejo"},
		}},
		intentos:      newFakeIntentos(),
		intentosLogin: newFakeIntentos(),
		tokens:        &fakeTokensRestablecer{tokens: map[string]string{}, ttls: map[string]time.Duration{}},
		mailer:        &fakeMailer{},
	}
	f.svc = service.NewRecuperacionService(registrocivil.NewMock(), f.postulantes, f.intentos,
		f.intentosLogin, f.tokens, f.mailer, "http://localhost:5173/restablecer")
	return f
}

func recuperar(f recuperacionFixture, in service.IdentidadInput) error {
	return f.svc.Recuperar(context.Background(), in)
}

// tokenDelCorreo extrae el token del enlace del último correo enviado.
func tokenDelCorreo(t *testing.T, f recuperacionFixture) string {
	t.Helper()
	if len(f.mailer.enviados) == 0 {
		t.Fatal("no se envió ningún correo")
	}
	texto := f.mailer.enviados[len(f.mailer.enviados)-1].Texto
	i := strings.Index(texto, "http://localhost:5173/restablecer?token=")
	if i < 0 {
		t.Fatalf("el correo no trae el enlace: %q", texto)
	}
	u, err := url.Parse(strings.Fields(texto[i:])[0])
	if err != nil {
		t.Fatal(err)
	}
	return u.Query().Get("token")
}

func TestRecuperar_EnviaEnlace(t *testing.T) {
	f := newRecuperacionFixture()
	if err := recuperar(f, service.IdentidadInput{Cedula: cedulaRecuperacion, CodigoDactilar: "V3141V5926"}); err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if len(f.mailer.enviados) != 1 || f.mailer.enviados[0].Para != emailRecuperacion {
		t.Fatalf("correos = %+v", f.mailer.enviados)
	}
	if token := tokenDelCorreo(t, f); f.tokens.tokens[token] != cedulaRecuperacion {
		t.Fatalf("el token del enlace no quedó guardado para la cédula")
	}
}

func TestRecuperar_IdentidadCorrectaSinCuentaNoRevelaNada(t *testing.T) {
	f := newRecuperacionFixture()
	// Cédula del Registro Civil sin cuenta en el sistema: misma respuesta, sin correo.
	err := recuperar(f, service.IdentidadInput{Cedula: registrocivil.CedulaFlujoFeliz, CodigoDactilar: "V1234V5678"})
	if err != nil || len(f.mailer.enviados) != 0 {
		t.Fatalf("se esperaba nil y ningún correo: %v, %d correos", err, len(f.mailer.enviados))
	}
}

func TestRecuperar_DatoIncorrectoYBloqueo(t *testing.T) {
	f := newRecuperacionFixture()
	malo := service.IdentidadInput{Cedula: cedulaRecuperacion, CodigoDactilar: "AAAAAAAAAA"}
	for i := 1; i <= 2; i++ {
		err := recuperar(f, malo)
		if !errors.Is(err, domain.ErrUnauthorized) || mensaje(err) != service.MsgIdentidadIncorrecta {
			t.Fatalf("intento %d: se esperaba 401, llegó %v", i, err)
		}
	}
	err := recuperar(f, malo)
	var de *domain.Error
	if !errors.As(err, &de) || !errors.Is(err, domain.ErrLocked) || de.Message != service.MsgRecuperarBloqueado {
		t.Fatalf("tercer intento: se esperaba 429, llegó %v", err)
	}
	// Con el bloqueo vigente, ni el dato correcto envía el correo.
	bueno := service.IdentidadInput{Cedula: cedulaRecuperacion, CodigoDactilar: "V3141V5926"}
	if err := recuperar(f, bueno); !errors.Is(err, domain.ErrLocked) {
		t.Fatalf("se esperaba 429 con el bloqueo vigente, llegó %v", err)
	}
	if len(f.mailer.enviados) != 0 {
		t.Fatalf("no debe enviarse correo durante el bloqueo")
	}
}

func TestRecuperar_CedulaInexistenteCuentaComoFallo(t *testing.T) {
	f := newRecuperacionFixture()
	err := recuperar(f, service.IdentidadInput{Cedula: "1710034065", FechaExpedicion: "01/01/2020"})
	if !errors.Is(err, domain.ErrUnauthorized) || f.intentos.fallos["1710034065"] != 1 {
		t.Fatalf("se esperaba 401 y un fallo contado, llegó %v", err)
	}
}

func TestRestablecer_CambiaPasswordUnaSolaVez(t *testing.T) {
	f := newRecuperacionFixture()
	if err := recuperar(f, service.IdentidadInput{Cedula: cedulaRecuperacion, FechaExpedicion: "22/03/2016"}); err != nil {
		t.Fatal(err)
	}
	token := tokenDelCorreo(t, f)
	f.intentosLogin.fallos[cedulaRecuperacion] = 2
	f.intentosLogin.bloqueos[cedulaRecuperacion] = 5 * time.Minute

	if err := f.svc.Restablecer(context.Background(), token, "nueva1234"); err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	hash := f.postulantes.cuentas[cedulaRecuperacion].PasswordHash
	if bcrypt.CompareHashAndPassword([]byte(hash), []byte("nueva1234")) != nil {
		t.Fatalf("la contraseña nueva no se guardó como hash bcrypt")
	}
	if f.intentosLogin.fallos[cedulaRecuperacion] != 0 || f.intentosLogin.bloqueos[cedulaRecuperacion] != 0 {
		t.Fatalf("no se borraron los bloqueos de login")
	}

	err := f.svc.Restablecer(context.Background(), token, "otra12345")
	if !errors.Is(err, domain.ErrUnauthorized) || mensaje(err) != service.MsgEnlaceInvalido {
		t.Fatalf("el enlace no debe servir dos veces, llegó %v", err)
	}
}

func TestRestablecer_TokenInvalido(t *testing.T) {
	f := newRecuperacionFixture()
	for _, token := range []string{"", "no-existe"} {
		if err := f.svc.Restablecer(context.Background(), token, "nueva1234"); !errors.Is(err, domain.ErrUnauthorized) {
			t.Fatalf("token %q: se esperaba 401, llegó %v", token, err)
		}
	}
}

func TestRestablecer_PasswordInvalidaNoGastaElToken(t *testing.T) {
	f := newRecuperacionFixture()
	f.tokens.tokens["tok"] = cedulaRecuperacion
	err := f.svc.Restablecer(context.Background(), "tok", "corta1")
	var de *domain.Error
	if !errors.As(err, &de) || !errors.Is(err, domain.ErrValidation) || de.Fields["password"] == "" {
		t.Fatalf("se esperaba 400 con fields.password, llegó %v", err)
	}
	if _, ok := f.tokens.tokens["tok"]; !ok {
		t.Fatalf("una contraseña inválida no debe gastar el enlace")
	}
}

func TestRecuperar_EnlaceVence30Minutos(t *testing.T) {
	f := newRecuperacionFixture()
	if err := recuperar(f, service.IdentidadInput{Cedula: cedulaRecuperacion, CodigoDactilar: "V3141V5926"}); err != nil {
		t.Fatal(err)
	}
	if ttl := f.tokens.ttls[tokenDelCorreo(t, f)]; ttl != 30*time.Minute {
		t.Fatalf("el enlace se emitió con TTL %v, se esperaban 30 minutos", ttl)
	}
}

func TestRecuperar_AciertoReiniciaElContador(t *testing.T) {
	f := newRecuperacionFixture()
	malo := service.IdentidadInput{Cedula: cedulaRecuperacion, CodigoDactilar: "AAAAAAAAAA"}
	bueno := service.IdentidadInput{Cedula: cedulaRecuperacion, CodigoDactilar: "V3141V5926"}
	for i := 1; i <= 2; i++ {
		if err := recuperar(f, malo); !errors.Is(err, domain.ErrUnauthorized) {
			t.Fatalf("fallo %d: se esperaba 401, llegó %v", i, err)
		}
	}
	if err := recuperar(f, bueno); err != nil {
		t.Fatalf("el acierto debe responder sin error: %v", err)
	}
	// Tras el acierto el contador vuelve a cero: un fallo más es el primero, no el tercero.
	if err := recuperar(f, malo); !errors.Is(err, domain.ErrUnauthorized) {
		t.Fatalf("se esperaba 401 y no un bloqueo, llegó %v", err)
	}
}

// intentosQueFallanAlDesbloquear simula que Redis falla justo al borrar el bloqueo.
type intentosQueFallanAlDesbloquear struct{ *fakeIntentos }

func (intentosQueFallanAlDesbloquear) Desbloquear(context.Context, string) error {
	return errors.New("redis caído")
}

func TestRestablecer_FalloAlDesbloquearNoAnulaElCambio(t *testing.T) {
	f := newRecuperacionFixture()
	f.svc = service.NewRecuperacionService(registrocivil.NewMock(), f.postulantes, f.intentos,
		intentosQueFallanAlDesbloquear{newFakeIntentos()}, f.tokens, f.mailer, "http://localhost:5173/restablecer")
	f.tokens.tokens["tok"] = cedulaRecuperacion

	if err := f.svc.Restablecer(context.Background(), "tok", "nueva1234"); err != nil {
		t.Fatalf("la contraseña ya cambió: no debe responder error, llegó %v", err)
	}
	hash := f.postulantes.cuentas[cedulaRecuperacion].PasswordHash
	if bcrypt.CompareHashAndPassword([]byte(hash), []byte("nueva1234")) != nil {
		t.Fatalf("la contraseña nueva no se guardó")
	}
}
