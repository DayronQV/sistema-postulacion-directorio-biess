package handler

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/http/middleware"
	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/service"
)

// RegistroUseCase es lo que el handler necesita del servicio de registro. Se define
// aquí para poder probar el handler con un servicio falso.
type RegistroUseCase interface {
	ValidarIdentidad(ctx context.Context, in service.IdentidadInput) (string, error)
	Registrar(ctx context.Context, in service.RegistroInput) (string, error)
}

// AuthHandler atiende las rutas /auth.
type AuthHandler struct {
	registro RegistroUseCase
}

// NewAuthHandler crea el handler de autenticación.
func NewAuthHandler(registro RegistroUseCase) *AuthHandler {
	return &AuthHandler{registro: registro}
}

type identidadRequest struct {
	Cedula          string `json:"cedula" binding:"required"`
	FechaExpedicion string `json:"fecha_expedicion"`
	CodigoDactilar  string `json:"codigo_dactilar"`
}

type identidadResponse struct {
	Ticket string `json:"ticket"`
}

// Identidad atiende POST /auth/identidad: primer paso del registro.
func (h *AuthHandler) Identidad(c *gin.Context) {
	var req identidadRequest
	if !middleware.BindJSON(c, &req) {
		return
	}
	ticket, err := h.registro.ValidarIdentidad(c.Request.Context(), service.IdentidadInput{
		Cedula:          req.Cedula,
		FechaExpedicion: req.FechaExpedicion,
		CodigoDactilar:  req.CodigoDactilar,
	})
	if err != nil {
		middleware.Fail(c, err)
		return
	}
	c.JSON(http.StatusOK, identidadResponse{Ticket: ticket})
}

type registroRequest struct {
	Ticket   string `json:"ticket" binding:"required"`
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type registroResponse struct {
	ID string `json:"id"`
}

// Registro atiende POST /auth/registro: segundo paso del registro.
func (h *AuthHandler) Registro(c *gin.Context) {
	var req registroRequest
	if !middleware.BindJSON(c, &req) {
		return
	}
	id, err := h.registro.Registrar(c.Request.Context(), service.RegistroInput{
		Ticket:   req.Ticket,
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		middleware.Fail(c, err)
		return
	}
	c.JSON(http.StatusCreated, registroResponse{ID: id})
}
