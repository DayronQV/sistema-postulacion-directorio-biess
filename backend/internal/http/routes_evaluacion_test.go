package http_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
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

// bancoEnMemoria imita el banco de preguntas para probar las rutas con el servicio real.
type bancoEnMemoria struct {
	preguntas []domain.Pregunta
	sigID     int
	cfg       *domain.ConfigEvaluacion
	intentos  bool
}

func (b *bancoEnMemoria) CrearPreguntas(_ context.Context, ps []domain.Pregunta) ([]domain.Pregunta, error) {
	for i := range ps {
		b.sigID++
		ps[i].ID = b.sigID
		b.preguntas = append(b.preguntas, ps[i])
	}
	return ps, nil
}

func (b *bancoEnMemoria) ListarPreguntas(_ context.Context, f domain.FiltroPreguntas) ([]domain.Pregunta, int, error) {
	var coinciden []domain.Pregunta
	for _, p := range b.preguntas {
		if f.Buscar == "" || strings.Contains(strings.ToLower(p.Enunciado), strings.ToLower(f.Buscar)) {
			coinciden = append(coinciden, p)
		}
	}
	desde := min((f.Pagina-1)*f.PorPagina, len(coinciden))
	hasta := min(desde+f.PorPagina, len(coinciden))
	return append([]domain.Pregunta{}, coinciden[desde:hasta]...), len(coinciden), nil
}

func (b *bancoEnMemoria) ObtenerPregunta(_ context.Context, id int) (domain.Pregunta, bool, error) {
	for _, p := range b.preguntas {
		if p.ID == id {
			return p, true, nil
		}
	}
	return domain.Pregunta{}, false, nil
}

func (b *bancoEnMemoria) ActualizarPregunta(_ context.Context, p domain.Pregunta) (bool, error) {
	for i := range b.preguntas {
		if b.preguntas[i].ID == p.ID {
			b.preguntas[i] = p
			return true, nil
		}
	}
	return false, nil
}

func (b *bancoEnMemoria) EliminarPregunta(_ context.Context, id int) (bool, error) {
	for i := range b.preguntas {
		if b.preguntas[i].ID == id {
			b.preguntas = append(b.preguntas[:i], b.preguntas[i+1:]...)
			return true, nil
		}
	}
	return false, nil
}

func (b *bancoEnMemoria) ContarPreguntas(context.Context) (int, error) { return len(b.preguntas), nil }
func (b *bancoEnMemoria) HayIntentos(context.Context) (bool, error)    { return b.intentos, nil }

func (b *bancoEnMemoria) ObtenerConfig(context.Context) (domain.ConfigEvaluacion, bool, error) {
	if b.cfg == nil {
		return domain.ConfigEvaluacion{}, false, nil
	}
	return *b.cfg, true, nil
}

func (b *bancoEnMemoria) GuardarConfig(_ context.Context, c domain.ConfigEvaluacion) error {
	b.cfg = &c
	return nil
}

func routerEvaluacion(b *bancoEnMemoria) *gin.Engine {
	h := handler.NewEvaluacionAdminHandler(service.NewEvaluacionAdminService(b))
	return apihttp.NewRouter(apihttp.Deps{
		Config: config.Config{RequestTimeout: time.Second},
		Logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
		Health: handler.NewHealthHandler(),
		Modules: []apihttp.RouteRegistrar{
			apihttp.AdminRoutes(handler.NewAdminAuthHandler(&fakeAdminLogin{}), emisorPrueba, apihttp.EvaluacionAdminModulo(h)),
		},
	})
}

const rutaEval = "/api/v1/admin/evaluacion"

func pedirEval(r *gin.Engine, metodo, ruta, tipo, cuerpo string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(metodo, ruta, strings.NewReader(cuerpo))
	if tipo != "" {
		req.Header.Set("Content-Type", tipo)
	}
	return servirEval(r, req)
}

var tokenAdministrador string

func servirEval(r *gin.Engine, req *nethttp.Request) *httptest.ResponseRecorder {
	if req.Header.Get("Authorization") == "" {
		req.Header.Set("Authorization", "Bearer "+tokenAdministrador)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func inicializarToken(t *testing.T) {
	t.Helper()
	tokenAdministrador = tokenParaRol(t, domain.RolAdministrador)
}

func preguntasJSON(n int) string {
	items := make([]string, n)
	for i := range items {
		items[i] = fmt.Sprintf(`{"enunciado":"Pregunta %d","opciones":["Uno","Dos","Tres"],"respuesta_correcta":%d}`, i+1, i%3)
	}
	return `{"preguntas":[` + strings.Join(items, ",") + `]}`
}

func csvPreguntas(n int) string {
	var sb strings.Builder
	sb.WriteString("enunciado,opciones,respuesta_correcta\n")
	for i := 1; i <= n; i++ {
		fmt.Fprintf(&sb, "Pregunta %d,Uno|Dos|Tres,%d\n", i, i%3)
	}
	return sb.String()
}

func cuerpoJSON(t *testing.T, w *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var m map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &m); err != nil {
		t.Fatalf("la respuesta no es JSON: %v (%s)", err, w.Body.String())
	}
	return m
}

// Solo el administrador entra a estas rutas, incluidas las de lectura: el listado trae la
// respuesta correcta de cada pregunta.
func TestEvaluacionAdmin_SoloElAdministrador(t *testing.T) {
	rutas := []struct{ metodo, ruta string }{
		{"POST", rutaEval + "/preguntas"},
		{"GET", rutaEval + "/preguntas"},
		{"PUT", rutaEval + "/preguntas/1"},
		{"DELETE", rutaEval + "/preguntas/1"},
		{"GET", rutaEval + "/config"},
		{"PUT", rutaEval + "/config"},
	}
	for _, rt := range rutas {
		for _, c := range []struct {
			nombre string
			rol    domain.Rol
			want   int
		}{{"sin token", "", 401}, {"postulante", domain.RolPostulante, 403}, {"analista", domain.RolAnalista, 403}} {
			t.Run(rt.metodo+" "+rt.ruta+" "+c.nombre, func(t *testing.T) {
				req := httptest.NewRequest(rt.metodo, rt.ruta, strings.NewReader(`{}`))
				req.Header.Set("Content-Type", "application/json")
				if c.rol != "" {
					req.Header.Set("Authorization", "Bearer "+tokenParaRol(t, c.rol))
				} else {
					req.Header.Set("Authorization", "")
				}
				w := httptest.NewRecorder()
				routerEvaluacion(&bancoEnMemoria{}).ServeHTTP(w, req)
				if w.Code != c.want {
					t.Fatalf("respuesta = %d, se esperaba %d: %s", w.Code, c.want, w.Body.String())
				}
			})
		}
	}
}

func TestEvaluacionAdmin_CrearPreguntaSuelta(t *testing.T) {
	inicializarToken(t)
	b := &bancoEnMemoria{}
	w := pedirEval(routerEvaluacion(b), "POST", rutaEval+"/preguntas", "application/json",
		`{"enunciado":"¿Capital de Ecuador?","opciones":["Guayaquil","Quito"],"respuesta_correcta":1}`)
	if w.Code != 201 {
		t.Fatalf("respuesta = %d %s", w.Code, w.Body.String())
	}
	m := cuerpoJSON(t, w)
	if m["id"] != float64(1) || m["enunciado"] != "¿Capital de Ecuador?" || m["respuesta_correcta"] != float64(1) {
		t.Fatalf("cuerpo = %v", m)
	}
	if len(m) != 4 {
		t.Fatalf("la pregunta del administrador tiene exactamente id, enunciado, opciones y respuesta_correcta: %v", m)
	}
	if len(b.preguntas) != 1 {
		t.Fatal("no se guardó")
	}
}

func TestEvaluacionAdmin_CargaMasivaJSONDe25(t *testing.T) {
	inicializarToken(t)
	b := &bancoEnMemoria{}
	w := pedirEval(routerEvaluacion(b), "POST", rutaEval+"/preguntas", "application/json", preguntasJSON(25))
	if w.Code != 201 || cuerpoJSON(t, w)["creadas"] != float64(25) || len(b.preguntas) != 25 {
		t.Fatalf("respuesta = %d %s, guardadas = %d", w.Code, w.Body.String(), len(b.preguntas))
	}
}

func TestEvaluacionAdmin_CargaMasivaCSVDe25(t *testing.T) {
	inicializarToken(t)
	b := &bancoEnMemoria{}
	w := pedirEval(routerEvaluacion(b), "POST", rutaEval+"/preguntas", "text/csv; charset=utf-8", csvPreguntas(25))
	if w.Code != 201 || cuerpoJSON(t, w)["creadas"] != float64(25) || len(b.preguntas) != 25 {
		t.Fatalf("respuesta = %d %s, guardadas = %d", w.Code, w.Body.String(), len(b.preguntas))
	}
}

func TestEvaluacionAdmin_CargaMasivaCSVPorFormulario(t *testing.T) {
	inicializarToken(t)
	b := &bancoEnMemoria{}
	var cuerpo bytes.Buffer
	mw := multipart.NewWriter(&cuerpo)
	parte, _ := mw.CreateFormFile("archivo", "banco.csv")
	_, _ = parte.Write([]byte(csvPreguntas(25)))
	_ = mw.Close()
	req := httptest.NewRequest("POST", rutaEval+"/preguntas", &cuerpo)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	w := servirEval(routerEvaluacion(b), req)
	if w.Code != 201 || cuerpoJSON(t, w)["creadas"] != float64(25) {
		t.Fatalf("respuesta = %d %s", w.Code, w.Body.String())
	}

	// Sin el campo "archivo" es 400.
	var vacio bytes.Buffer
	mw = multipart.NewWriter(&vacio)
	_ = mw.WriteField("otro", "x")
	_ = mw.Close()
	req = httptest.NewRequest("POST", rutaEval+"/preguntas", &vacio)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	if w := servirEval(routerEvaluacion(b), req); w.Code != 400 || decodeError(t, w).Fields["archivo"] == "" {
		t.Fatalf("sin archivo: %d %s", w.Code, w.Body.String())
	}
}

func TestEvaluacionAdmin_CargaConErroresReportaPorPosicionYNoGuardaNada(t *testing.T) {
	inicializarToken(t)
	b := &bancoEnMemoria{}
	r := routerEvaluacion(b)

	malo := `{"preguntas":[{"enunciado":"Bien","opciones":["a","b"],"respuesta_correcta":0},{"enunciado":"","opciones":["a"],"respuesta_correcta":3}]}`
	w := pedirEval(r, "POST", rutaEval+"/preguntas", "application/json", malo)
	e := decodeError(t, w)
	if w.Code != 400 || e.Code != "VALIDATION_ERROR" || e.Fields["preguntas[1].enunciado"] == "" || e.Fields["preguntas[1].opciones"] == "" {
		t.Fatalf("JSON: %d %s", w.Code, w.Body.String())
	}

	w = pedirEval(r, "POST", rutaEval+"/preguntas", "text/csv", "enunciado,opciones,respuesta_correcta\nBien,a|b,0\nMal,a|b,xx\n")
	e = decodeError(t, w)
	if w.Code != 400 || e.Fields["filas[3].respuesta_correcta"] == "" {
		t.Fatalf("CSV: %d %s", w.Code, w.Body.String())
	}
	if len(b.preguntas) != 0 {
		t.Fatal("una carga con errores no guarda nada")
	}
}

func TestEvaluacionAdmin_ErroresDeFormato(t *testing.T) {
	inicializarToken(t)
	r := routerEvaluacion(&bancoEnMemoria{})
	casos := []struct {
		nombre, tipo, cuerpo string
		want                 int
		campo                string
	}{
		{"tipo de contenido no admitido", "text/plain", "hola", 400, "content-type"},
		{"sin tipo de contenido", "", "hola", 400, "content-type"},
		{"JSON mal formado", "application/json", `{"enunciado":`, 400, ""},
		{"índice de respuesta que no es número", "application/json", `{"enunciado":"x","opciones":["a","b"],"respuesta_correcta":"uno"}`, 400, "respuesta_correcta"},
		{"CSV demasiado grande", "text/csv", "enunciado,opciones,respuesta_correcta\n" + strings.Repeat("x", 1<<20+10), 422, "archivo"},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			w := pedirEval(r, "POST", rutaEval+"/preguntas", c.tipo, c.cuerpo)
			e := decodeError(t, w)
			if w.Code != c.want || e.Code != "VALIDATION_ERROR" || (c.campo != "" && e.Fields[c.campo] == "") {
				t.Fatalf("respuesta = %d %s", w.Code, w.Body.String())
			}
		})
	}
}

func TestEvaluacionAdmin_ListarConBusquedaYPaginas(t *testing.T) {
	inicializarToken(t)
	b := &bancoEnMemoria{}
	r := routerEvaluacion(b)
	pedirEval(r, "POST", rutaEval+"/preguntas", "application/json", preguntasJSON(25))

	w := pedirEval(r, "GET", rutaEval+"/preguntas?por_pagina=10&pagina=3", "", "")
	m := cuerpoJSON(t, w)
	items := m["items"].([]any)
	if w.Code != 200 || len(items) != 5 || m["total"] != float64(25) || m["pagina"] != float64(3) || m["por_pagina"] != float64(10) {
		t.Fatalf("respuesta = %d %s", w.Code, w.Body.String())
	}
	if _, ok := items[0].(map[string]any)["respuesta_correcta"]; !ok {
		t.Fatal("el administrador ve la respuesta correcta en el listado")
	}

	w = pedirEval(r, "GET", rutaEval+"/preguntas?buscar=pregunta%2025", "", "")
	if m := cuerpoJSON(t, w); m["total"] != float64(1) {
		t.Fatalf("búsqueda: %s", w.Body.String())
	}

	// Sin preguntas, "items" es una lista vacía (no null).
	w = pedirEval(routerEvaluacion(&bancoEnMemoria{}), "GET", rutaEval+"/preguntas", "", "")
	if !strings.Contains(w.Body.String(), `"items":[]`) {
		t.Fatalf("items debe ser []: %s", w.Body.String())
	}

	w = pedirEval(r, "GET", rutaEval+"/preguntas?pagina=abc", "", "")
	if w.Code != 400 || decodeError(t, w).Fields["pagina"] == "" {
		t.Fatalf("página inválida: %d %s", w.Code, w.Body.String())
	}
}

func TestEvaluacionAdmin_EditarYEliminar(t *testing.T) {
	inicializarToken(t)
	b := &bancoEnMemoria{}
	r := routerEvaluacion(b)
	pedirEval(r, "POST", rutaEval+"/preguntas", "application/json", preguntasJSON(3))

	w := pedirEval(r, "PUT", rutaEval+"/preguntas/2", "application/json", `{"enunciado":"Cambiada","opciones":["x","y"],"respuesta_correcta":1}`)
	if w.Code != 200 || cuerpoJSON(t, w)["enunciado"] != "Cambiada" || b.preguntas[1].Enunciado != "Cambiada" {
		t.Fatalf("editar: %d %s", w.Code, w.Body.String())
	}
	if w := pedirEval(r, "PUT", rutaEval+"/preguntas/99", "application/json", `{"enunciado":"x","opciones":["x","y"],"respuesta_correcta":0}`); w.Code != 404 || decodeError(t, w).Code != "NOT_FOUND" {
		t.Fatalf("editar inexistente: %d %s", w.Code, w.Body.String())
	}
	if w := pedirEval(r, "DELETE", rutaEval+"/preguntas/2", "", ""); w.Code != 204 || w.Body.Len() != 0 || len(b.preguntas) != 2 {
		t.Fatalf("eliminar: %d %s", w.Code, w.Body.String())
	}
	if w := pedirEval(r, "DELETE", rutaEval+"/preguntas/2", "", ""); w.Code != 404 {
		t.Fatalf("eliminar dos veces: %d", w.Code)
	}
	if w := pedirEval(r, "DELETE", rutaEval+"/preguntas/abc", "", ""); w.Code != 404 {
		t.Fatalf("eliminar con id inválido: %d", w.Code)
	}

	b.intentos = true
	for _, c := range []struct{ metodo, ruta, cuerpo string }{
		{"PUT", rutaEval + "/preguntas/1", `{"enunciado":"x","opciones":["x","y"],"respuesta_correcta":0}`},
		{"DELETE", rutaEval + "/preguntas/1", ""},
	} {
		w := pedirEval(r, c.metodo, c.ruta, "application/json", c.cuerpo)
		if w.Code != 409 || decodeError(t, w).Code != "CONFLICT" {
			t.Fatalf("%s con evaluación iniciada: %d %s", c.metodo, w.Code, w.Body.String())
		}
	}
}

func TestEvaluacionAdmin_Configuracion(t *testing.T) {
	inicializarToken(t)
	b := &bancoEnMemoria{}
	r := routerEvaluacion(b)

	w := pedirEval(r, "GET", rutaEval+"/config", "", "")
	m := cuerpoJSON(t, w)
	if w.Code != 200 || m["configurada"] != false || m["duracion_minutos"] != float64(30) || m["num_preguntas"] != float64(10) {
		t.Fatalf("sin configurar: %d %s", w.Code, w.Body.String())
	}
	// El horario lo fija la fase Evaluación: la configuración ya no tiene fecha de inicio.
	if _, existe := m["inicia_en"]; existe || len(m) != 4 {
		t.Fatalf("la configuración solo trae duración, preguntas, total del banco y configurada: %v", m)
	}

	pedirEval(r, "POST", rutaEval+"/preguntas", "application/json", preguntasJSON(12))
	cuerpo := `{"duracion_minutos":45,"num_preguntas":10}`
	w = pedirEval(r, "PUT", rutaEval+"/config", "application/json", cuerpo)
	m = cuerpoJSON(t, w)
	if w.Code != 200 || m["configurada"] != true || m["num_preguntas"] != float64(10) || m["total_banco"] != float64(12) {
		t.Fatalf("guardar: %d %s", w.Code, w.Body.String())
	}
	if _, existe := m["inicia_en"]; existe {
		t.Fatalf("la respuesta no debe traer inicia_en: %s", w.Body.String())
	}
	// Un inicia_en que mande una versión anterior de la pantalla se ignora: no es error.
	w = pedirEval(r, "PUT", rutaEval+"/config", "application/json", `{"inicia_en":"2026-11-20T09:00:00-05:00","duracion_minutos":40,"num_preguntas":10}`)
	if w.Code != 200 || cuerpoJSON(t, w)["duracion_minutos"] != float64(40) {
		t.Fatalf("con un inicia_en de más: %d %s", w.Code, w.Body.String())
	}
	w = pedirEval(r, "PUT", rutaEval+"/config", "application/json", cuerpo)
	if w.Code != 200 {
		t.Fatalf("volver a 45 minutos: %d %s", w.Code, w.Body.String())
	}
	if w := pedirEval(r, "GET", rutaEval+"/config", "", ""); cuerpoJSON(t, w)["duracion_minutos"] != float64(45) {
		t.Fatalf("lectura: %s", w.Body.String())
	}

	// Más preguntas que las del banco: 422 con el formato de la API.
	w = pedirEval(r, "PUT", rutaEval+"/config", "application/json", `{"duracion_minutos":45,"num_preguntas":13}`)
	e := decodeError(t, w)
	if w.Code != 422 || e.Code != "VALIDATION_ERROR" || e.Fields["num_preguntas"] == "" {
		t.Fatalf("más que el banco: %d %s", w.Code, w.Body.String())
	}
	if b.cfg.NumPreguntas != 10 {
		t.Fatal("la configuración rechazada no debe guardarse")
	}

	for _, c := range []struct{ nombre, cuerpo, campo string }{
		{"sin campos", `{}`, "duracion_minutos"},
		{"sin número de preguntas", `{"duracion_minutos":45}`, "num_preguntas"},
		{"duración que no es número", `{"duracion_minutos":"x","num_preguntas":1}`, "duracion_minutos"},
		{"duración cero", `{"duracion_minutos":0,"num_preguntas":1}`, "duracion_minutos"},
	} {
		w := pedirEval(r, "PUT", rutaEval+"/config", "application/json", c.cuerpo)
		if w.Code != 400 || decodeError(t, w).Fields[c.campo] == "" {
			t.Fatalf("%s: %d %s", c.nombre, w.Code, w.Body.String())
		}
	}
}
