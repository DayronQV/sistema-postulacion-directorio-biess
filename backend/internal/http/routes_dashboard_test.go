package http_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/domain"
	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/http/handler"
)

type fakeDashboardUseCase struct {
	datos domain.DashboardPostulaciones
	err   error
}

func (f fakeDashboardUseCase) Obtener(context.Context) (domain.DashboardPostulaciones, error) {
	return f.datos, f.err
}

func newDashboardRouter(f fakeDashboardUseCase) *gin.Engine {
	h := handler.NewAdminDashboardHandler(f)
	return newAdminRouterConModulos(h.Rutas)
}

func TestDashboard_LecturaParaAnalistaYAdministrador(t *testing.T) {
	datos := domain.DashboardPostulaciones{
		PorCargo:  domain.ConteosPorCargo{Afiliados: 3, Jubilados: 2, Empleador: 1},
		PorEstado: domain.ConteosPorEstado{Borrador: 1, Enviada: 4, Cerrada: 1},
		Total:     6,
	}
	for _, rol := range []domain.Rol{domain.RolAnalista, domain.RolAdministrador} {
		w := pedirAdmin(newDashboardRouter(fakeDashboardUseCase{datos: datos}), "GET", "/api/v1/admin/dashboard", tokenParaRol(t, rol), "")
		if w.Code != 200 || !strings.Contains(w.Body.String(), `"afiliados":3`) ||
			!strings.Contains(w.Body.String(), `"enviada":4`) || !strings.Contains(w.Body.String(), `"total":6`) {
			t.Fatalf("rol %s: respuesta = %d %s", rol, w.Code, w.Body.String())
		}
	}
}

func TestDashboard_RechazaPostulanteYSesionAusente(t *testing.T) {
	for _, token := range []string{"", tokenParaRol(t, domain.RolPostulante)} {
		w := pedirAdmin(newDashboardRouter(fakeDashboardUseCase{}), "GET", "/api/v1/admin/dashboard", token, "")
		if w.Code != 401 && w.Code != 403 {
			t.Fatalf("respuesta = %d %s", w.Code, w.Body.String())
		}
	}
}

func TestDashboard_ErrorUsaElFormatoComun(t *testing.T) {
	w := pedirAdmin(newDashboardRouter(fakeDashboardUseCase{err: errors.New("boom")}), "GET", "/api/v1/admin/dashboard", tokenParaRol(t, domain.RolAnalista), "")
	if w.Code != 500 || decodeError(t, w).Code != "INTERNAL_ERROR" {
		t.Fatalf("respuesta = %d %s", w.Code, w.Body.String())
	}
}
