package handler_test

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"mime"
	"mime/multipart"
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

// fakeDocumentos imita al servicio y guarda lo que recibió.
type fakeDocumentos struct {
	postulante string
	tipo       string
	nombre     string
	contenido  []byte
	documento  domain.Documento
	err        error
	llamadas   int
}

func (f *fakeDocumentos) Subir(_ context.Context, postulanteID, tipo, nombre string, contenido []byte) (domain.Documento, error) {
	f.llamadas++
	f.postulante, f.tipo, f.nombre, f.contenido = postulanteID, tipo, nombre, contenido
	return f.documento, f.err
}

func (f *fakeDocumentos) Descargar(_ context.Context, postulanteID, _ string) (domain.Documento, error) {
	f.llamadas++
	f.postulante = postulanteID
	return f.documento, f.err
}

func (f *fakeDocumentos) Eliminar(_ context.Context, postulanteID, _ string) error {
	f.llamadas++
	f.postulante = postulanteID
	return f.err
}

func newDocumentoRouter(uc handler.DocumentoUseCase) *gin.Engine {
	return apihttp.NewRouter(apihttp.Deps{
		Config:  config.Config{RequestTimeout: 5 * time.Second},
		Logger:  slog.New(slog.NewTextHandler(io.Discard, nil)),
		Health:  handler.NewHealthHandler(),
		Modules: []apihttp.RouteRegistrar{apihttp.DocumentoRoutes(handler.NewDocumentoHandler(uc), fakeTokens{})},
	})
}

// multipartPDF arma el cuerpo de una subida. Si contenido es nil no agrega el archivo.
func multipartPDF(t *testing.T, tipo, nombre string, contenido []byte) (*bytes.Buffer, string) {
	t.Helper()
	var cuerpo bytes.Buffer
	w := multipart.NewWriter(&cuerpo)
	if err := w.WriteField("tipo", tipo); err != nil {
		t.Fatal(err)
	}
	if contenido != nil {
		parte, err := w.CreateFormFile("archivo", nombre)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := parte.Write(contenido); err != nil {
			t.Fatal(err)
		}
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	return &cuerpo, w.FormDataContentType()
}

func subir(r *gin.Engine, token string, cuerpo io.Reader, tipoContenido string) *httptest.ResponseRecorder {
	req := httptest.NewRequest("POST", "/api/v1/postulacion/documentos", cuerpo)
	req.Header.Set("Content-Type", tipoContenido)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestDocumentoSubir_201ConElFormatoDelContrato(t *testing.T) {
	uc := &fakeDocumentos{documento: domain.Documento{ID: "doc-1", Nombre: "titulo.pdf", Tamano: 8}}
	cuerpo, ct := multipartPDF(t, "formacion", `C:\fakepath\titulo.pdf`, []byte("%PDF-1.7"))

	w := subir(newDocumentoRouter(uc), "postulante", cuerpo, ct)

	if w.Code != 201 {
		t.Fatalf("respuesta = %d %s", w.Code, w.Body.String())
	}
	want := `{"documento_id":"doc-1","nombre":"titulo.pdf","tamano_bytes":8}`
	if strings.TrimSpace(w.Body.String()) != want {
		t.Fatalf("cuerpo = %s; se esperaba %s", w.Body.String(), want)
	}
	if uc.postulante != "p-1" || uc.tipo != "formacion" || uc.nombre != "titulo.pdf" || string(uc.contenido) != "%PDF-1.7" {
		t.Fatalf("el servicio recibió postulante %q, tipo %q, nombre %q, contenido %q", uc.postulante, uc.tipo, uc.nombre, uc.contenido)
	}
}

func TestDocumentoSubir_CuerpoDemasiadoGrande_422SinLlegarAlServicio(t *testing.T) {
	uc := &fakeDocumentos{}
	cuerpo, ct := multipartPDF(t, "organizacion", "grande.pdf", bytes.Repeat([]byte("x"), 6*1024*1024))

	w := subir(newDocumentoRouter(uc), "postulante", cuerpo, ct)

	if w.Code != 422 || !strings.Contains(w.Body.String(), `"code":"VALIDATION_ERROR"`) {
		t.Fatalf("respuesta = %d %s", w.Code, w.Body.String())
	}
	if uc.llamadas != 0 {
		t.Fatal("un cuerpo demasiado grande no debe llegar al servicio")
	}
}

func TestDocumentoSubir_SinArchivo_400(t *testing.T) {
	uc := &fakeDocumentos{}
	cuerpo, ct := multipartPDF(t, "formacion", "", nil)

	w := subir(newDocumentoRouter(uc), "postulante", cuerpo, ct)

	if w.Code != 400 || !strings.Contains(w.Body.String(), `"archivo"`) || uc.llamadas != 0 {
		t.Fatalf("respuesta = %d %s (llamadas %d)", w.Code, w.Body.String(), uc.llamadas)
	}
}

func TestDocumentos_SinSesionOConOtroRol(t *testing.T) {
	for _, token := range []string{"", "analista"} {
		uc := &fakeDocumentos{}
		cuerpo, ct := multipartPDF(t, "formacion", "titulo.pdf", []byte("%PDF-1.7"))
		w := subir(newDocumentoRouter(uc), token, cuerpo, ct)
		if w.Code != 401 && w.Code != 403 {
			t.Errorf("token %q: respuesta = %d %s", token, w.Code, w.Body.String())
		}
		if uc.llamadas != 0 {
			t.Errorf("token %q llegó al servicio", token)
		}
	}
}

func TestDocumentoDescargar_PDFConNombreCodificado(t *testing.T) {
	uc := &fakeDocumentos{documento: domain.Documento{Nombre: `Título "final".pdf`, Contenido: []byte("%PDF-1.7")}}
	req := httptest.NewRequest("GET", "/api/v1/postulacion/documentos/doc-1", nil)
	req.Header.Set("Authorization", "Bearer postulante")
	w := httptest.NewRecorder()
	newDocumentoRouter(uc).ServeHTTP(w, req)

	if w.Code != 200 || w.Header().Get("Content-Type") != "application/pdf" || w.Body.String() != "%PDF-1.7" {
		t.Fatalf("respuesta = %d %q %s", w.Code, w.Header().Get("Content-Type"), w.Body.String())
	}
	_, params, err := mime.ParseMediaType(w.Header().Get("Content-Disposition"))
	if err != nil || params["filename"] != `Título "final".pdf` {
		t.Fatalf("Content-Disposition = %q (%v); el nombre no se puede recuperar", w.Header().Get("Content-Disposition"), err)
	}
}

func TestDocumentoEliminar_CodigosDeRespuesta(t *testing.T) {
	casos := []struct {
		err  error
		want int
	}{
		{nil, 204},
		{domain.Conflict("El documento está siendo usado por la postulación", nil), 409},
		{domain.NotFound("Documento no encontrado"), 404},
	}
	for _, caso := range casos {
		req := httptest.NewRequest("DELETE", "/api/v1/postulacion/documentos/doc-1", nil)
		req.Header.Set("Authorization", "Bearer postulante")
		w := httptest.NewRecorder()
		newDocumentoRouter(&fakeDocumentos{err: caso.err}).ServeHTTP(w, req)
		if w.Code != caso.want {
			t.Errorf("error %v: respuesta = %d %s; se esperaba %d", caso.err, w.Code, w.Body.String(), caso.want)
		}
	}
}
