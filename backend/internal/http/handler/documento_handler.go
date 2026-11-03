package handler

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"path"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/domain"
	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/http/middleware"
)

// maximoCuerpoMultipart acota el cuerpo de la subida antes de leerlo: el mayor PDF
// admitido (4 MB de organización) más el margen del formulario multipart. Los límites
// por tipo y el total de 50 MB los aplica el servicio.
const maximoCuerpoMultipart = 5 * 1024 * 1024

// DocumentoUseCase es lo que el handler necesita del servicio de documentos. Se define
// aquí para poder probar el handler con un servicio falso.
type DocumentoUseCase interface {
	Subir(ctx context.Context, postulanteID, tipo, nombre string, contenido []byte) (domain.Documento, error)
	Descargar(ctx context.Context, postulanteID, documentoID string) (domain.Documento, error)
	Eliminar(ctx context.Context, postulanteID, documentoID string) error
}

// DocumentoHandler atiende las rutas /postulacion/documentos.
type DocumentoHandler struct {
	documentos DocumentoUseCase
}

// NewDocumentoHandler crea el handler de documentos.
func NewDocumentoHandler(documentos DocumentoUseCase) *DocumentoHandler {
	return &DocumentoHandler{documentos: documentos}
}

// usuarioID devuelve el postulante autenticado; si no hay sesión responde 401.
func usuarioID(c *gin.Context) (string, bool) {
	claims, ok := middleware.ClaimsFrom(c)
	if !ok {
		middleware.Fail(c, domain.Unauthorized("Debe iniciar sesión para continuar"))
		return "", false
	}
	return claims.UsuarioID, true
}

// Subir atiende POST /postulacion/documentos (multipart con archivo y tipo).
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

// Descargar atiende GET /postulacion/documentos/:id. Un documento ajeno responde 404,
// igual que uno inexistente, para no revelar que existe.
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
	c.Header("Content-Disposition", disposicionAdjunto(d.Nombre))
	c.Data(http.StatusOK, "application/pdf", d.Contenido)
}

// disposicionAdjunto arma Content-Disposition con el nombre bien codificado: los nombres
// con tildes o ñ usan filename* (RFC 2231) y las comillas no rompen la cabecera.
func disposicionAdjunto(nombre string) string {
	if v := mime.FormatMediaType("attachment", map[string]string{"filename": nombre}); v != "" {
		return v
	}
	return `attachment; filename="documento.pdf"`
}

// Eliminar atiende DELETE /postulacion/documentos/:id: 204 si se borró, 409 si una
// sección lo usa y 404 si no es del postulante.
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
