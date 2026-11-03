package handler

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/domain"
	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/http/middleware"
	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/service"
)

// MeritoUseCase es lo que el handler necesita para marcar el mérito de una postulación.
type MeritoUseCase interface {
	MarcarMerito(ctx context.Context, postulacionID string, supero bool) (service.ResultadoMerito, error)
}

// ConsultaPostulacionUseCase reúne las lecturas de la nómina administrativa.
type ConsultaPostulacionUseCase interface {
	Listar(ctx context.Context, filtro service.FiltroPostulacionesInput) (service.PaginaPostulaciones, error)
	Obtener(ctx context.Context, postulacionID string) (domain.Postulacion, error)
	DescargarDocumento(ctx context.Context, postulacionID, documentoID string) (domain.Documento, error)
}

// MatrizUseCase genera las exportaciones del cuadro comparativo.
type MatrizUseCase interface {
	Exportar(ctx context.Context, filtro service.FiltroMatrizInput) (service.ArchivoMatriz, error)
}

// AdminPostulacionHandler atiende las acciones del panel sobre las postulaciones.
type AdminPostulacionHandler struct {
	meritos   MeritoUseCase
	consultas ConsultaPostulacionUseCase
	matriz    MatrizUseCase
}

// NewAdminPostulacionHandler crea el handler.
func NewAdminPostulacionHandler(meritos MeritoUseCase) *AdminPostulacionHandler {
	return &AdminPostulacionHandler{meritos: meritos}
}

// ConConsultas conecta las lecturas de la nómina y el expediente completo.
func (h *AdminPostulacionHandler) ConConsultas(consultas ConsultaPostulacionUseCase) *AdminPostulacionHandler {
	h.consultas = consultas
	return h
}

// ConMatriz conecta la generación de exportaciones de la nómina.
func (h *AdminPostulacionHandler) ConMatriz(matriz MatrizUseCase) *AdminPostulacionHandler {
	h.matriz = matriz
	return h
}

type resumenPostulacionResponse struct {
	ID        string `json:"id"`
	Codigo    string `json:"codigo"`
	Cedula    string `json:"cedula"`
	Nombres   string `json:"nombres"`
	Apellidos string `json:"apellidos"`
	Cargo     string `json:"cargo"`
	Estado    string `json:"estado"`
	Elegible  bool   `json:"elegible"`
}

// Listar atiende GET /admin/postulaciones con filtros y paginación.
func (h *AdminPostulacionHandler) Listar(c *gin.Context) {
	pagina, ok := enteroDeConsulta(c, "pagina")
	if !ok {
		return
	}
	r, err := h.consultas.Listar(c.Request.Context(), service.FiltroPostulacionesInput{
		Cargo: c.Query("cargo"), Estado: c.Query("estado"), Elegible: c.Query("elegible"), Pagina: pagina,
	})
	if err != nil {
		middleware.Fail(c, err)
		return
	}
	items := make([]resumenPostulacionResponse, 0, len(r.Items))
	for _, item := range r.Items {
		items = append(items, resumenPostulacionResponse{
			ID: item.ID, Codigo: item.Codigo, Cedula: item.Cedula, Nombres: item.Nombres,
			Apellidos: item.Apellidos, Cargo: item.Cargo, Estado: item.Estado, Elegible: item.Elegible,
		})
	}
	c.JSON(http.StatusOK, gin.H{"items": items, "total": r.Total, "pagina": r.Pagina, "por_pagina": r.PorPagina})
}

// Obtener atiende GET /admin/postulaciones/{id}.
func (h *AdminPostulacionHandler) Obtener(c *gin.Context) {
	p, err := h.consultas.Obtener(c.Request.Context(), c.Param("id"))
	if err != nil {
		middleware.Fail(c, err)
		return
	}
	c.JSON(http.StatusOK, respuesta(p))
}

// DescargarDocumento atiende la descarga de un certificado del expediente.
func (h *AdminPostulacionHandler) DescargarDocumento(c *gin.Context) {
	d, err := h.consultas.DescargarDocumento(c.Request.Context(), c.Param("id"), c.Param("docId"))
	if err != nil {
		middleware.Fail(c, err)
		return
	}
	c.Header("Content-Disposition", disposicionAdjunto(d.Nombre))
	c.Data(http.StatusOK, "application/pdf", d.Contenido)
}

// ExportarMatriz atiende GET /admin/matriz en Excel o PDF.
func (h *AdminPostulacionHandler) ExportarMatriz(c *gin.Context) {
	archivo, err := h.matriz.Exportar(c.Request.Context(), service.FiltroMatrizInput{
		Formato: c.Query("formato"), Cargo: c.Query("cargo"),
		Estado: c.Query("estado"), Elegible: c.Query("elegible"),
	})
	if err != nil {
		middleware.Fail(c, err)
		return
	}
	c.Header("Content-Disposition", disposicionAdjunto(archivo.Nombre))
	c.Data(http.StatusOK, archivo.ContentType, archivo.Contenido)
}

type meritoRequest struct {
	// Puntero para exigir el campo: un false explícito es válido, ausente no.
	Supero *bool `json:"supero" binding:"required"`
}

type meritoResponse struct {
	ID            string `json:"id"`
	SuperoMerito  bool   `json:"supero_merito"`
	EstadoVisible string `json:"estado_visible"`
}

// MarcarMerito atiende PUT /admin/postulaciones/{id}/merito.
func (h *AdminPostulacionHandler) MarcarMerito(c *gin.Context) {
	var req meritoRequest
	if !middleware.BindJSON(c, &req) {
		return
	}
	r, err := h.meritos.MarcarMerito(c.Request.Context(), c.Param("id"), *req.Supero)
	if err != nil {
		middleware.Fail(c, err)
		return
	}
	c.JSON(http.StatusOK, meritoResponse{ID: r.ID, SuperoMerito: r.SuperoMerito, EstadoVisible: r.EstadoVisible})
}

// Rutas registra las rutas de escritura de este módulo en el panel (solo administrador).
func (h *AdminPostulacionHandler) Rutas(lectura, escritura *gin.RouterGroup) {
	lectura.GET("/matriz", h.ExportarMatriz)
	lectura.GET("/postulaciones", h.Listar)
	lectura.GET("/postulaciones/:id/documentos/:docId", h.DescargarDocumento)
	lectura.GET("/postulaciones/:id", h.Obtener)
	escritura.PUT("/postulaciones/:id/merito", h.MarcarMerito)
}
