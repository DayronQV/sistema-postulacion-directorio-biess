package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/auth"
	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/domain"
	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/service"
)

const passwordCorrecta = "clave1234"

type loginFixture struct {
	svc         *service.AuthService
	intentos    *fakeIntentos
	postulantes *fakePostulantes
	tokens      *auth.JWTIssuer
}

// newLoginFixture registra una cuenta para la cédula de prueba, con passwordCorrecta.
func newLoginFixture(t *testing.T) loginFixture {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte(passwordCorrecta), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	tokens, err := auth.NewJWTIssuer("secreto-de-prueba-con-mas-de-32-caracteres", time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	f := loginFixture{
		intentos: newFakeIntentos(),
		postulantes: &fakePostulantes{
			cuentas:        map[string]domain.Cuenta{cedula: {ID: "id-1", Cedula: cedula, PasswordHash: string(hash)}},
			conPostulacion: map[string]bool{},
		},
		tokens: tokens,
	}
	f.svc = service.NewAuthService(f.postulantes, f.intentos, f.tokens)
	return f
}

func login(f loginFixture, password string) (service.LoginResult, error) {
	return f.svc.Login(context.Background(), service.LoginInput{Cedula: cedula, Password: password})
}

func TestLogin_Exito(t *testing.T) {
	f := newLoginFixture(t)
	f.intentos.fallos[cedula] = 2 // un login correcto antes del tercero reinicia el contador
	res, err := login(f, passwordCorrecta)
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if res.ExpiraEn != time.Hour || !res.PrimerIngreso {
		t.Fatalf("resultado = %+v", res)
	}
	if f.intentos.fallos[cedula] != 0 {
		t.Fatalf("el contador no se reinició")
	}
	// El token emitido sirve para una ruta protegida: lo valida el mismo emisor.
	claims, err := f.tokens.Validar(res.Token)
	if err != nil {
		t.Fatalf("el token no es válido: %v", err)
	}
	if claims.UsuarioID != "id-1" || claims.Cedula != cedula || claims.Rol != domain.RolPostulante {
		t.Fatalf("claims = %+v", claims)
	}
}

func TestLogin_PrimerIngreso(t *testing.T) {
	f := newLoginFixture(t)
	f.postulantes.conPostulacion["id-1"] = true
	res, err := login(f, passwordCorrecta)
	if err != nil || res.PrimerIngreso {
		t.Fatalf("con postulación no es primer ingreso: %+v, %v", res, err)
	}
}

func TestLogin_PasswordIncorrecta(t *testing.T) {
	f := newLoginFixture(t)
	_, err := login(f, "otra1234")
	if !errors.Is(err, domain.ErrUnauthorized) || mensaje(err) != service.MsgCredencialesIncorrectas {
		t.Fatalf("se esperaba 401 %q, llegó %v", service.MsgCredencialesIncorrectas, err)
	}
	if f.intentos.fallos[cedula] != 1 {
		t.Fatalf("fallos = %d, se esperaba 1", f.intentos.fallos[cedula])
	}
}

func TestLogin_CedulaSinCuentaMismoMensaje(t *testing.T) {
	f := newLoginFixture(t)
	const sinCuenta = "1712345675"
	_, err := f.svc.Login(context.Background(), service.LoginInput{Cedula: sinCuenta, Password: passwordCorrecta})
	if !errors.Is(err, domain.ErrUnauthorized) || mensaje(err) != service.MsgCredencialesIncorrectas {
		t.Fatalf("una cédula sin cuenta debe responder igual que una contraseña incorrecta, llegó %v", err)
	}
	if f.intentos.fallos[sinCuenta] != 1 {
		t.Fatalf("la cédula sin cuenta no contó como fallo")
	}
}

func TestLogin_TercerFalloBloquea(t *testing.T) {
	f := newLoginFixture(t)
	for i := 1; i <= 2; i++ {
		if _, err := login(f, "mala1234"); !errors.Is(err, domain.ErrUnauthorized) {
			t.Fatalf("intento %d: se esperaba 401, llegó %v", i, err)
		}
	}
	_, err := login(f, "mala1234")
	var de *domain.Error
	if !errors.As(err, &de) || !errors.Is(err, domain.ErrLocked) || de.RetryAfter != service.BloqueoLogin {
		t.Fatalf("tercer intento: se esperaba 429 de 10 minutos, llegó %v", err)
	}
	if f.intentos.bloqueos[cedula] != service.BloqueoLogin {
		t.Fatalf("no se creó el bloqueo")
	}
}

func TestLogin_BloqueoVigenteAunqueLaPasswordSeaCorrecta(t *testing.T) {
	f := newLoginFixture(t)
	f.intentos.bloqueos[cedula] = 3 * time.Minute
	_, err := login(f, passwordCorrecta)
	var de *domain.Error
	if !errors.As(err, &de) || !errors.Is(err, domain.ErrLocked) || de.RetryAfter != 3*time.Minute {
		t.Fatalf("se esperaba 429 con el TTL restante, llegó %v", err)
	}
}

func TestLogin_CedulaMalFormada(t *testing.T) {
	f := newLoginFixture(t)
	_, err := f.svc.Login(context.Background(), service.LoginInput{Cedula: "123", Password: passwordCorrecta})
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("se esperaba 400, llegó %v", err)
	}
	if len(f.intentos.fallos) != 0 {
		t.Fatalf("un error de formato no debe contar como intento")
	}
}
