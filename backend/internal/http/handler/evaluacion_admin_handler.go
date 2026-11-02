package handler

import (
	"bytes"
	"context"
	"errors"
	"io"
	"mime"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/domain"
	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/http/middleware"
	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/service"
)

// maximoArchivoPreguntas es el tamaño máximo de un CSV de preguntas.
const maximoArchivoPreguntas = 1 << 20 // 1 MiB

// EvaluacionAdminUseCase es lo que el handler necesita del servicio del banco de preguntas.
type EvaluacionAdminUseCase interface {
	CrearPregunta(ctx context.Context, in service.PreguntaInput) (domain.Pregunta, error)
	CrearPreguntas(ctx context.Context, entradas []service.PreguntaInput) (int, error)
	CrearPreguntasCSV(ctx context.Context, r io.Reader) (int, error)
	ListarPreguntas(ctx context.Context, buscar string, pagina, porPagina int) (service.PaginaPreguntas, error)
	ActualizarPregunta(ctx context.Context, id string, in service.PreguntaInput) (domain.Pregunta, error)
	EliminarPregunta(ctx context.Context, id string) error
	ObtenerConfig(ctx context.Context) (service.ConfigEvaluacion, error)
	GuardarConfig(ctx context.Context, in service.ConfigInput) (service.ConfigEvaluacion, error)
}

// EvaluacionAdminHandler atiende el banco de preguntas y la configuración de la evaluación
// en el panel de administración.
type EvaluacionAdminHandler struct {
	evaluacion EvaluacionAdminUseCase
}

// NewEvaluacionAdminHandler crea el handler.
func NewEvaluacionAdminHandler(evaluacion EvaluacionAdminUseCase) *EvaluacionAdminHandler {
	return &EvaluacionAdminHandler{evaluacion: evaluacion}
}

// preguntaRequest es una pregunta en JSON. respuesta_correcta es el índice (desde 0) de la
// opción correcta; es un puntero para distinguir "no vino" de 0.
type preguntaRequest struct {
	Enunciado         string   `json:"enunciado"`
	Opciones          []string `json:"opciones"`
	RespuestaCorrecta *int     `json:"respuesta_correcta"`
}

func (p preguntaRequest) entrada() service.PreguntaInput {
	return service.PreguntaInput{Enunciado: p.Enunciado, Opciones: p.Opciones, RespuestaCorrecta: p.RespuestaCorrecta}
}

// crearPreguntasRequest admite una pregunta suelta o, con "preguntas", una carga en bloque.
// Los campos de la pregunta suelta van declarados aquí y no como una estructura embebida:
// con una embebida, el nombre del campo en un error de tipo cambia según la versión de Go
// ("preguntaRequest.respuesta_correcta" en vez de "respuesta_correcta").
type crearPreguntasRequest struct {
	Enunciado         string             `json:"enunciado"`
	Opciones          []string           `json:"opciones"`
	RespuestaCorrecta *int               `json:"respuesta_correcta"`
	Preguntas         *[]preguntaRequest `json:"preguntas"`
}

func (r crearPreguntasRequest) suelta() preguntaRequest {
	return preguntaRequest{Enunciado: r.Enunciado, Opciones: r.Opciones, RespuestaCorrecta: r.RespuestaCorrecta}
}

type configRequest struct {
	IniciaEn        string `json:"inicia_en"`
	DuracionMinutos *int   `json:"duracion_minutos"`
	NumPreguntas    *int   `json:"num_preguntas"`
}

type configResponse struct {
	IniciaEn        *time.Time `json:"inicia_en"`
	DuracionMinutos int        `json:"duracion_minutos"`
	NumPreguntas    int        `json:"num_preguntas"`
	TotalBanco      int        `json:"total_banco"`
	Configurada     bool       `json:"configurada"`
}

// respuestaConfig arma la respuesta. La hora de inicio sale siempre en UTC, igual al guardar
// y al leer, sin importar el desfase con el que la envió el administrador.
func respuestaConfig(c service.ConfigEvaluacion) configResponse {
	var inicia *time.Time
	if c.IniciaEn != nil {
		utc := c.IniciaEn.UTC()
		inicia = &utc
	}
	return configResponse{
		IniciaEn: inicia, DuracionMinutos: c.DuracionMinutos, NumPreguntas: c.NumPreguntas,
		TotalBanco: c.TotalBanco, Configurada: c.Configurada,
	}
}

// CrearPreguntas atiende POST /admin/evaluacion/preguntas. Acepta:
//   - application/json con una pregunta ({enunciado, opciones, respuesta_correcta}): 201 con
//     la pregunta creada;
//   - application/json con {preguntas: [...]}: carga en bloque, 201 {creadas};
//   - text/csv, o multipart con el campo "archivo": carga en bloque desde un CSV.
//
// Una carga en bloque guarda todas las preguntas o ninguna.
func (h *EvaluacionAdminHandler) CrearPreguntas(c *gin.Context) {
	tipo, _, _ := mime.ParseMediaType(c.GetHeader("Content-Type"))
	switch tipo {
	case "application/json":
		h.crearDesdeJSON(c)
	case "text/csv":
		h.crearDesdeCSV(c, c.Request.Body)
	case "multipart/form-data":
		h.crearDesdeMultipart(c)
	default:
		middleware.Fail(c, domain.Validation("Tipo de contenido no admitido", map[string]string{
			"content-type": "Use application/json, text/csv o multipart/form-data con el campo archivo",
		}))
	}
}

func (h *EvaluacionAdminHandler) crearDesdeJSON(c *gin.Context) {
	var req crearPreguntasRequest
	if !middleware.BindJSON(c, &req) {
		return
	}
	if req.Preguntas != nil {
		entradas := make([]service.PreguntaInput, len(*req.Preguntas))
		for i, p := range *req.Preguntas {
			entradas[i] = p.entrada()
		}
		n, err := h.evaluacion.CrearPreguntas(c.Request.Context(), entradas)
		if err != nil {
			middleware.Fail(c, err)
			return
		}
		c.JSON(http.StatusCreated, gin.H{"creadas": n})
		return
	}
	p, err := h.evaluacion.CrearPregunta(c.Request.Context(), req.suelta().entrada())
	if err != nil {
		middleware.Fail(c, err)
		return
	}
	c.JSON(http.StatusCreated, p)
}

func (h *EvaluacionAdminHandler) crearDesdeMultipart(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maximoArchivoPreguntas+(64<<10))
	if err := c.Request.ParseMultipartForm(maximoArchivoPreguntas); err != nil {
		middleware.Fail(c, errorDeArchivo(err))
		return
	}
	defer c.Request.MultipartForm.RemoveAll()
	archivo, _, err := c.Request.FormFile("archivo")
	if err != nil {
		middleware.Fail(c, domain.Validation("Datos inválidos", map[string]string{"archivo": "Debe seleccionar un archivo"}))
		return
	}
	defer archivo.Close()
	h.crearDesdeCSV(c, archivo)
}

func (h *EvaluacionAdminHandler) crearDesdeCSV(c *gin.Context, origen io.Reader) {
	contenido, err := io.ReadAll(io.LimitReader(origen, maximoArchivoPreguntas+1))
	if err != nil {
		middleware.Fail(c, errorDeArchivo(err))
		return
	}
	if len(contenido) > maximoArchivoPreguntas {
		middleware.Fail(c, domain.Unprocessable("El archivo supera el tamaño permitido", map[string]string{"archivo": "El archivo no puede pasar de 1 MB"}))
		return
	}
	n, err := h.evaluacion.CrearPreguntasCSV(c.Request.Context(), bytes.NewReader(contenido))
	if err != nil {
		middleware.Fail(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"creadas": n})
}

// errorDeArchivo traduce un fallo al leer el cuerpo: demasiado grande es 422; otro, 400.
func errorDeArchivo(err error) error {
	var maxErr *http.MaxBytesError
	if errors.As(err, &maxErr) {
		return domain.Unprocessable("El archivo supera el tamaño permitido", map[string]string{"archivo": "El archivo no puede pasar de 1 MB"})
	}
	return domain.Validation("No se pudo leer el archivo", map[string]string{"archivo": "No se pudo leer el archivo"})
}

// ListarPreguntas atiende GET /admin/evaluacion/preguntas?buscar=&pagina=&por_pagina=.
func (h *EvaluacionAdminHandler) ListarPreguntas(c *gin.Context) {
	pagina, ok := enteroDeConsulta(c, "pagina")
	if !ok {
		return
	}
	porPagina, ok := enteroDeConsulta(c, "por_pagina")
	if !ok {
		return
	}
	r, err := h.evaluacion.ListarPreguntas(c.Request.Context(), c.Query("buscar"), pagina, porPagina)
	if err != nil {
		middleware.Fail(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": r.Items, "total": r.Total, "pagina": r.Pagina, "por_pagina": r.PorPagina})
}

// enteroDeConsulta lee un parámetro entero opcional; uno que no es número responde 400.
func enteroDeConsulta(c *gin.Context, nombre string) (int, bool) {
	texto := c.Query(nombre)
	if texto == "" {
		return 0, true
	}
	n, err := strconv.Atoi(texto)
	if err != nil {
		middleware.Fail(c, domain.Validation("Datos inválidos", map[string]string{nombre: "Debe ser un número entero"}))
		return 0, false
	}
	return n, true
}

// ActualizarPregunta atiende PUT /admin/evaluacion/preguntas/{id}.
func (h *EvaluacionAdminHandler) ActualizarPregunta(c *gin.Context) {
	var req preguntaRequest
	if !middleware.BindJSON(c, &req) {
		return
	}
	p, err := h.evaluacion.ActualizarPregunta(c.Request.Context(), c.Param("id"), req.entrada())
	if err != nil {
		middleware.Fail(c, err)
		return
	}
	c.JSON(http.StatusOK, p)
}

// EliminarPregunta atiende DELETE /admin/evaluacion/preguntas/{id}.
func (h *EvaluacionAdminHandler) EliminarPregunta(c *gin.Context) {
	if err := h.evaluacion.EliminarPregunta(c.Request.Context(), c.Param("id")); err != nil {
		middleware.Fail(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// ObtenerConfig atiende GET /admin/evaluacion/config.
func (h *EvaluacionAdminHandler) ObtenerConfig(c *gin.Context) {
	cfg, err := h.evaluacion.ObtenerConfig(c.Request.Context())
	if err != nil {
		middleware.Fail(c, err)
		return
	}
	c.JSON(http.StatusOK, respuestaConfig(cfg))
}

// GuardarConfig atiende PUT /admin/evaluacion/config.
func (h *EvaluacionAdminHandler) GuardarConfig(c *gin.Context) {
	var req configRequest
	if !middleware.BindJSON(c, &req) {
		return
	}
	cfg, err := h.evaluacion.GuardarConfig(c.Request.Context(), service.ConfigInput{
		IniciaEn: req.IniciaEn, DuracionMinutos: req.DuracionMinutos, NumPreguntas: req.NumPreguntas,
	})
	if err != nil {
		middleware.Fail(c, err)
		return
	}
	c.JSON(http.StatusOK, respuestaConfig(cfg))
}
