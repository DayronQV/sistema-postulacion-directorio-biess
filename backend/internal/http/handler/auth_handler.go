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

// LoginUseCase es lo que el handler necesita del servicio de inicio de sesión.
type LoginUseCase interface {
	Login(ctx context.Context, in service.LoginInput) (service.LoginResult, error)
}

// RecuperacionUseCase es lo que el handler necesita del servicio de recuperación de
// contraseña.
type RecuperacionUseCase interface {
	Recuperar(ctx context.Context, in service.IdentidadInput) error
	Restablecer(ctx context.Context, token, password string) error
}

// AuthHandler atiende las rutas /auth.
type AuthHandler struct {
	registro     RegistroUseCase
	login        LoginUseCase
	recuperacion RecuperacionUseCase
}

// NewAuthHandler crea el handler de autenticación.
func NewAuthHandler(registro RegistroUseCase, login LoginUseCase, recuperacion RecuperacionUseCase) *AuthHandler {
	return &AuthHandler{registro: registro, login: login, recuperacion: recuperacion}
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

type loginRequest struct {
	Cedula   string `json:"cedula" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type loginResponse struct {
	Token         string `json:"token"`
	ExpiresIn     int    `json:"expires_in"` // segundos
	PrimerIngreso bool   `json:"primer_ingreso"`
}

// Login atiende POST /auth/login: inicio de sesión del postulante.
func (h *AuthHandler) Login(c *gin.Context) {
	var req loginRequest
	if !middleware.BindJSON(c, &req) {
		return
	}
	res, err := h.login.Login(c.Request.Context(), service.LoginInput{
		Cedula:   req.Cedula,
		Password: req.Password,
	})
	if err != nil {
		middleware.Fail(c, err)
		return
	}
	c.JSON(http.StatusOK, loginResponse{
		Token:         res.Token,
		ExpiresIn:     int(res.ExpiraEn.Seconds()),
		PrimerIngreso: res.PrimerIngreso,
	})
}

// Recuperar atiende POST /auth/recuperar: valida la identidad y envía el enlace para
// restablecer la contraseña. Responde 202 sin cuerpo tenga o no cuenta la cédula.
func (h *AuthHandler) Recuperar(c *gin.Context) {
	var req identidadRequest
	if !middleware.BindJSON(c, &req) {
		return
	}
	err := h.recuperacion.Recuperar(c.Request.Context(), service.IdentidadInput{
		Cedula:          req.Cedula,
		FechaExpedicion: req.FechaExpedicion,
		CodigoDactilar:  req.CodigoDactilar,
	})
	if err != nil {
		middleware.Fail(c, err)
		return
	}
	c.Status(http.StatusAccepted)
}

type restablecerRequest struct {
	Token    string `json:"token" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// Restablecer atiende POST /auth/restablecer: cambia la contraseña con el token del
// enlace. Responde 204 sin cuerpo.
func (h *AuthHandler) Restablecer(c *gin.Context) {
	var req restablecerRequest
	if !middleware.BindJSON(c, &req) {
		return
	}
	if err := h.recuperacion.Restablecer(c.Request.Context(), req.Token, req.Password); err != nil {
		middleware.Fail(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
