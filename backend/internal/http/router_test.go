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

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/config"
	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/domain"
	apihttp "github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/http"
	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/http/handler"
	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/http/middleware"
)

func init() { gin.SetMode(gin.TestMode) }

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
