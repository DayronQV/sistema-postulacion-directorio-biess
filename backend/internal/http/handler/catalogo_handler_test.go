package handler_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/config"
	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/domain"
	apihttp "github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/http"
	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/http/handler"
)

type fakeCatalogoUseCase struct {
	items []domain.ItemCatalogo
	err   error
	id    int // último id recibido
}

func (f *fakeCatalogoUseCase) Paises(context.Context) ([]domain.ItemCatalogo, error) {
	return f.items, f.err
}

func (f *fakeCatalogoUseCase) Provincias(_ context.Context, paisID int) ([]domain.ItemCatalogo, error) {
	f.id = paisID
	return f.items, f.err
}

func (f *fakeCatalogoUseCase) Cantones(_ context.Context, provinciaID int) ([]domain.ItemCatalogo, error) {
	f.id = provinciaID
	return f.items, f.err
}

func get(uc handler.CatalogoUseCase, path string) *httptest.ResponseRecorder {
	r := apihttp.NewRouter(apihttp.Deps{
		Config:  config.Config{RequestTimeout: time.Second},
		Logger:  slog.New(slog.NewTextHandler(io.Discard, nil)),
		Health:  handler.NewHealthHandler(),
		Modules: []apihttp.RouteRegistrar{apihttp.CatalogoRoutes(handler.NewCatalogoHandler(uc))},
	})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", path, nil))
	return w
}

func TestCatalogos_ListasConElFormatoDeLaApi(t *testing.T) {
	uc := &fakeCatalogoUseCase{items: []domain.ItemCatalogo{{ID: 60, Nombre: "Ecuador"}, {ID: 1, Nombre: "Afganistán"}}}
	want := `[{"id":60,"nombre":"Ecuador"},{"id":1,"nombre":"Afganistán"}]`

	for _, path := range []string{
		"/api/v1/catalogos/paises",
		"/api/v1/catalogos/paises/60/provincias",
		"/api/v1/catalogos/provincias/17/cantones",
	} {
		w := get(uc, path)
		if w.Code != 200 || strings.TrimSpace(w.Body.String()) != want {
			t.Errorf("GET %s = %d %s", path, w.Code, w.Body.String())
		}
	}
}

func TestCatalogos_PasaElIdDeLaRutaAlServicio(t *testing.T) {
	uc := &fakeCatalogoUseCase{}

	get(uc, "/api/v1/catalogos/paises/60/provincias")
	if uc.id != 60 {
		t.Fatalf("paisID = %d; se esperaba 60", uc.id)
	}
	get(uc, "/api/v1/catalogos/provincias/17/cantones")
	if uc.id != 17 {
		t.Fatalf("provinciaID = %d; se esperaba 17", uc.id)
	}
}

func TestCatalogos_ListaVaciaEsArregloNoNull(t *testing.T) {
	w := get(&fakeCatalogoUseCase{items: nil}, "/api/v1/catalogos/paises/1/provincias")
	if w.Code != 200 || strings.TrimSpace(w.Body.String()) != "[]" {
		t.Fatalf("respuesta = %d %s", w.Code, w.Body.String())
	}
}

func TestCatalogos_IdNoNumerico_400(t *testing.T) {
	for _, path := range []string{
		"/api/v1/catalogos/paises/abc/provincias",
		"/api/v1/catalogos/provincias/1.5/cantones",
		"/api/v1/catalogos/paises/99999999999/provincias", // fuera del rango de un INTEGER
	} {
		uc := &fakeCatalogoUseCase{}
		w := get(uc, path)
		if w.Code != 400 || !strings.Contains(w.Body.String(), `"code":"VALIDATION_ERROR"`) || !strings.Contains(w.Body.String(), `"fields":{"id"`) {
			t.Errorf("GET %s = %d %s", path, w.Code, w.Body.String())
		}
		if uc.id != 0 {
			t.Errorf("GET %s llegó al servicio con id %d", path, uc.id)
		}
	}
}

func TestCatalogos_IdInexistente_404ConFormatoDeError(t *testing.T) {
	uc := &fakeCatalogoUseCase{err: domain.NotFound("El país no existe")}

	w := get(uc, "/api/v1/catalogos/paises/999/provincias")
	if w.Code != 404 {
		t.Fatalf("estado = %d; se esperaba 404", w.Code)
	}
	want := `{"error":{"code":"NOT_FOUND","message":"El país no existe"}}`
	if strings.TrimSpace(w.Body.String()) != want {
		t.Fatalf("cuerpo = %s", w.Body.String())
	}
}

func TestCatalogos_ErrorInterno_500SinDetalles(t *testing.T) {
	uc := &fakeCatalogoUseCase{err: errors.New("pq: password authentication failed")}

	w := get(uc, "/api/v1/catalogos/paises")
	if w.Code != 500 || strings.Contains(w.Body.String(), "password") || !strings.Contains(w.Body.String(), `"code":"INTERNAL_ERROR"`) {
		t.Fatalf("respuesta = %d %s", w.Code, w.Body.String())
	}
}
