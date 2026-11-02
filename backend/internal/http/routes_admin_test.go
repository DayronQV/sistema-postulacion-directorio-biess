package http_test

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	nethttp "net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/config"
	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/domain"
	apihttp "github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/http"
	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/http/handler"
	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/service"
)

type fakeAdminLogin struct {
	sesion service.AdminSesion
	err    error
}

func (f *fakeAdminLogin) Login(context.Context, service.AdminLoginInput) (service.AdminSesion, error) {
	return f.sesion, f.err
}

// moduloDePrueba simula un módulo del panel con una ruta de lectura y una de escritura.
func moduloDePrueba(lectura, escritura *gin.RouterGroup) {
	lectura.GET("/prueba", func(c *gin.Context) { c.JSON(nethttp.StatusOK, gin.H{"ok": true}) })
	escritura.PUT("/prueba", func(c *gin.Context) { c.JSON(nethttp.StatusOK, gin.H{"ok": true}) })
}

func newAdminRouter(login *fakeAdminLogin) *gin.Engine {
	return apihttp.NewRouter(apihttp.Deps{
		Config: config.Config{RequestTimeout: time.Second},
		Logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
		Health: handler.NewHealthHandler(),
		Modules: []apihttp.RouteRegistrar{
			apihttp.AdminRoutes(handler.NewAdminAuthHandler(login), emisorPrueba, moduloDePrueba),
		},
	})
}

func pedirAdmin(r *gin.Engine, metodo, ruta, token, cuerpo string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(metodo, ruta, strings.NewReader(cuerpo))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestAdminLogin_RespondeTokenYRol(t *testing.T) {
	r := newAdminRouter(&fakeAdminLogin{sesion: service.AdminSesion{
		Token: "jwt", ExpiraEn: time.Hour, Usuario: "analista", Rol: domain.RolAnalista,
	}})
	w := pedirAdmin(r, "POST", "/api/v1/admin/auth/login", "", `{"usuario":"analista","password":"x"}`)
	if w.Code != 200 {
		t.Fatalf("respuesta = %d %s", w.Code, w.Body.String())
	}
	var got map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if got["token"] != "jwt" || got["expires_in"] != float64(3600) || got["rol"] != "analista" || got["usuario"] != "analista" {
		t.Fatalf("cuerpo = %v", got)
	}
}

func TestAdminLogin_CamposObligatorios(t *testing.T) {
	r := newAdminRouter(&fakeAdminLogin{})
	w := pedirAdmin(r, "POST", "/api/v1/admin/auth/login", "", `{"usuario":"analista"}`)
	if w.Code != 400 || decodeError(t, w).Fields["password"] == "" {
		t.Fatalf("respuesta = %d %s", w.Code, w.Body.String())
	}
}

// Criterios de la issue #36: un analista no puede escribir (403) y un postulante no entra
// al panel (403); sin sesión, 401.
func TestAdminRutas_ProtegidasPorRol(t *testing.T) {
	r := newAdminRouter(&fakeAdminLogin{})
	casos := []struct {
		nombre, metodo, ruta string
		rol                  domain.Rol
		want                 int
	}{
		{"sin token", "GET", "/api/v1/admin/prueba", "", 401},
		{"postulante no entra al panel", "GET", "/api/v1/admin/prueba", domain.RolPostulante, 403},
		{"postulante no ve la sesión", "GET", "/api/v1/admin/sesion", domain.RolPostulante, 403},
		{"analista lee", "GET", "/api/v1/admin/prueba", domain.RolAnalista, 200},
		{"analista no escribe", "PUT", "/api/v1/admin/prueba", domain.RolAnalista, 403},
		{"administrador lee", "GET", "/api/v1/admin/prueba", domain.RolAdministrador, 200},
		{"administrador escribe", "PUT", "/api/v1/admin/prueba", domain.RolAdministrador, 200},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			token := ""
			if c.rol != "" {
				token = tokenParaRol(t, c.rol)
			}
			w := pedirAdmin(r, c.metodo, c.ruta, token, "")
			if w.Code != c.want {
				t.Fatalf("respuesta = %d, se esperaba %d: %s", w.Code, c.want, w.Body.String())
			}
		})
	}
}

func TestAdminSesion_DevuelveUsuarioYRol(t *testing.T) {
	r := newAdminRouter(&fakeAdminLogin{})
	e, err := emisorPrueba.Emitir(domain.Claims{UsuarioID: "a-2", Usuario: "administrador", Rol: domain.RolAdministrador})
	if err != nil {
		t.Fatal(err)
	}
	w := pedirAdmin(r, "GET", "/api/v1/admin/sesion", e.Token, "")
	if w.Code != 200 || !strings.Contains(w.Body.String(), `"usuario":"administrador"`) || !strings.Contains(w.Body.String(), `"rol":"administrador"`) {
		t.Fatalf("respuesta = %d %s", w.Code, w.Body.String())
	}
}
