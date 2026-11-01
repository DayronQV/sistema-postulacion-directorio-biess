package http_test

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	nethttp "net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/auth"
	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/config"
	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/domain"
	apihttp "github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/http"
	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/http/handler"
	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/http/middleware"
)

func init() { gin.SetMode(gin.TestMode) }

// emisorPrueba firma los tokens de las rutas protegidas de esta prueba.
var emisorPrueba = func() *auth.JWTIssuer {
	i, err := auth.NewJWTIssuer("secreto-de-prueba-con-mas-de-32-caracteres", time.Hour)
	if err != nil {
		panic(err)
	}
	return i
}()

func tokenParaRol(t *testing.T, rol domain.Rol) string {
	t.Helper()
	e, err := emisorPrueba.Emitir(domain.Claims{UsuarioID: "u-1", Cedula: "1712345675", Rol: rol})
	if err != nil {
		t.Fatal(err)
	}
	return e.Token
}

// bindReq simula el cuerpo de un endpoint real, con validaciones `binding`.
type bindReq struct {
	Cedula string `json:"cedula" binding:"required"`
	Email  string `json:"email" binding:"required,email"`
	Edad   int    `json:"edad"`
}

// newTestRouter arma el router real con rutas de prueba que simulan lo que harán los
// módulos: devolver errores de dominio, fallar o tardar.
func newTestRouter(timeout time.Duration) *gin.Engine {
	cfg := config.Config{
		RequestTimeout:     timeout,
		CORSAllowedOrigins: []string{"http://localhost:5173"},
	}
	test := func(api *gin.RouterGroup) {
		api.GET("/ok", func(c *gin.Context) { c.JSON(nethttp.StatusOK, gin.H{"ok": true}) })
		api.GET("/validation", func(c *gin.Context) {
			middleware.Fail(c, domain.Validation("Datos inválidos", map[string]string{"cedula": "Cédula no válida"}))
		})
		api.GET("/locked", func(c *gin.Context) {
			middleware.Fail(c, domain.Locked("registro fallido, inténtelo nuevamente en 10 minutos", 10*time.Minute))
		})
		api.GET("/internal", func(c *gin.Context) {
			middleware.Fail(c, errors.New("password authentication failed for user biess"))
		})
		api.GET("/panic", func(c *gin.Context) { panic("boom") })
		// Rutas protegidas: una con sesión y otra solo para administradores.
		api.GET("/protegida", middleware.RequireAuth(emisorPrueba), func(c *gin.Context) {
			claims, _ := middleware.ClaimsFrom(c)
			c.JSON(nethttp.StatusOK, gin.H{"usuario": claims.UsuarioID, "rol": claims.Rol})
		})
		api.GET("/solo-admin", middleware.RequireAuth(emisorPrueba), middleware.RequireRole(domain.RolAdministrador),
			func(c *gin.Context) { c.JSON(nethttp.StatusOK, gin.H{"ok": true}) })
		// Handler que usa el helper BindJSON.
		api.POST("/bind", func(c *gin.Context) {
			var req bindReq
			if !middleware.BindJSON(c, &req) {
				return
			}
			c.JSON(nethttp.StatusOK, gin.H{"cedula": req.Cedula})
		})
		// Handler que pasa el error de Gin tal cual, sin usar el helper.
		api.POST("/rawbind", func(c *gin.Context) {
			var req bindReq
			if err := c.ShouldBindJSON(&req); err != nil {
				middleware.Fail(c, err)
				return
			}
			c.JSON(nethttp.StatusOK, gin.H{"ok": true})
		})
		api.GET("/slow", func(c *gin.Context) {
			select {
			case <-c.Request.Context().Done():
				middleware.Fail(c, c.Request.Context().Err())
			case <-time.After(2 * time.Second):
				c.JSON(nethttp.StatusOK, gin.H{"ok": true})
			}
		})
	}
	return apihttp.NewRouter(apihttp.Deps{
		Config:  cfg,
		Logger:  slog.New(slog.NewTextHandler(io.Discard, nil)),
		Health:  handler.NewHealthHandler(),
		Modules: []apihttp.RouteRegistrar{test},
	})
}

func do(r *gin.Engine, method, path string, headers map[string]string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, nil)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func doJSON(r *gin.Engine, path, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest("POST", path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func decodeError(t *testing.T, w *httptest.ResponseRecorder) middleware.ErrorPayload {
	t.Helper()
	var body middleware.ErrorBody
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("la respuesta no es JSON de error: %v\n%s", err, w.Body.String())
	}
	return body.Error
}

func TestHealth(t *testing.T) {
	w := do(newTestRouter(time.Second), "GET", "/health", nil)
	if w.Code != 200 || strings.TrimSpace(w.Body.String()) != `{"status":"ok"}` {
		t.Fatalf("/health = %d %s", w.Code, w.Body.String())
	}
}

func TestModuleRoutesAreUnderAPIV1(t *testing.T) {
	r := newTestRouter(time.Second)
	if w := do(r, "GET", "/api/v1/ok", nil); w.Code != 200 {
		t.Fatalf("/api/v1/ok = %d", w.Code)
	}
	if w := do(r, "GET", "/ok", nil); w.Code != 404 {
		t.Fatalf("/ok sin prefijo debería ser 404, got %d", w.Code)
	}
}

func TestDomainErrorUsesAPIFormat(t *testing.T) {
	w := do(newTestRouter(time.Second), "GET", "/api/v1/validation", nil)
	if w.Code != 400 {
		t.Fatalf("status = %d", w.Code)
	}
	e := decodeError(t, w)
	if e.Code != "VALIDATION_ERROR" || e.Message != "Datos inválidos" || e.Fields["cedula"] != "Cédula no válida" {
		t.Fatalf("cuerpo incorrecto: %+v", e)
	}
}

func TestLockedIncludesRetryAfter(t *testing.T) {
	w := do(newTestRouter(time.Second), "GET", "/api/v1/locked", nil)
	if w.Code != 429 {
		t.Fatalf("status = %d", w.Code)
	}
	e := decodeError(t, w)
	if e.Code != "LOCKED" || e.RetryAfterSeconds != 600 {
		t.Fatalf("cuerpo incorrecto: %+v", e)
	}
	if got := w.Header().Get("Retry-After"); got != "600" {
		t.Fatalf("Retry-After = %q", got)
	}
}

func TestInternalErrorDoesNotLeakDetails(t *testing.T) {
	w := do(newTestRouter(time.Second), "GET", "/api/v1/internal", nil)
	if w.Code != 500 {
		t.Fatalf("status = %d", w.Code)
	}
	if strings.Contains(w.Body.String(), "password") {
		t.Fatalf("el 500 expone detalles internos: %s", w.Body.String())
	}
	if e := decodeError(t, w); e.Code != "INTERNAL_ERROR" {
		t.Fatalf("code = %s", e.Code)
	}
}

func TestPanicIsRecoveredAsInternalError(t *testing.T) {
	w := do(newTestRouter(time.Second), "GET", "/api/v1/panic", nil)
	if w.Code != 500 || decodeError(t, w).Code != "INTERNAL_ERROR" {
		t.Fatalf("panic = %d %s", w.Code, w.Body.String())
	}
	if strings.Contains(w.Body.String(), "boom") {
		t.Fatal("el 500 expone el valor del pánico")
	}
}

// Un cuerpo inválido es un error del cliente (400 con `fields`), nunca un 500.
func TestInvalidBodyIsBadRequestNotInternalError(t *testing.T) {
	r := newTestRouter(time.Second)
	cases := []struct {
		name    string
		path    string
		body    string
		fields  []string
		message string
	}{
		{"JSON truncado", "/api/v1/bind", `{"cedula": `, nil, "no es un JSON válido"},
		{"JSON con sintaxis inválida", "/api/v1/bind", `{cedula: 1}`, nil, "no es un JSON válido"},
		{"cuerpo vacío", "/api/v1/bind", ``, nil, "vacío"},
		{"campos obligatorios ausentes", "/api/v1/bind", `{}`, []string{"cedula", "email"}, "Datos inválidos"},
		{"correo con formato inválido", "/api/v1/bind", `{"cedula":"1","email":"x"}`, []string{"email"}, "Datos inválidos"},
		{"tipo de dato incorrecto", "/api/v1/bind", `{"cedula":"1","email":"a@b.com","edad":"mucha"}`, []string{"edad"}, "Datos inválidos"},
		{"sin el helper: JSON truncado", "/api/v1/rawbind", `{"cedula": `, nil, "no es un JSON válido"},
		{"sin el helper: campos ausentes", "/api/v1/rawbind", `{}`, []string{"cedula", "email"}, "Datos inválidos"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			w := doJSON(r, c.path, c.body)
			if w.Code != 400 {
				t.Fatalf("status = %d, se esperaba 400: %s", w.Code, w.Body.String())
			}
			e := decodeError(t, w)
			if e.Code != "VALIDATION_ERROR" || !strings.Contains(e.Message, c.message) {
				t.Fatalf("cuerpo incorrecto: %+v", e)
			}
			for _, f := range c.fields {
				if e.Fields[f] == "" {
					t.Errorf("falta el campo %q en fields: %v", f, e.Fields)
				}
			}
		})
	}
}

func TestValidBodyPassesBinding(t *testing.T) {
	w := doJSON(newTestRouter(time.Second), "/api/v1/bind", `{"cedula":"1712345675","email":"a@b.com"}`)
	if w.Code != 200 {
		t.Fatalf("status = %d: %s", w.Code, w.Body.String())
	}
}

// Criterio de la issue: una ruta protegida responde 401 sin token, 403 con rol
// incorrecto y 200 con token válido, todo con el formato de error de la API.
func TestRutaProtegida(t *testing.T) {
	r := newTestRouter(time.Second)
	conToken := func(ruta, token string) *httptest.ResponseRecorder {
		h := map[string]string{}
		if token != "" {
			h["Authorization"] = "Bearer " + token
		}
		return do(r, "GET", ruta, h)
	}

	w := conToken("/api/v1/protegida", "")
	if w.Code != 401 || decodeError(t, w).Code != "UNAUTHORIZED" {
		t.Fatalf("sin token: %d %s, se esperaba 401 UNAUTHORIZED", w.Code, w.Body.String())
	}

	w = conToken("/api/v1/protegida", "token-inventado")
	if w.Code != 401 {
		t.Fatalf("token inválido: %d %s, se esperaba 401", w.Code, w.Body.String())
	}

	w = conToken("/api/v1/solo-admin", tokenParaRol(t, domain.RolPostulante))
	if w.Code != 403 || decodeError(t, w).Code != "FORBIDDEN" {
		t.Fatalf("rol incorrecto: %d %s, se esperaba 403 FORBIDDEN", w.Code, w.Body.String())
	}

	w = conToken("/api/v1/protegida", tokenParaRol(t, domain.RolPostulante))
	if w.Code != 200 || !strings.Contains(w.Body.String(), `"usuario":"u-1"`) {
		t.Fatalf("token válido: %d %s, se esperaba 200 con la identidad", w.Code, w.Body.String())
	}

	w = conToken("/api/v1/solo-admin", tokenParaRol(t, domain.RolAdministrador))
	if w.Code != 200 {
		t.Fatalf("rol correcto: %d %s, se esperaba 200", w.Code, w.Body.String())
	}
}

func TestUnknownRouteReturnsNotFoundInAPIFormat(t *testing.T) {
	w := do(newTestRouter(time.Second), "GET", "/api/v1/no-existe", nil)
	if w.Code != 404 || decodeError(t, w).Code != "NOT_FOUND" {
		t.Fatalf("ruta inexistente = %d %s", w.Code, w.Body.String())
	}
}

func TestRequestTimeoutCancelsContext(t *testing.T) {
	start := time.Now()
	w := do(newTestRouter(50*time.Millisecond), "GET", "/api/v1/slow", nil)
	if time.Since(start) > time.Second {
		t.Fatal("la petición no se canceló por timeout")
	}
	if w.Code != 500 {
		t.Fatalf("status = %d", w.Code)
	}
	if e := decodeError(t, w); e.Code != "INTERNAL_ERROR" || !strings.Contains(e.Message, "tardó") {
		t.Fatalf("cuerpo incorrecto: %+v", e)
	}
}

func TestRequestIDIsGeneratedAndPropagated(t *testing.T) {
	r := newTestRouter(time.Second)

	if id := do(r, "GET", "/health", nil).Header().Get("X-Request-ID"); len(id) < 8 {
		t.Fatalf("no se generó X-Request-ID: %q", id)
	}
	if id := do(r, "GET", "/health", map[string]string{"X-Request-ID": "abc-123"}).Header().Get("X-Request-ID"); id != "abc-123" {
		t.Fatalf("no se conservó el id del cliente: %q", id)
	}
	// Un id con caracteres peligrosos se reemplaza.
	if id := do(r, "GET", "/health", map[string]string{"X-Request-ID": "a b\"c"}).Header().Get("X-Request-ID"); id == "a b\"c" {
		t.Fatal("se aceptó un X-Request-ID inseguro")
	}
}

func TestCORS(t *testing.T) {
	r := newTestRouter(time.Second)

	w := do(r, "GET", "/health", map[string]string{"Origin": "http://localhost:5173"})
	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:5173" {
		t.Fatalf("origen permitido: Allow-Origin = %q", got)
	}

	w = do(r, "GET", "/health", map[string]string{"Origin": "http://malicioso.com"})
	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("origen no permitido recibió Allow-Origin = %q", got)
	}

	w = do(r, "OPTIONS", "/api/v1/ok", map[string]string{
		"Origin":                         "http://localhost:5173",
		"Access-Control-Request-Method":  "POST",
		"Access-Control-Request-Headers": "authorization",
	})
	if w.Code != 204 || !strings.Contains(w.Header().Get("Access-Control-Allow-Headers"), "Authorization") {
		t.Fatalf("preflight = %d, headers %v", w.Code, w.Header())
	}
}
