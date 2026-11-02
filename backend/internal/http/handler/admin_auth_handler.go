package handler

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/domain"
	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/http/middleware"
	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/service"
)

// AdminLoginUseCase es lo que el handler necesita del servicio de login del panel.
type AdminLoginUseCase interface {
	Login(ctx context.Context, in service.AdminLoginInput) (service.AdminSesion, error)
}

// AdminAuthHandler atiende el inicio de sesión del panel de administración.
type AdminAuthHandler struct {
	login AdminLoginUseCase
}

// NewAdminAuthHandler crea el handler.
func NewAdminAuthHandler(login AdminLoginUseCase) *AdminAuthHandler {
	return &AdminAuthHandler{login: login}
}

type adminLoginRequest struct {
	Usuario  string `json:"usuario" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type adminSesionResponse struct {
	Token     string     `json:"token"`
	ExpiresIn int        `json:"expires_in"` // segundos
	Usuario   string     `json:"usuario"`
	Rol       domain.Rol `json:"rol"`
}

// Login atiende POST /admin/auth/login.
func (h *AdminAuthHandler) Login(c *gin.Context) {
	var req adminLoginRequest
	if !middleware.BindJSON(c, &req) {
		return
	}
	s, err := h.login.Login(c.Request.Context(), service.AdminLoginInput{Usuario: req.Usuario, Password: req.Password})
	if err != nil {
		middleware.Fail(c, err)
		return
	}
	c.JSON(http.StatusOK, adminSesionResponse{
		Token:     s.Token,
		ExpiresIn: int(s.ExpiraEn.Seconds()),
		Usuario:   s.Usuario,
		Rol:       s.Rol,
	})
}

// Sesion atiende GET /admin/sesion: el usuario y el rol del token. Sirve al panel para
// confirmar que la sesión sigue siendo válida.
func (h *AdminAuthHandler) Sesion(c *gin.Context) {
	claims, _ := middleware.ClaimsFrom(c)
	c.JSON(http.StatusOK, gin.H{"usuario": claims.Usuario, "rol": claims.Rol})
}
