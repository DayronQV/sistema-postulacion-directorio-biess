package http_test

import (
	"context"
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

// fakeEvaluacion devuelve lo que se le configure y recuerda con qué lo llamaron.
type fakeEvaluacion struct {
	sesion   service.SesionEvaluacion
	creado   bool
	pregunta service.PreguntaEvaluacion
	err      error

	postulanteID string
	n            int

	resultado service.ResultadoFinal
	opcion    *int
	guardo    bool
}

func (f *fakeEvaluacion) GuardarRespuesta(_ context.Context, postulanteID string, n int, opcion *int) error {
	f.postulanteID, f.n, f.opcion, f.guardo = postulanteID, n, opcion, true
	return f.err
}

func (f *fakeEvaluacion) Finalizar(_ context.Context, postulanteID string) (service.ResultadoFinal, error) {
	f.postulanteID = postulanteID
	return f.resultado, f.err
}

func (f *fakeEvaluacion) Iniciar(_ context.Context, postulanteID string) (service.SesionEvaluacion, bool, error) {
	f.postulanteID = postulanteID
	return f.sesion, f.creado, f.err
}

func (f *fakeEvaluacion) Pregunta(_ context.Context, postulanteID string, n int) (service.PreguntaEvaluacion, error) {
	f.postulanteID, f.n = postulanteID, n
	return f.pregunta, f.err
}

func routerEvaluacionPostulante(f *fakeEvaluacion) *gin.Engine {
	return apihttp.NewRouter(apihttp.Deps{
		Config: config.Config{RequestTimeout: time.Second},
		Logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
		Health: handler.NewHealthHandler(),
		Modules: []apihttp.RouteRegistrar{
			apihttp.EvaluacionRoutes(handler.NewEvaluacionHandler(f), emisorPrueba),
		},
	})
}

func pedirPostulante(r *gin.Engine, metodo, ruta, token string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(metodo, ruta, nil)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestEvaluacionPostulante_SoloElPostulante(t *testing.T) {
	for _, ruta := range []struct{ metodo, ruta string }{
		{"POST", "/api/v1/evaluacion/iniciar"},
		{"GET", "/api/v1/evaluacion/pregunta/1"},
	} {
		for _, c := range []struct {
			nombre string
			rol    domain.Rol
			want   int
		}{{"sin token", "", 401}, {"analista", domain.RolAnalista, 403}, {"administrador", domain.RolAdministrador, 403}, {"postulante", domain.RolPostulante, 200}} {
			t.Run(ruta.metodo+" "+c.nombre, func(t *testing.T) {
				f := &fakeEvaluacion{}
				token := ""
				if c.rol != "" {
					token = tokenParaRol(t, c.rol)
				}
				w := pedirPostulante(routerEvaluacionPostulante(f), ruta.metodo, ruta.ruta, token)
				if w.Code != c.want {
					t.Fatalf("respuesta = %d, se esperaba %d: %s", w.Code, c.want, w.Body.String())
				}
			})
		}
	}
}

func TestEvaluacionPostulante_IniciarResponde201Si201CreaYElMismoExamenConUn200(t *testing.T) {
	inicio := time.Date(2026, 11, 20, 9, 0, 0, 0, time.FixedZone("ECT", -5*3600))
	f := &fakeEvaluacion{creado: true, sesion: service.SesionEvaluacion{
		TotalPreguntas: 10, Inicio: inicio, Cierre: inicio.Add(30 * time.Minute), TiempoRestante: 29*time.Minute + 59*time.Second + 900*time.Millisecond,
	}}
	r := routerEvaluacionPostulante(f)
	token := tokenParaRol(t, domain.RolPostulante)

	w := pedirPostulante(r, "POST", "/api/v1/evaluacion/iniciar", token)
	m := cuerpoJSON(t, w)
	if w.Code != 201 || m["total_preguntas"] != float64(10) || m["inicio"] != "2026-11-20T14:00:00Z" ||
		m["cierre"] != "2026-11-20T14:30:00Z" || m["tiempo_restante_segundos"] != float64(1799) {
		t.Fatalf("crear: %d %s", w.Code, w.Body.String())
	}
	if f.postulanteID == "" {
		t.Fatal("el servicio debe recibir el id del postulante del token")
	}

	f.creado = false
	if w := pedirPostulante(r, "POST", "/api/v1/evaluacion/iniciar", token); w.Code != 200 {
		t.Fatalf("el examen ya iniciado debe responder 200: %d", w.Code)
	}
}

func TestEvaluacionPostulante_IniciarTraduceLosErroresAlFormatoDeLaApi(t *testing.T) {
	casos := []struct {
		err  error
		code int
		tipo string
	}{
		{domain.Forbidden(service.MsgEvaluacionSinMerito), 403, "FORBIDDEN"},
		{&domain.Error{Base: domain.ErrForbidden, Message: service.MsgEvaluacionNoComienza, Fields: map[string]string{"inicia_en": "2026-11-20T14:00:00Z"}}, 403, "FORBIDDEN"},
		{domain.Conflict(service.MsgEvaluacionTerminada, nil), 409, "CONFLICT"},
		{domain.Unprocessable(service.MsgEvaluacionBancoCorto, nil), 422, "VALIDATION_ERROR"},
		{&domain.Error{Base: domain.ErrForbidden, Message: service.MsgEvaluacionPlazoVencido, Fields: map[string]string{"cierra_en": "2026-11-20T14:30:00Z"}}, 403, "FORBIDDEN"},
	}
	for _, c := range casos {
		w := pedirPostulante(routerEvaluacionPostulante(&fakeEvaluacion{err: c.err}), "POST", "/api/v1/evaluacion/iniciar", tokenParaRol(t, domain.RolPostulante))
		e := decodeError(t, w)
		if w.Code != c.code || e.Code != c.tipo || e.Message != c.err.Error() {
			t.Fatalf("%v → %d %s", c.err, w.Code, w.Body.String())
		}
	}
	// El error del horario dice cuándo empieza.
	w := pedirPostulante(routerEvaluacionPostulante(&fakeEvaluacion{err: casos[1].err}), "POST", "/api/v1/evaluacion/iniciar", tokenParaRol(t, domain.RolPostulante))
	if decodeError(t, w).Fields["inicia_en"] != "2026-11-20T14:00:00Z" {
		t.Fatalf("falta inicia_en: %s", w.Body.String())
	}
}

// Criterio de la issue #42/#43: la respuesta correcta nunca se envía al postulante.
func TestEvaluacionPostulante_LaPreguntaNuncaIncluyeLaRespuestaCorrecta(t *testing.T) {
	cierre := time.Date(2026, 11, 20, 14, 30, 0, 0, time.UTC)
	f := &fakeEvaluacion{pregunta: service.PreguntaEvaluacion{
		Numero: 3, Total: 10, Cierre: cierre, TiempoRestante: 12*time.Minute + 5*time.Second,
		Pregunta: domain.PreguntaPublica{ID: 77, Enunciado: "¿Capital de Ecuador?", Opciones: []string{"Cuenca", "Quito", "Guayaquil"}},
	}}
	w := pedirPostulante(routerEvaluacionPostulante(f), "GET", "/api/v1/evaluacion/pregunta/3", tokenParaRol(t, domain.RolPostulante))
	if w.Code != 200 || f.n != 3 {
		t.Fatalf("respuesta = %d %s (n = %d)", w.Code, w.Body.String(), f.n)
	}
	m := cuerpoJSON(t, w)
	if m["numero"] != float64(3) || m["total"] != float64(10) || m["id"] != float64(77) || m["enunciado"] != "¿Capital de Ecuador?" ||
		m["cierre"] != "2026-11-20T14:30:00Z" || m["tiempo_restante_segundos"] != float64(725) {
		t.Fatalf("cuerpo = %v", m)
	}
	if opciones, ok := m["opciones"].([]any); !ok || len(opciones) != 3 || opciones[1] != "Quito" {
		t.Fatalf("opciones = %v", m["opciones"])
	}
	cuerpo := strings.ToLower(w.Body.String())
	for _, prohibida := range []string{"correcta", "respuesta"} {
		if strings.Contains(cuerpo, prohibida) {
			t.Fatalf("la respuesta no debe contener %q: %s", prohibida, w.Body.String())
		}
	}
}

func TestEvaluacionPostulante_PreguntaConNumeroInvalidoOErrorEnElServicio(t *testing.T) {
	token := tokenParaRol(t, domain.RolPostulante)
	for _, ruta := range []string{"/api/v1/evaluacion/pregunta/abc", "/api/v1/evaluacion/pregunta/1.5", "/api/v1/evaluacion/pregunta/-"} {
		f := &fakeEvaluacion{}
		w := pedirPostulante(routerEvaluacionPostulante(f), "GET", ruta, token)
		if w.Code != 404 || decodeError(t, w).Code != "NOT_FOUND" || f.n != 0 {
			t.Fatalf("%s: %d %s", ruta, w.Code, w.Body.String())
		}
	}
	for _, c := range []struct {
		err  error
		code int
		tipo string
	}{
		{domain.NotFound(service.MsgEvaluacionNoIniciada), 404, "NOT_FOUND"},
		{domain.Conflict(service.MsgEvaluacionTerminada, nil), 409, "CONFLICT"},
	} {
		w := pedirPostulante(routerEvaluacionPostulante(&fakeEvaluacion{err: c.err}), "GET", "/api/v1/evaluacion/pregunta/1", token)
		if w.Code != c.code || decodeError(t, w).Code != c.tipo {
			t.Fatalf("%v → %d %s", c.err, w.Code, w.Body.String())
		}
	}
}

func TestEvaluacionPostulante_ElTiempoRestanteNuncaEsNegativo(t *testing.T) {
	f := &fakeEvaluacion{creado: true, sesion: service.SesionEvaluacion{TotalPreguntas: 1, TiempoRestante: -5 * time.Second}}
	w := pedirPostulante(routerEvaluacionPostulante(f), "POST", "/api/v1/evaluacion/iniciar", tokenParaRol(t, domain.RolPostulante))
	if cuerpoJSON(t, w)["tiempo_restante_segundos"] != float64(0) {
		t.Fatalf("cuerpo = %s", w.Body.String())
	}
}
