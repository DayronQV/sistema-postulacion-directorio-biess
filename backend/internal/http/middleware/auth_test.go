package middleware_test

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/auth"
	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/domain"
	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/http/middleware"
)

const secretoPrueba = "secreto-de-prueba-con-mas-de-32-caracteres"

func init() { gin.SetMode(gin.TestMode) }

func nuevoEmisor(t *testing.T, ttl time.Duration) *auth.JWTIssuer {
	t.Helper()
	i, err := auth.NewJWTIssuer(secretoPrueba, ttl)
	if err != nil {
		t.Fatal(err)
	}
	return i
}

func tokenDe(t *testing.T, i *auth.JWTIssuer, rol domain.Rol) string {
	t.Helper()
	e, err := i.Emitir(domain.Claims{UsuarioID: "u-1", Cedula: "1712345675", Rol: rol})
	if err != nil {
		t.Fatal(err)
	}
	return e.Token
}

// router arma tres rutas de prueba: una solo con sesión, una para administradores y
// una para analistas o administradores.
func router(tokens domain.TokenIssuer) *gin.Engine {
	r := gin.New()
	r.Use(middleware.ErrorHandler(slog.New(slog.NewTextHandler(io.Discard, nil))))

	r.GET("/privada", middleware.RequireAuth(tokens), func(c *gin.Context) {
		claims, ok := middleware.ClaimsFrom(c)
		c.JSON(http.StatusOK, gin.H{"ok": ok, "usuario": claims.UsuarioID, "rol": claims.Rol})
	})
	r.GET("/admin", middleware.RequireAuth(tokens), middleware.RequireRole(domain.RolAdministrador),
		func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"ok": true}) })
	r.GET("/panel", middleware.RequireAuth(tokens), middleware.RequireRole(domain.RolAnalista, domain.RolAdministrador),
		func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"ok": true}) })
	// Mal armada a propósito: pide un rol sin haber pedido la sesión antes.
	r.GET("/sin-sesion", middleware.RequireRole(domain.RolAdministrador),
		func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"ok": true}) })
	return r
}

func pedir(r *gin.Engine, ruta, authorization string) *httptest.ResponseRecorder {
	req := httptest.NewRequest("GET", ruta, nil)
	if authorization != "" {
		req.Header.Set("Authorization", authorization)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func codigoDeError(t *testing.T, w *httptest.ResponseRecorder) string {
	t.Helper()
	var body middleware.ErrorBody
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("la respuesta no es el JSON de error de la API: %v\n%s", err, w.Body.String())
	}
	return body.Error.Code
}

func TestRequireAuthSinTokenOMalFormado(t *testing.T) {
	r := router(nuevoEmisor(t, time.Hour))
	casos := map[string]string{
		"sin cabecera":              "",
		"esquema distinto (Basic)":  "Basic dXNlcjpwYXNz",
		"Bearer sin token":          "Bearer",
		"Bearer con espacios solos": "Bearer    ",
		"token sin esquema":         "abc.def.ghi",
		"token inválido":            "Bearer no-es-un-token",
	}
	for nombre, cabecera := range casos {
		t.Run(nombre, func(t *testing.T) {
			w := pedir(r, "/privada", cabecera)
			if w.Code != 401 || codigoDeError(t, w) != middleware.CodeUnauthorized {
				t.Fatalf("respuesta = %d %s, se esperaba 401 UNAUTHORIZED", w.Code, w.Body.String())
			}
		})
	}
}

func TestRequireAuthTokenVencido(t *testing.T) {
	// Un emisor cuyos tokens vencen casi de inmediato.
	corto := nuevoEmisor(t, time.Millisecond)
	token := tokenDe(t, corto, domain.RolPostulante)
	time.Sleep(1100 * time.Millisecond) // el vencimiento del JWT se guarda en segundos

	w := pedir(router(corto), "/privada", "Bearer "+token)
	if w.Code != 401 || codigoDeError(t, w) != middleware.CodeUnauthorized {
		t.Fatalf("respuesta = %d %s, se esperaba 401 UNAUTHORIZED", w.Code, w.Body.String())
	}
}

func TestRequireAuthTokenValidoGuardaLaIdentidad(t *testing.T) {
	emisor := nuevoEmisor(t, time.Hour)
	for _, esquema := range []string{"Bearer", "bearer", "BEARER"} {
		w := pedir(router(emisor), "/privada", esquema+" "+tokenDe(t, emisor, domain.RolPostulante))
		if w.Code != 200 {
			t.Fatalf("%s: status = %d: %s", esquema, w.Code, w.Body.String())
		}
		var body struct {
			Ok      bool
			Usuario string
			Rol     string
		}
		if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
			t.Fatal(err)
		}
		if !body.Ok || body.Usuario != "u-1" || body.Rol != "postulante" {
			t.Fatalf("%s: el handler no recibió la identidad: %+v", esquema, body)
		}
	}
}

func TestRequireRole(t *testing.T) {
	emisor := nuevoEmisor(t, time.Hour)
	r := router(emisor)
	casos := []struct {
		nombre string
		ruta   string
		rol    domain.Rol
		status int
	}{
		{"postulante en ruta de administrador", "/admin", domain.RolPostulante, 403},
		{"analista en ruta de administrador", "/admin", domain.RolAnalista, 403},
		{"administrador en ruta de administrador", "/admin", domain.RolAdministrador, 200},
		{"postulante en ruta de panel", "/panel", domain.RolPostulante, 403},
		{"analista en ruta de panel (varios roles)", "/panel", domain.RolAnalista, 200},
		{"administrador en ruta de panel (varios roles)", "/panel", domain.RolAdministrador, 200},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			w := pedir(r, c.ruta, "Bearer "+tokenDe(t, emisor, c.rol))
			if w.Code != c.status {
				t.Fatalf("status = %d, se esperaba %d: %s", w.Code, c.status, w.Body.String())
			}
			if c.status == 403 && codigoDeError(t, w) != middleware.CodeForbidden {
				t.Fatalf("el 403 no usa el código FORBIDDEN: %s", w.Body.String())
			}
		})
	}
}

func TestRequireRoleSinSesionResponde401(t *testing.T) {
	w := pedir(router(nuevoEmisor(t, time.Hour)), "/sin-sesion", "")
	if w.Code != 401 || codigoDeError(t, w) != middleware.CodeUnauthorized {
		t.Fatalf("respuesta = %d %s, se esperaba 401 UNAUTHORIZED", w.Code, w.Body.String())
	}
}

func TestRequireRoleSinRolesEsUnErrorDeProgramacion(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("RequireRole() sin roles debe fallar al armar las rutas")
		}
	}()
	middleware.RequireRole()
}

func TestClaimsFromSinSesion(t *testing.T) {
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	if _, ok := middleware.ClaimsFrom(c); ok {
		t.Fatal("ClaimsFrom debe devolver false si la ruta no pasó por RequireAuth")
	}
}
