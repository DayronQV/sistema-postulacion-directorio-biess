// Package handler contiene los handlers HTTP: leen la petición, llaman al servicio
// y escriben la respuesta. No contienen lógica de negocio ni SQL.
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// HealthHandler responde al chequeo de vida del servicio.
type HealthHandler struct{}

// NewHealthHandler crea el handler de /health.
func NewHealthHandler() *HealthHandler { return &HealthHandler{} }

// Get responde 200 {"status":"ok"} mientras el proceso esté vivo.
func (h *HealthHandler) Get(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
