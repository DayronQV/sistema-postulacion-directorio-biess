package handler

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/domain"
	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/http/middleware"
)

const maximoCuerpoMultipart = 5 * 1024 * 1024

type DocumentoUseCase interface {
	Subir(ctx context.Context, postulanteID, tipo, nombre string, contenido []byte) (domain.Documento, error)
	Descargar(ctx context.Context, postulanteID, documentoID string) (domain.Documento, error)
	Eliminar(ctx context.Context, postulanteID, documentoID string) error
}

type DocumentoHandler struct {
	documentos DocumentoUseCase
}

func NewDocumentoHandler(documentos DocumentoUseCase) *DocumentoHandler {
	return &DocumentoHandler{documentos: documentos}
}

func usuarioID(c *gin.Context) (string, bool) {
	claims, ok := middleware.ClaimsFrom(c)
	if !ok {
		middleware.Fail(c, domain.Unauthorized("Debe iniciar sesión para continuar"))
		return "", false
	}
	return claims.UsuarioID, true
}

func (h *DocumentoHandler) Subir(c *gin.Context) {
	id, ok := usuarioID(c)
	if !ok {
		return
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maximoCuerpoMultipart)
	if err := c.Request.ParseMultipartForm(maximoCuerpoMultipart); err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			middleware.Fail(c, domain.Unprocessable("El archivo supera el tamaño permitido", map[string]string{"archivo": "El cuerpo de la subida es demasiado grande"}))
			return
		}
		middleware.Fail(c, domain.Validation("Formulario de archivo inválido", map[string]string{"archivo": "No se pudo leer el archivo"}))
		return
	}
	defer c.Request.MultipartForm.RemoveAll()
	archivo, cabecera, err := c.Request.FormFile("archivo")
	if err != nil {
		middleware.Fail(c, domain.Validation("Datos inválidos", map[string]string{"archivo": "Debe seleccionar un archivo"}))
		return
	}
	defer archivo.Close()
	contenido, err := io.ReadAll(archivo)
	if err != nil {
		middleware.Fail(c, fmt.Errorf("leyendo el documento: %w", err))
		return
	}
	nombre := path.Base(strings.ReplaceAll(cabecera.Filename, "\\", "/"))
	d, err := h.documentos.Subir(c.Request.Context(), id, c.Request.FormValue("tipo"), nombre, contenido)
	if err != nil {
		middleware.Fail(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"documento_id": d.ID, "nombre": d.Nombre, "tamano_bytes": d.Tamano})
}

func (h *DocumentoHandler) Descargar(c *gin.Context) {
	id, ok := usuarioID(c)
	if !ok {
		return
	}
	d, err := h.documentos.Descargar(c.Request.Context(), id, c.Param("id"))
	if err != nil {
		middleware.Fail(c, err)
		return
	}
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, strings.ReplaceAll(d.Nombre, `"`, "")))
	c.Data(http.StatusOK, "application/pdf", d.Contenido)
}

func (h *DocumentoHandler) Eliminar(c *gin.Context) {
	id, ok := usuarioID(c)
	if !ok {
		return
	}
	if err := h.documentos.Eliminar(c.Request.Context(), id, c.Param("id")); err != nil {
		middleware.Fail(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
