package handler

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/domain"
	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/http/middleware"
	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/service"
)

// FasesUseCase es lo que el handler necesita para gestionar las fases del concurso.
type FasesUseCase interface {
	Listar(ctx context.Context) ([]domain.Fase, error)
	Actualizar(ctx context.Context, id int, c service.CambiosFase) (domain.Fase, error)
}

// AdminFasesHandler atiende las fases del concurso en el panel (EF-13).
type AdminFasesHandler struct {
	fases FasesUseCase
}

// NewAdminFasesHandler crea el handler.
func NewAdminFasesHandler(fases FasesUseCase) *AdminFasesHandler {
	return &AdminFasesHandler{fases: fases}
}

// PUT reemplaza el estado de la fase. Los dos interruptores son obligatorios (por eso son
// punteros): un cuerpo vacío o incompleto no debe apagar la publicación ni el formulario sin
// querer. Las fechas son opcionales: ausente o null es sin fecha. Llegan como texto RFC 3339
// y se leen aquí, para responder 400 con el campo si vienen mal formadas.
type faseRequest struct {
	Publicada          *bool   `json:"publicada"`
	HabilitaFormulario *bool   `json:"habilita_formulario"`
	FechaInicio        *string `json:"fecha_inicio"`
	FechaFin           *string `json:"fecha_fin"`
}

// fechaOpcional lee una fecha RFC 3339; null o vacía es sin fecha.
func fechaOpcional(v *string, campo string, fields map[string]string) *time.Time {
	if v == nil || *v == "" {
		return nil
	}
	t, err := time.Parse(time.RFC3339, *v)
	if err != nil {
		fields[campo] = "Fecha no válida"
		return nil
	}
	return &t
}

type faseResponse struct {
	ID                 int        `json:"id"`
	Nombre             string     `json:"nombre"`
	Orden              int        `json:"orden"`
	Publicada          bool       `json:"publicada"`
	HabilitaFormulario bool       `json:"habilita_formulario"`
	FechaInicio        *time.Time `json:"fecha_inicio"`
	FechaFin           *time.Time `json:"fecha_fin"`
}

func toFaseResponse(f domain.Fase) faseResponse {
	return faseResponse{
		ID: f.ID, Nombre: f.Nombre, Orden: f.Orden,
		Publicada: f.Publicada, HabilitaFormulario: f.HabilitaFormulario,
		FechaInicio: f.FechaInicio, FechaFin: f.FechaFin,
	}
}

// Listar atiende GET /admin/fases.
func (h *AdminFasesHandler) Listar(c *gin.Context) {
	fases, err := h.fases.Listar(c.Request.Context())
	if err != nil {
		middleware.Fail(c, err)
		return
	}
	resp := make([]faseResponse, 0, len(fases))
	for _, f := range fases {
		resp = append(resp, toFaseResponse(f))
	}
	c.JSON(http.StatusOK, resp)
}

// Actualizar atiende PUT /admin/fases/{id}.
func (h *AdminFasesHandler) Actualizar(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		middleware.Fail(c, domain.NotFound("La fase no existe"))
		return
	}
	var req faseRequest
	if !middleware.BindJSON(c, &req) {
		return
	}
	fields := map[string]string{}
	if req.Publicada == nil {
		fields["publicada"] = "Campo obligatorio"
	}
	if req.HabilitaFormulario == nil {
		fields["habilita_formulario"] = "Campo obligatorio"
	}
	cambios := service.CambiosFase{
		FechaInicio: fechaOpcional(req.FechaInicio, "fecha_inicio", fields),
		FechaFin:    fechaOpcional(req.FechaFin, "fecha_fin", fields),
	}
	if len(fields) > 0 {
		middleware.Fail(c, domain.Validation("Datos inválidos", fields))
		return
	}
	cambios.Publicada, cambios.HabilitaFormulario = *req.Publicada, *req.HabilitaFormulario
	f, err := h.fases.Actualizar(c.Request.Context(), id, cambios)
	if err != nil {
		middleware.Fail(c, err)
		return
	}
	c.JSON(http.StatusOK, toFaseResponse(f))
}

// Rutas registra las rutas del módulo: el analista consulta, solo el administrador cambia.
func (h *AdminFasesHandler) Rutas(lectura, escritura *gin.RouterGroup) {
	lectura.GET("/fases", h.Listar)
	escritura.PUT("/fases/:id", h.Actualizar)
}
