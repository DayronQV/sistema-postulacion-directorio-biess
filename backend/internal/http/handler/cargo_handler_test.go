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
)

type fakeCargo struct {
	id        string
	cargo     string
	llamadas  int
	resultado domain.ResultadoCargo
	err       error
}

func (f *fakeCargo) ValidarCargo(_ context.Context, id, cargo string) (domain.ResultadoCargo, error) {
	f.llamadas++
	f.id, f.cargo = id, cargo
	return f.resultado, f.err
}

func newCargoRouter(uc handler.CargoUseCase) *gin.Engine {
	return apihttp.NewRouter(apihttp.Deps{
		Config:  config.Config{RequestTimeout: time.Second},
		Logger:  slog.New(slog.NewTextHandler(io.Discard, nil)),
		Health:  handler.NewHealthHandler(),
		Modules: []apihttp.RouteRegistrar{apihttp.CargoRoutes(handler.NewCargoHandler(uc), fakeTokens{})},
	})
}

func sendCargo(r *gin.Engine, token, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest("POST", "/api/v1/postulacion/cargo/validar", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestCargoValidar_Elegible(t *testing.T) {
	uc := &fakeCargo{resultado: domain.ResultadoCargo{Cargo: "jubilados", Elegible: true, Mensaje: "Postulante acredita jubilación"}}

	w := sendCargo(newCargoRouter(uc), "postulante", `{"cargo":"jubilados"}`)
	if w.Code != 200 {
		t.Fatalf("respuesta = %d %s", w.Code, w.Body.String())
	}
	var got struct {
		Cargo    string `json:"cargo"`
		Elegible bool   `json:"elegible"`
		Mensaje  string `json:"mensaje"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if got.Cargo != "jubilados" || !got.Elegible || got.Mensaje != "Postulante acredita jubilación" {
		t.Fatalf("cuerpo inesperado: %s", w.Body.String())
	}
	if uc.id != "p-1" || uc.cargo != "jubilados" {
		t.Fatalf("el servicio recibió id=%q cargo=%q", uc.id, uc.cargo)
	}
}

func TestCargoValidar_NoElegibleResponde422ConElMensajeLiteral(t *testing.T) {
	mensaje := "Su postulación para el cargo de Miembro Principal y Alterno en Representación del Sector Empleador ante el Directorio del BIESS, NO puede continuar, por cuanto no es empleador"
	uc := &fakeCargo{err: domain.Unprocessable(mensaje, nil)}

	w := sendCargo(newCargoRouter(uc), "postulante", `{"cargo":"empleador"}`)
	if w.Code != 422 {
		t.Fatalf("respuesta = %d %s", w.Code, w.Body.String())
	}
	var got struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if got.Error.Code != "VALIDATION_ERROR" || got.Error.Message != mensaje {
		t.Fatalf("error inesperado: %s", w.Body.String())
	}
}

func TestCargoValidar_ErroresDelServicio(t *testing.T) {
	casos := map[string]struct {
		err    error
		status int
		code   string
	}{
		"cargo no válido":      {domain.Validation("Datos inválidos", map[string]string{"cargo": "Cargo no válido"}), 400, "VALIDATION_ERROR"},
		"ya no es borrador":    {domain.Conflict("La postulación ya no se puede modificar", nil), 409, "CONFLICT"},
		"error interno":        {context.DeadlineExceeded, 500, "INTERNAL_ERROR"},
		"postulante no existe": {domain.NotFound("Postulante no encontrado"), 404, "NOT_FOUND"},
	}
	for nombre, c := range casos {
		t.Run(nombre, func(t *testing.T) {
			w := sendCargo(newCargoRouter(&fakeCargo{err: c.err}), "postulante", `{"cargo":"afiliados"}`)
			if w.Code != c.status || !strings.Contains(w.Body.String(), `"code":"`+c.code+`"`) {
				t.Fatalf("respuesta = %d %s", w.Code, w.Body.String())
			}
		})
	}
}

func TestCargoValidar_CuerpoInvalidoResponde400SinLlamarAlServicio(t *testing.T) {
	for nombre, body := range map[string]string{
		"sin cargo":        `{}`,
		"cargo vacío":      `{"cargo":""}`,
		"JSON mal formado": `{"cargo":`,
		"tipo incorrecto":  `{"cargo":1}`,
	} {
		t.Run(nombre, func(t *testing.T) {
			uc := &fakeCargo{}
			w := sendCargo(newCargoRouter(uc), "postulante", body)
			if w.Code != 400 || !strings.Contains(w.Body.String(), `"code":"VALIDATION_ERROR"`) || uc.llamadas != 0 {
				t.Fatalf("respuesta = %d %s, llamadas = %d", w.Code, w.Body.String(), uc.llamadas)
			}
		})
	}
}

func TestCargoValidar_ExigeSesionDePostulante(t *testing.T) {
	uc := &fakeCargo{}
	r := newCargoRouter(uc)

	if w := sendCargo(r, "", `{"cargo":"afiliados"}`); w.Code != 401 {
		t.Fatalf("sin token = %d", w.Code)
	}
	if w := sendCargo(r, "analista", `{"cargo":"afiliados"}`); w.Code != 403 {
		t.Fatalf("con rol analista = %d", w.Code)
	}
	if uc.llamadas != 0 {
		t.Fatalf("el servicio no debía llamarse")
	}
}
