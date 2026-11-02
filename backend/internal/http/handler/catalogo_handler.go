package handler

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/domain"
	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/http/middleware"
)

// CatalogoUseCase es lo que el handler necesita del servicio de catálogos. Se define
// aquí para poder probar el handler con un servicio falso.
type CatalogoUseCase interface {
	Paises(ctx context.Context) ([]domain.ItemCatalogo, error)
	Provincias(ctx context.Context, paisID int) ([]domain.ItemCatalogo, error)
	Cantones(ctx context.Context, provinciaID int) ([]domain.ItemCatalogo, error)
}

// CatalogoHandler atiende las rutas /catalogos.
type CatalogoHandler struct {
	catalogos CatalogoUseCase
}

// NewCatalogoHandler crea el handler de catálogos.
func NewCatalogoHandler(catalogos CatalogoUseCase) *CatalogoHandler {
	return &CatalogoHandler{catalogos: catalogos}
}

// catalogoCacheControl iguala el TTL de la caché de Redis (1 h) para que el navegador
// no vuelva a pedir las listas en cada apertura de la cascada.
const catalogoCacheControl = "public, max-age=3600"

type itemCatalogoResponse struct {
	ID     int    `json:"id"`
	Nombre string `json:"nombre"`
}

// Paises atiende GET /catalogos/paises.
func (h *CatalogoHandler) Paises(c *gin.Context) {
	items, err := h.catalogos.Paises(c.Request.Context())
	respondItems(c, items, err)
}

// Provincias atiende GET /catalogos/paises/:id/provincias.
func (h *CatalogoHandler) Provincias(c *gin.Context) {
	id, ok := pathID(c)
	if !ok {
		return
	}
	items, err := h.catalogos.Provincias(c.Request.Context(), id)
	respondItems(c, items, err)
}

// Cantones atiende GET /catalogos/provincias/:id/cantones.
func (h *CatalogoHandler) Cantones(c *gin.Context) {
	id, ok := pathID(c)
	if !ok {
		return
	}
	items, err := h.catalogos.Cantones(c.Request.Context(), id)
	respondItems(c, items, err)
}

// pathID lee el id de la ruta. Si no es un entero válido responde 400 y devuelve false.
func pathID(c *gin.Context) (int, bool) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 32)
	if err != nil {
		middleware.Fail(c, domain.Validation("Datos inválidos", map[string]string{
			"id": "El identificador debe ser un número entero",
		}))
		return 0, false
	}
	return int(id), true
}

// respondItems escribe la lista como arreglo JSON (vacío, nunca null) o el error.
func respondItems(c *gin.Context, items []domain.ItemCatalogo, err error) {
	if err != nil {
		middleware.Fail(c, err)
		return
	}
	out := make([]itemCatalogoResponse, len(items))
	for i, it := range items {
		out[i] = itemCatalogoResponse{ID: it.ID, Nombre: it.Nombre}
	}
	c.Header("Cache-Control", catalogoCacheControl)
	c.JSON(http.StatusOK, out)
}
