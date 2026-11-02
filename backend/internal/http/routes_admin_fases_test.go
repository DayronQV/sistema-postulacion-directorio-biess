package http_test

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/config"
	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/domain"
	apihttp "github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/http"
	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/http/handler"
	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/service"
)

type fakeFasesUseCase struct {
	id      int
	cambios service.CambiosFase
	err     error
}

func (f *fakeFasesUseCase) Listar(context.Context) ([]domain.Fase, error) {
	return []domain.Fase{
		{ID: 1, Nombre: domain.FasePostulacion, Orden: 1, Publicada: true, HabilitaFormulario: true},
		{ID: 2, Nombre: domain.FaseMeritos, Orden: 2},
	}, nil
}

func (f *fakeFasesUseCase) Actualizar(_ context.Context, id int, c service.CambiosFase) (domain.Fase, error) {
	f.id, f.cambios = id, c
	if f.err != nil {
		return domain.Fase{}, f.err
	}
	return domain.Fase{ID: id, Nombre: domain.FasePostulacion, Orden: 1, Publicada: c.Publicada,
		HabilitaFormulario: c.HabilitaFormulario, FechaInicio: c.FechaInicio, FechaFin: c.FechaFin}, nil
}

func newFasesRouter(f *fakeFasesUseCase) *gin.Engine {
	h := handler.NewAdminFasesHandler(f)
	return apihttp.NewRouter(apihttp.Deps{
		Config: config.Config{RequestTimeout: time.Second},
		Logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
		Health: handler.NewHealthHandler(),
		Modules: []apihttp.RouteRegistrar{
			apihttp.AdminRoutes(handler.NewAdminAuthHandler(&fakeAdminLogin{}), emisorPrueba, h.Rutas),
		},
	})
}

const cuerpoFase = `{"publicada":true,"habilita_formulario":true,"fecha_inicio":"2026-11-01T00:00:00-05:00","fecha_fin":null}`

// Criterio de la issue #37: el analista consulta las fases; solo el administrador las cambia.
func TestFases_RutasPorRol(t *testing.T) {
	casos := []struct {
		nombre         string
		rol            domain.Rol
		listar, editar int
	}{
		{"sin token", "", 401, 401},
		{"postulante", domain.RolPostulante, 403, 403},
		{"analista", domain.RolAnalista, 200, 403},
		{"administrador", domain.RolAdministrador, 200, 200},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			token := ""
			if c.rol != "" {
				token = tokenParaRol(t, c.rol)
			}
			r := newFasesRouter(&fakeFasesUseCase{})
			if w := pedirAdmin(r, "GET", "/api/v1/admin/fases", token, ""); w.Code != c.listar {
				t.Fatalf("GET = %d, se esperaba %d: %s", w.Code, c.listar, w.Body.String())
			}
			if w := pedirAdmin(r, "PUT", "/api/v1/admin/fases/1", token, cuerpoFase); w.Code != c.editar {
				t.Fatalf("PUT = %d, se esperaba %d: %s", w.Code, c.editar, w.Body.String())
			}
		})
	}
}

func TestFases_ListarDevuelveElContrato(t *testing.T) {
	w := pedirAdmin(newFasesRouter(&fakeFasesUseCase{}), "GET", "/api/v1/admin/fases",
		tokenParaRol(t, domain.RolAnalista), "")
	var got []map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil || len(got) != 2 {
		t.Fatalf("cuerpo = %s", w.Body.String())
	}
	for _, campo := range []string{"id", "nombre", "orden", "publicada", "habilita_formulario", "fecha_inicio", "fecha_fin"} {
		if _, ok := got[0][campo]; !ok {
			t.Fatalf("falta %q en %v", campo, got[0])
		}
	}
}

func TestFases_ActualizarPasaLosCambios(t *testing.T) {
	f := &fakeFasesUseCase{}
	w := pedirAdmin(newFasesRouter(f), "PUT", "/api/v1/admin/fases/1",
		tokenParaRol(t, domain.RolAdministrador), cuerpoFase)
	if w.Code != 200 {
		t.Fatalf("respuesta = %d %s", w.Code, w.Body.String())
	}
	inicio := time.Date(2026, 11, 1, 5, 0, 0, 0, time.UTC)
	if f.id != 1 || !f.cambios.Publicada || !f.cambios.HabilitaFormulario ||
		f.cambios.FechaInicio == nil || !f.cambios.FechaInicio.Equal(inicio) || f.cambios.FechaFin != nil {
		t.Fatalf("id = %d, cambios = %+v", f.id, f.cambios)
	}
}

func TestFases_ErroresConElFormatoDeLaApi(t *testing.T) {
	admin := tokenParaRol(t, domain.RolAdministrador)
	casos := []struct {
		nombre, ruta, cuerpo string
		err                  error
		code                 int
		tipo                 string
	}{
		{"id no numérico", "/api/v1/admin/fases/abc", cuerpoFase, nil, 404, "NOT_FOUND"},
		{"fase inexistente", "/api/v1/admin/fases/9", cuerpoFase, domain.NotFound("La fase no existe"), 404, "NOT_FOUND"},
		{"regla del servicio", "/api/v1/admin/fases/2", cuerpoFase,
			domain.Validation("Datos inválidos", map[string]string{"habilita_formulario": "x"}), 400, "VALIDATION_ERROR"},
		{"fecha mal formada", "/api/v1/admin/fases/1", `{"fecha_inicio":"01/11/2026"}`, nil, 400, "VALIDATION_ERROR"},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			w := pedirAdmin(newFasesRouter(&fakeFasesUseCase{err: c.err}), "PUT", c.ruta, admin, c.cuerpo)
			if w.Code != c.code || decodeError(t, w).Code != c.tipo {
				t.Fatalf("respuesta = %d %s", w.Code, w.Body.String())
			}
		})
	}
}
