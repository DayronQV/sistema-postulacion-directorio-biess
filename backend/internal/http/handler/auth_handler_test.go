package handler_test

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
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

func init() { gin.SetMode(gin.TestMode) }

type fakeRegistro struct {
	got    service.IdentidadInput
	ticket string
	err    error

	gotRegistro service.RegistroInput
	id          string
}

func (f *fakeRegistro) ValidarIdentidad(_ context.Context, in service.IdentidadInput) (string, error) {
	f.got = in
	return f.ticket, f.err
}

func (f *fakeRegistro) Registrar(_ context.Context, in service.RegistroInput) (string, error) {
	f.gotRegistro = in
	return f.id, f.err
}

type fakeLogin struct {
	got service.LoginInput
	res service.LoginResult
	err error
}

func (f *fakeLogin) Login(_ context.Context, in service.LoginInput) (service.LoginResult, error) {
	f.got = in
	return f.res, f.err
}

func newRouter(reg handler.RegistroUseCase) *gin.Engine {
	return newRouterCon(reg, &fakeLogin{})
}

func newRouterCon(reg handler.RegistroUseCase, login handler.LoginUseCase) *gin.Engine {
	return apihttp.NewRouter(apihttp.Deps{
		Config:  config.Config{RequestTimeout: time.Second},
		Logger:  slog.New(slog.NewTextHandler(io.Discard, nil)),
		Health:  handler.NewHealthHandler(),
		Modules: []apihttp.RouteRegistrar{apihttp.AuthRoutes(handler.NewAuthHandler(reg, login))},
	})
}

func post(r *gin.Engine, body string) *httptest.ResponseRecorder {
	return postTo(r, "/api/v1/auth/identidad", body)
}

func postTo(r *gin.Engine, path, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest("POST", path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestIdentidad_OK(t *testing.T) {
	reg := &fakeRegistro{ticket: "abc"}
	w := post(newRouter(reg), `{"cedula":"0924681356","codigo_dactilar":"E5566V7788"}`)
	if w.Code != 200 || strings.TrimSpace(w.Body.String()) != `{"ticket":"abc"}` {
		t.Fatalf("respuesta = %d %s", w.Code, w.Body.String())
	}
	if reg.got.Cedula != "0924681356" || reg.got.CodigoDactilar != "E5566V7788" {
		t.Fatalf("el handler no pasó los datos al servicio: %+v", reg.got)
	}
}

func TestIdentidad_Locked(t *testing.T) {
	reg := &fakeRegistro{err: domain.Locked(service.MsgRegistroBloqueado, 600*time.Second)}
	w := post(newRouter(reg), `{"cedula":"0924681356","fecha_expedicion":"05/12/2018"}`)
	if w.Code != 429 || w.Header().Get("Retry-After") != "600" {
		t.Fatalf("respuesta = %d, Retry-After %q", w.Code, w.Header().Get("Retry-After"))
	}
	var body struct {
		Error struct {
			Code              string `json:"code"`
			Message           string `json:"message"`
			RetryAfterSeconds int    `json:"retry_after_seconds"`
		} `json:"error"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Error.Code != "LOCKED" || body.Error.Message != service.MsgRegistroBloqueado || body.Error.RetryAfterSeconds != 600 {
		t.Fatalf("cuerpo = %+v", body.Error)
	}
}

func TestIdentidad_SinCedula(t *testing.T) {
	reg := &fakeRegistro{}
	w := post(newRouter(reg), `{"codigo_dactilar":"E5566V7788"}`)
	if w.Code != 400 || !strings.Contains(w.Body.String(), `"cedula"`) {
		t.Fatalf("respuesta = %d %s", w.Code, w.Body.String())
	}
}

func TestRegistro_Created(t *testing.T) {
	reg := &fakeRegistro{id: "uuid-1"}
	w := postTo(newRouter(reg), "/api/v1/auth/registro",
		`{"ticket":"t1","email":"ana@correo.com","password":"clave1234"}`)
	if w.Code != 201 || strings.TrimSpace(w.Body.String()) != `{"id":"uuid-1"}` {
		t.Fatalf("respuesta = %d %s", w.Code, w.Body.String())
	}
	if reg.gotRegistro != (service.RegistroInput{Ticket: "t1", Email: "ana@correo.com", Password: "clave1234"}) {
		t.Fatalf("el handler no pasó los datos al servicio: %+v", reg.gotRegistro)
	}
}

func TestRegistro_Conflict(t *testing.T) {
	reg := &fakeRegistro{err: domain.ConflictoEmail()}
	w := postTo(newRouter(reg), "/api/v1/auth/registro",
		`{"ticket":"t1","email":"ana@correo.com","password":"clave1234"}`)
	if w.Code != 409 || !strings.Contains(w.Body.String(), `"email"`) {
		t.Fatalf("respuesta = %d %s", w.Code, w.Body.String())
	}
}

func TestRegistro_CamposObligatorios(t *testing.T) {
	w := postTo(newRouter(&fakeRegistro{}), "/api/v1/auth/registro", `{"email":"ana@correo.com"}`)
	if w.Code != 400 || !strings.Contains(w.Body.String(), `"ticket"`) || !strings.Contains(w.Body.String(), `"password"`) {
		t.Fatalf("respuesta = %d %s", w.Code, w.Body.String())
	}
}

func TestLogin_OK(t *testing.T) {
	login := &fakeLogin{res: service.LoginResult{Token: "jwt", ExpiraEn: time.Hour, PrimerIngreso: true}}
	w := postTo(newRouterCon(&fakeRegistro{}, login), "/api/v1/auth/login",
		`{"cedula":"0924681356","password":"clave1234"}`)
	if w.Code != 200 || strings.TrimSpace(w.Body.String()) != `{"token":"jwt","expires_in":3600,"primer_ingreso":true}` {
		t.Fatalf("respuesta = %d %s", w.Code, w.Body.String())
	}
	if login.got != (service.LoginInput{Cedula: "0924681356", Password: "clave1234"}) {
		t.Fatalf("el handler no pasó las credenciales: %+v", login.got)
	}
}

func TestLogin_Incorrecto(t *testing.T) {
	login := &fakeLogin{err: domain.Unauthorized(service.MsgCredencialesIncorrectas)}
	w := postTo(newRouterCon(&fakeRegistro{}, login), "/api/v1/auth/login",
		`{"cedula":"0924681356","password":"mala1234"}`)
	if w.Code != 401 || !strings.Contains(w.Body.String(), service.MsgCredencialesIncorrectas) {
		t.Fatalf("respuesta = %d %s", w.Code, w.Body.String())
	}
}

func TestLogin_Bloqueado(t *testing.T) {
	login := &fakeLogin{err: domain.Locked(service.MsgLoginBloqueado, 600*time.Second)}
	w := postTo(newRouterCon(&fakeRegistro{}, login), "/api/v1/auth/login",
		`{"cedula":"0924681356","password":"clave1234"}`)
	if w.Code != 429 || w.Header().Get("Retry-After") != "600" || !strings.Contains(w.Body.String(), `"retry_after_seconds":600`) {
		t.Fatalf("respuesta = %d %s", w.Code, w.Body.String())
	}
}

func TestLogin_CamposObligatorios(t *testing.T) {
	w := postTo(newRouter(&fakeRegistro{}), "/api/v1/auth/login", `{"cedula":"0924681356"}`)
	if w.Code != 400 || !strings.Contains(w.Body.String(), `"password"`) {
		t.Fatalf("respuesta = %d %s", w.Code, w.Body.String())
	}
}
