package http_test

import (
	"context"
	"io"
	"log/slog"
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

type fakeConsultasPostulacion struct {
	filtro        service.FiltroPostulacionesInput
	pagina        service.PaginaPostulaciones
	postulacionID string
	documentoID   string
	detalle       domain.Postulacion
	documento     domain.Documento
	err           error
}

func (f *fakeConsultasPostulacion) Listar(_ context.Context, filtro service.FiltroPostulacionesInput) (service.PaginaPostulaciones, error) {
	f.filtro = filtro
	return f.pagina, f.err
}

func (f *fakeConsultasPostulacion) Obtener(_ context.Context, postulacionID string) (domain.Postulacion, error) {
	f.postulacionID = postulacionID
	return f.detalle, f.err
}

func (f *fakeConsultasPostulacion) DescargarDocumento(_ context.Context, postulacionID, documentoID string) (domain.Documento, error) {
	f.postulacionID, f.documentoID = postulacionID, documentoID
	return f.documento, f.err
}

func newConsultasRouter(f *fakeConsultasPostulacion) *gin.Engine {
	h := handler.NewAdminPostulacionHandler(&fakeMerito{}).ConConsultas(f)
	return apihttp.NewRouter(apihttp.Deps{
		Config: config.Config{RequestTimeout: time.Second},
		Logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
		Health: handler.NewHealthHandler(),
		Modules: []apihttp.RouteRegistrar{
			apihttp.AdminRoutes(handler.NewAdminAuthHandler(&fakeAdminLogin{}), emisorPrueba, h.Rutas),
		},
	})
}

func TestNominaAdmin_ListaConFiltrosParaAmbosRoles(t *testing.T) {
	for _, rol := range []domain.Rol{domain.RolAnalista, domain.RolAdministrador} {
		f := &fakeConsultasPostulacion{pagina: service.PaginaPostulaciones{
			Items: []domain.ResumenPostulacionAdmin{{
				ID: "11111111-2222-3333-4444-555555555555", Codigo: "BIESS-2026-000001",
				Cedula: "0912345678", Nombres: "Ana", Apellidos: "Pérez", Cargo: domain.CargoAfiliados,
				Estado: domain.EstadoEnviada, Elegible: true,
			}},
			Total: 21, Pagina: 2, PorPagina: 20,
		}}
		w := pedirAdmin(newConsultasRouter(f), "GET", "/api/v1/admin/postulaciones?cargo=afiliados&estado=enviada&elegible=true&pagina=2", tokenParaRol(t, rol), "")
		if w.Code != 200 || !strings.Contains(w.Body.String(), `"codigo":"BIESS-2026-000001"`) ||
			!strings.Contains(w.Body.String(), `"total":21`) || !strings.Contains(w.Body.String(), `"por_pagina":20`) {
			t.Fatalf("rol %s: respuesta = %d %s", rol, w.Code, w.Body.String())
		}
		if f.filtro.Cargo != "afiliados" || f.filtro.Estado != "enviada" || f.filtro.Elegible != "true" || f.filtro.Pagina != 2 {
			t.Fatalf("filtro = %+v", f.filtro)
		}
	}
}

func TestNominaAdmin_PaginaNoNumericaEs400(t *testing.T) {
	w := pedirAdmin(newConsultasRouter(&fakeConsultasPostulacion{}), "GET", "/api/v1/admin/postulaciones?pagina=dos", tokenParaRol(t, domain.RolAnalista), "")
	if w.Code != 400 || decodeError(t, w).Fields["pagina"] == "" {
		t.Fatalf("respuesta = %d %s", w.Code, w.Body.String())
	}
}

func TestNominaAdmin_DetalleUsaElFormularioCompleto(t *testing.T) {
	f := &fakeConsultasPostulacion{detalle: domain.Postulacion{
		ID: "11111111-2222-3333-4444-555555555555", Codigo: "BIESS-2026-000001", Estado: domain.EstadoEnviada,
		Personal:   domain.InformacionPersonal{Cedula: "0912345678", Nombres: "Ana", Apellidos: "Pérez", FechaNacimiento: time.Date(1990, 1, 2, 0, 0, 0, 0, time.UTC)},
		Formulario: domain.Formulario{Cargo: domain.CargoAfiliados, DuracionExperiencia: domain.Duracion{Anios: 5}},
	}}
	ruta := "/api/v1/admin/postulaciones/11111111-2222-3333-4444-555555555555"
	w := pedirAdmin(newConsultasRouter(f), "GET", ruta, tokenParaRol(t, domain.RolAnalista), "")
	if w.Code != 200 || !strings.Contains(w.Body.String(), `"codigo":"BIESS-2026-000001"`) ||
		!strings.Contains(w.Body.String(), `"duracion_experiencia_total":{"anios":5`) {
		t.Fatalf("respuesta = %d %s", w.Code, w.Body.String())
	}
}

func TestNominaAdmin_DescargaDocumentoConNombre(t *testing.T) {
	f := &fakeConsultasPostulacion{documento: domain.Documento{Nombre: "certificado académico.pdf", Contenido: []byte("%PDF-prueba")}}
	ruta := "/api/v1/admin/postulaciones/11111111-2222-3333-4444-555555555555/documentos/aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"
	w := pedirAdmin(newConsultasRouter(f), "GET", ruta, tokenParaRol(t, domain.RolAdministrador), "")
	if w.Code != 200 || w.Header().Get("Content-Type") != "application/pdf" || string(w.Body.Bytes()) != "%PDF-prueba" ||
		!strings.Contains(w.Header().Get("Content-Disposition"), "attachment") {
		t.Fatalf("respuesta = %d, cabeceras = %v, cuerpo = %q", w.Code, w.Header(), w.Body.String())
	}
}

func TestNominaAdmin_PostulanteNoPuedeConsultar(t *testing.T) {
	w := pedirAdmin(newConsultasRouter(&fakeConsultasPostulacion{}), "GET", "/api/v1/admin/postulaciones", tokenParaRol(t, domain.RolPostulante), "")
	if w.Code != 403 {
		t.Fatalf("respuesta = %d %s", w.Code, w.Body.String())
	}
}
