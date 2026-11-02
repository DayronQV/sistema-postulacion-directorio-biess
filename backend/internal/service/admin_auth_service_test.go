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

type fakeUsuariosAdmin struct {
	usuarios map[string]domain.UsuarioAdmin
}

func (f *fakeUsuariosAdmin) BuscarPorUsuario(_ context.Context, usuario string) (domain.UsuarioAdmin, error) {
	u, ok := f.usuarios[usuario]
	if !ok {
		return domain.UsuarioAdmin{}, domain.NotFound("El usuario no existe")
	}
	return u, nil
}

type adminFixture struct {
	svc      *service.AdminAuthService
	intentos *fakeIntentos
	tokens   *auth.JWTIssuer
}

func newAdminFixture(t *testing.T) adminFixture {
	t.Helper()
	hash := func(p string) string {
		h, err := bcrypt.GenerateFromPassword([]byte(p), bcrypt.MinCost)
		if err != nil {
			t.Fatal(err)
		}
		return string(h)
	}
	tokens, err := auth.NewJWTIssuer("secreto-de-prueba-con-mas-de-32-caracteres", time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	f := adminFixture{intentos: newFakeIntentos(), tokens: tokens}
	usuarios := &fakeUsuariosAdmin{usuarios: map[string]domain.UsuarioAdmin{
		"analista":      {ID: "a-1", Usuario: "analista", PasswordHash: hash("Analista2026"), Rol: domain.RolAnalista},
		"administrador": {ID: "a-2", Usuario: "administrador", PasswordHash: hash("Admin2026"), Rol: domain.RolAdministrador},
	}}
	f.svc = service.NewAdminAuthService(usuarios, f.intentos, tokens)
	return f
}

func loginAdmin(f adminFixture, usuario, password string) (service.AdminSesion, error) {
	return f.svc.Login(context.Background(), service.AdminLoginInput{Usuario: usuario, Password: password})
}

func TestAdminLogin_EmiteTokenConElRolDelUsuario(t *testing.T) {
	f := newAdminFixture(t)
	for usuario, datos := range map[string]struct {
		password string
		rol      domain.Rol
	}{
		"analista":      {"Analista2026", domain.RolAnalista},
		"administrador": {"Admin2026", domain.RolAdministrador},
	} {
		t.Run(usuario, func(t *testing.T) {
			// El usuario se compara sin importar mayúsculas ni espacios.
			s, err := loginAdmin(f, "  "+usuario+" ", datos.password)
			if err != nil {
				t.Fatalf("error inesperado: %v", err)
			}
			claims, err := f.tokens.Validar(s.Token)
			if err != nil {
				t.Fatalf("token inválido: %v", err)
			}
			if claims.Rol != datos.rol || claims.Usuario != usuario || s.Rol != datos.rol || s.ExpiraEn != time.Hour {
				t.Fatalf("sesión = %+v, claims = %+v", s, claims)
			}
		})
	}
}

func TestAdminLogin_CredencialesIncorrectasMismoMensaje(t *testing.T) {
	f := newAdminFixture(t)
	for _, c := range []struct{ usuario, password string }{
		{"analista", "otra1234"},
		{"no-existe", "Analista2026"},
	} {
		_, err := loginAdmin(f, c.usuario, c.password)
		if !errors.Is(err, domain.ErrUnauthorized) || mensaje(err) != service.MsgCredencialesIncorrectas {
			t.Fatalf("%s: se esperaba 401 con el mensaje del requerimiento, llegó %v", c.usuario, err)
		}
		if f.intentos.fallos[c.usuario] != 1 {
			t.Fatalf("%s: el fallo no se contó", c.usuario)
		}
	}
}

func TestAdminLogin_TercerFalloBloqueaDiezMinutos(t *testing.T) {
	f := newAdminFixture(t)
	for i := 1; i <= 2; i++ {
		if _, err := loginAdmin(f, "administrador", "mala1234"); !errors.Is(err, domain.ErrUnauthorized) {
			t.Fatalf("intento %d: se esperaba 401, llegó %v", i, err)
		}
	}
	_, err := loginAdmin(f, "administrador", "mala1234")
	var de *domain.Error
	if !errors.As(err, &de) || !errors.Is(err, domain.ErrLocked) || de.RetryAfter != service.BloqueoLogin {
		t.Fatalf("tercer intento: se esperaba 429 de 10 minutos, llegó %v", err)
	}
	// Con el bloqueo vigente, ni la contraseña correcta entra.
	if _, err := loginAdmin(f, "administrador", "Admin2026"); !errors.Is(err, domain.ErrLocked) {
		t.Fatalf("se esperaba 429 con el bloqueo vigente, llegó %v", err)
	}
}

func TestAdminLogin_AciertoReiniciaElContador(t *testing.T) {
	f := newAdminFixture(t)
	f.intentos.fallos["analista"] = 2
	if _, err := loginAdmin(f, "analista", "Analista2026"); err != nil {
		t.Fatal(err)
	}
	if f.intentos.fallos["analista"] != 0 {
		t.Fatal("el contador no se reinició")
	}
}

func TestAdminLogin_CamposVacios(t *testing.T) {
	f := newAdminFixture(t)
	_, err := loginAdmin(f, " ", "")
	var de *domain.Error
	if !errors.As(err, &de) || !errors.Is(err, domain.ErrValidation) || de.Fields["usuario"] == "" || de.Fields["password"] == "" {
		t.Fatalf("se esperaba 400 con fields, llegó %v", err)
	}
	if len(f.intentos.fallos) != 0 {
		t.Fatal("un error de formato no cuenta como intento")
	}
}
