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

func pedirConCuerpo(r *gin.Engine, metodo, ruta, token, cuerpo string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(metodo, ruta, strings.NewReader(cuerpo))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestRespuesta_GuardaYResponde204(t *testing.T) {
	f := &fakeEvaluacion{}
	w := pedirConCuerpo(routerEvaluacionPostulante(f), "PUT", "/api/v1/evaluacion/respuesta/2", tokenParaRol(t, domain.RolPostulante), `{"opcion":0}`)
	if w.Code != 204 || w.Body.Len() != 0 {
		t.Fatalf("respuesta = %d %s", w.Code, w.Body.String())
	}
	if !f.guardo || f.n != 2 || f.opcion == nil || *f.opcion != 0 || f.postulanteID == "" {
		t.Fatalf("el servicio debe recibir n = 2 y opción 0: n = %d, opción = %v", f.n, f.opcion)
	}
	// Sin la opción, el servicio recibe nil y decide (400).
	f = &fakeEvaluacion{}
	pedirConCuerpo(routerEvaluacionPostulante(f), "PUT", "/api/v1/evaluacion/respuesta/2", tokenParaRol(t, domain.RolPostulante), `{}`)
	if !f.guardo || f.opcion != nil {
		t.Fatalf("sin opción debe llegar nil: %v", f.opcion)
	}
}

func TestRespuesta_Errores(t *testing.T) {
	token := tokenParaRol(t, domain.RolPostulante)
	casos := []struct {
		nombre, ruta, cuerpo string
		err                  error
		code                 int
		tipo                 string
	}{
		{"pregunta no numérica", "/api/v1/evaluacion/respuesta/abc", `{"opcion":0}`, nil, 404, "NOT_FOUND"},
		{"opción que no es número", "/api/v1/evaluacion/respuesta/1", `{"opcion":"a"}`, nil, 400, "VALIDATION_ERROR"},
		{"JSON roto", "/api/v1/evaluacion/respuesta/1", `{"opcion":`, nil, 400, "VALIDATION_ERROR"},
		{"opción fuera de rango", "/api/v1/evaluacion/respuesta/1", `{"opcion":9}`,
			domain.Validation("Datos inválidos", map[string]string{"opcion": "x"}), 400, "VALIDATION_ERROR"},
		{"examen terminado", "/api/v1/evaluacion/respuesta/1", `{"opcion":0}`,
			domain.Conflict(service.MsgEvaluacionTerminada, nil), 409, "CONFLICT"},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			w := pedirConCuerpo(routerEvaluacionPostulante(&fakeEvaluacion{err: c.err}), "PUT", c.ruta, token, c.cuerpo)
			if w.Code != c.code || decodeError(t, w).Code != c.tipo {
				t.Fatalf("respuesta = %d %s", w.Code, w.Body.String())
			}
		})
	}
}

func TestFinalizar_DevuelveLaNota(t *testing.T) {
	fin := time.Date(2026, 11, 20, 9, 40, 0, 0, time.FixedZone("ECT", -5*3600))
	f := &fakeEvaluacion{resultado: service.ResultadoFinal{Nota: 66.67, Aciertos: 2, Total: 3, FinalizadoEn: fin}}
	w := pedirConCuerpo(routerEvaluacionPostulante(f), "POST", "/api/v1/evaluacion/finalizar", tokenParaRol(t, domain.RolPostulante), "")
	m := cuerpoJSON(t, w)
	if w.Code != 200 || m["nota"] != 66.67 || m["aciertos"] != float64(2) || m["total"] != float64(3) || m["finalizado_at"] != "2026-11-20T14:40:00Z" {
		t.Fatalf("respuesta = %d %s", w.Code, w.Body.String())
	}
	w = pedirConCuerpo(routerEvaluacionPostulante(&fakeEvaluacion{err: domain.NotFound(service.MsgEvaluacionNoIniciada)}), "POST", "/api/v1/evaluacion/finalizar", tokenParaRol(t, domain.RolPostulante), "")
	if w.Code != 404 || decodeError(t, w).Code != "NOT_FOUND" {
		t.Fatalf("sin examen: %d %s", w.Code, w.Body.String())
	}
}

func TestPregunta_TraeLaOpcionElegida(t *testing.T) {
	elegida := 2
	f := &fakeEvaluacion{pregunta: service.PreguntaEvaluacion{
		Numero: 1, Total: 3, Respuesta: &elegida,
		Pregunta: domain.PreguntaPublica{ID: 5, Enunciado: "x", Opciones: []string{"a", "b", "c"}},
	}}
	w := pedirPostulante(routerEvaluacionPostulante(f), "GET", "/api/v1/evaluacion/pregunta/1", tokenParaRol(t, domain.RolPostulante))
	if m := cuerpoJSON(t, w); w.Code != 200 || m["opcion_elegida"] != float64(2) {
		t.Fatalf("respuesta = %d %s", w.Code, w.Body.String())
	}
}

func TestRespuestaYFinalizar_SoloElPostulante(t *testing.T) {
	for _, ruta := range []struct{ metodo, ruta string }{
		{"PUT", "/api/v1/evaluacion/respuesta/1"},
		{"POST", "/api/v1/evaluacion/finalizar"},
	} {
		for _, c := range []struct {
			rol  domain.Rol
			want int
		}{{"", 401}, {domain.RolAnalista, 403}, {domain.RolAdministrador, 403}} {
			token := ""
			if c.rol != "" {
				token = tokenParaRol(t, c.rol)
			}
			if w := pedirConCuerpo(routerEvaluacionPostulante(&fakeEvaluacion{}), ruta.metodo, ruta.ruta, token, `{"opcion":0}`); w.Code != c.want {
				t.Fatalf("%s %s con rol %q: %d, se esperaba %d", ruta.metodo, ruta.ruta, c.rol, w.Code, c.want)
			}
		}
	}
}

// --- Resultados para el administrador ---

type fakeResultados struct {
	resultados []domain.ResultadoEvaluacion
	reporte    service.ReportePDF
	err        error
	id         string
}

func (f *fakeResultados) Resultados(context.Context) ([]domain.ResultadoEvaluacion, error) {
	return f.resultados, f.err
}

func (f *fakeResultados) Reporte(_ context.Context, id string) (service.ReportePDF, error) {
	f.id = id
	return f.reporte, f.err
}

func routerResultados(f *fakeResultados) *gin.Engine {
	return apihttp.NewRouter(apihttp.Deps{
		Config: config.Config{RequestTimeout: time.Second},
		Logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
		Health: handler.NewHealthHandler(),
		Modules: []apihttp.RouteRegistrar{
			apihttp.AdminRoutes(handler.NewAdminAuthHandler(&fakeAdminLogin{}), emisorPrueba,
				apihttp.EvaluacionResultadosModulo(handler.NewEvaluacionResultadosHandler(f))),
		},
	})
}

func TestResultados_SoloElAdministrador(t *testing.T) {
	for _, ruta := range []string{"/api/v1/admin/evaluacion/resultados", "/api/v1/admin/evaluacion/resultados/x/reporte"} {
		for _, c := range []struct {
			rol  domain.Rol
			want int
		}{{"", 401}, {domain.RolPostulante, 403}, {domain.RolAnalista, 403}} {
			token := ""
			if c.rol != "" {
				token = tokenParaRol(t, c.rol)
			}
			if w := pedirPostulante(routerResultados(&fakeResultados{}), "GET", ruta, token); w.Code != c.want {
				t.Fatalf("%s con rol %q: %d, se esperaba %d", ruta, c.rol, w.Code, c.want)
			}
		}
	}
}

func TestResultados_ListaConEstadoNotaYAciertos(t *testing.T) {
	inicio := time.Date(2026, 11, 20, 14, 0, 0, 0, time.UTC)
	fin, nota := inicio.Add(40*time.Minute), 75.0
	f := &fakeResultados{resultados: []domain.ResultadoEvaluacion{
		{IntentoID: "a", Codigo: "BIESS-1", Nombres: "Ana", Apellidos: "Núñez", Cedula: "1712345675", Cargo: "afiliados",
			Inicio: inicio, Cierre: inicio.Add(time.Hour), FinalizadoEn: &fin, Nota: &nota, TotalPreguntas: 8},
		{IntentoID: "b", Codigo: "BIESS-2", Nombres: "Luis", Apellidos: "Pérez", Cedula: "0912345675",
			Inicio: inicio, Cierre: inicio.Add(time.Hour), TotalPreguntas: 8},
	}}
	w := pedirPostulante(routerResultados(f), "GET", "/api/v1/admin/evaluacion/resultados", tokenParaRol(t, domain.RolAdministrador))
	items, _ := cuerpoJSON(t, w)["items"].([]any)
	if w.Code != 200 || len(items) != 2 {
		t.Fatalf("respuesta = %d %s", w.Code, w.Body.String())
	}
	a, b := items[0].(map[string]any), items[1].(map[string]any)
	if a["estado"] != "finalizado" || a["nota"] != float64(75) || a["aciertos"] != float64(6) || a["finalizado_at"] != "2026-11-20T14:40:00Z" {
		t.Fatalf("finalizado = %v", a)
	}
	if b["estado"] != "en_curso" || b["nota"] != nil || b["aciertos"] != nil || b["finalizado_at"] != nil {
		t.Fatalf("en curso = %v", b)
	}
}

func TestResultados_DescargaElReportePDF(t *testing.T) {
	f := &fakeResultados{reporte: service.ReportePDF{Nombre: "reporte-evaluacion-BIESS-1.pdf", Contenido: []byte("%PDF-1.3 contenido")}}
	w := pedirPostulante(routerResultados(f), "GET", "/api/v1/admin/evaluacion/resultados/abc-123/reporte", tokenParaRol(t, domain.RolAdministrador))
	if w.Code != 200 || w.Header().Get("Content-Type") != "application/pdf" || !strings.HasPrefix(w.Body.String(), "%PDF") {
		t.Fatalf("respuesta = %d %q %q", w.Code, w.Header().Get("Content-Type"), w.Body.String())
	}
	if disp := w.Header().Get("Content-Disposition"); !strings.Contains(disp, "attachment") || !strings.Contains(disp, "reporte-evaluacion-BIESS-1.pdf") || f.id != "abc-123" {
		t.Fatalf("Content-Disposition = %q, id = %q", disp, f.id)
	}
	for _, c := range []struct {
		err  error
		code int
	}{{domain.NotFound(service.MsgResultadoNoExiste), 404}, {domain.Conflict(service.MsgReporteEnCurso, nil), 409}} {
		w := pedirPostulante(routerResultados(&fakeResultados{err: c.err}), "GET", "/api/v1/admin/evaluacion/resultados/x/reporte", tokenParaRol(t, domain.RolAdministrador))
		if w.Code != c.code || decodeError(t, w).Message != c.err.Error() {
			t.Fatalf("%v → %d %s", c.err, w.Code, w.Body.String())
		}
	}
}
