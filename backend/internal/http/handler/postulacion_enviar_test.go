package handler_test

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/domain"
)

func enviar(uc *fakePostulacion, token string) *httptest.ResponseRecorder {
	req := httptest.NewRequest("POST", "/api/v1/postulacion/enviar", strings.NewReader(""))
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	newPostulacionRouter(uc).ServeHTTP(w, req)
	return w
}

func TestPostulacionEnviar_OK(t *testing.T) {
	p := postulacionDeEjemplo()
	p.Estado = domain.EstadoEnviada
	uc := &fakePostulacion{resultado: p}

	w := enviar(uc, "postulante")
	if w.Code != 200 {
		t.Fatalf("respuesta = %d %s", w.Code, w.Body.String())
	}
	var got struct {
		Codigo string `json:"codigo"`
		Estado string `json:"estado"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if uc.enviarID != "p-1" || got.Estado != "enviada" || got.Codigo != "BIESS-2026-000001" {
		t.Fatalf("enviarID = %q, cuerpo = %s", uc.enviarID, w.Body.String())
	}
}

func TestPostulacionEnviar_ErroresConElFormatoDeLaApi(t *testing.T) {
	casos := []struct {
		nombre string
		err    error
		code   int
		tipo   string
	}{
		{"incompleta", domain.Unprocessable("La postulación está incompleta", map[string]string{"cargo": "Seleccione el cargo"}), 422, "VALIDATION_ERROR"},
		{"ya enviada", domain.Conflict("La postulación ya fue enviada", nil), 409, "CONFLICT"},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			w := enviar(&fakePostulacion{err: c.err}, "postulante")
			var got struct {
				Error struct {
					Code   string            `json:"code"`
					Fields map[string]string `json:"fields"`
				} `json:"error"`
			}
			if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
				t.Fatal(err)
			}
			if w.Code != c.code || got.Error.Code != c.tipo {
				t.Fatalf("respuesta = %d %s", w.Code, w.Body.String())
			}
			if c.code == 422 && got.Error.Fields["cargo"] == "" {
				t.Fatalf("falta fields.cargo: %s", w.Body.String())
			}
		})
	}
}

func TestPostulacionEnviar_ExigeSesionDePostulante(t *testing.T) {
	if w := enviar(&fakePostulacion{}, ""); w.Code != 401 {
		t.Fatalf("sin token = %d", w.Code)
	}
	if w := enviar(&fakePostulacion{}, "analista"); w.Code != 403 {
		t.Fatalf("analista = %d", w.Code)
	}
}
