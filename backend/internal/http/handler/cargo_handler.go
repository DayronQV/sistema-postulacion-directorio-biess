package handler

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/domain"
	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/http/middleware"
)

// CargoUseCase es lo que el handler necesita del servicio para validar el cargo. Se
// define aquí para poder probar el handler con un servicio falso.
type CargoUseCase interface {
	ValidarCargo(ctx context.Context, postulanteID, cargo string) (domain.ResultadoCargo, error)
}

// CargoHandler atiende POST /postulacion/cargo/validar.
type CargoHandler struct {
	cargo CargoUseCase
}

// NewCargoHandler crea el handler de la validación del cargo.
func NewCargoHandler(cargo CargoUseCase) *CargoHandler {
	return &CargoHandler{cargo: cargo}
}

type validarCargoRequest struct {
	Cargo string `json:"cargo" binding:"required"`
}

type validarCargoResponse struct {
	Cargo    string `json:"cargo"`
	Elegible bool   `json:"elegible"`
	Mensaje  string `json:"mensaje"`
}

// Validar consulta el Historial Laboral para el cargo pedido. Si el postulante cumple la
// calidad, responde 200 sin escribir nada; si no, la postulación queda cerrada y responde
// 422 con el mensaje del requerimiento.
func (h *CargoHandler) Validar(c *gin.Context) {
	id, ok := postulanteID(c)
	if !ok {
		return
	}
	var req validarCargoRequest
	if !middleware.BindJSON(c, &req) {
		return
	}
	res, err := h.cargo.ValidarCargo(c.Request.Context(), id, req.Cargo)
	if err != nil {
		middleware.Fail(c, err)
		return
	}
	c.JSON(http.StatusOK, validarCargoResponse{Cargo: res.Cargo, Elegible: res.Elegible, Mensaje: res.Mensaje})
}
