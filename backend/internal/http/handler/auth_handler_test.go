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
}

func (f *fakeRegistro) ValidarIdentidad(_ context.Context, in service.IdentidadInput) (string, error) {
	f.got = in
	return f.ticket, f.err
}

func newRouter(reg handler.RegistroUseCase) *gin.Engine {
	return apihttp.NewRouter(apihttp.Deps{
		Config:  config.Config{RequestTimeout: time.Second},
		Logger:  slog.New(slog.NewTextHandler(io.Discard, nil)),
		Health:  handler.NewHealthHandler(),
		Modules: []apihttp.RouteRegistrar{apihttp.AuthRoutes(handler.NewAuthHandler(reg))},
	})
}

func post(r *gin.Engine, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest("POST", "/api/v1/auth/identidad", strings.NewReader(body))
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
