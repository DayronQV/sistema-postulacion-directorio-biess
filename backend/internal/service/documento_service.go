package service

import (
	"bytes"
	"context"
	"strings"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/domain"
)

const (
	megabyte              int64 = 1024 * 1024
	limiteDocumentoComun        = 2 * megabyte
	limiteOrganizacion          = 4 * megabyte
	limiteTotalDocumentos       = 50 * megabyte
)

// DocumentoService aplica las reglas de los archivos antes de persistirlos.
type DocumentoService struct {
	documentos domain.Documentos
}

func NewDocumentoService(documentos domain.Documentos) *DocumentoService {
	return &DocumentoService{documentos: documentos}
}

func limitePorTipo(tipo string) (int64, bool) {
	switch tipo {
	case domain.DocumentoOrganizacion:
		return limiteOrganizacion, true
	case domain.DocumentoFormacion, domain.DocumentoExperiencia, domain.DocumentoPublicacion, domain.DocumentoCapacitacion:
		return limiteDocumentoComun, true
	default:
		return 0, false
	}
}

// Subir valida el tipo, el contenido PDF y los límites por archivo y postulante.
func (s *DocumentoService) Subir(ctx context.Context, postulanteID, tipo, nombre string, contenido []byte) (domain.Documento, error) {
	tipo = strings.TrimSpace(strings.ToLower(tipo))
	limite, valido := limitePorTipo(tipo)
	if !valido {
		return domain.Documento{}, domain.Unprocessable("Tipo de documento no válido", map[string]string{"tipo": "Tipo de documento no válido"})
	}
	if len(contenido) == 0 {
		return domain.Documento{}, domain.Unprocessable("El archivo PDF está vacío", map[string]string{"archivo": "El archivo está vacío"})
	}
	if !bytes.HasPrefix(contenido, []byte("%PDF")) {
		return domain.Documento{}, domain.Unprocessable("El archivo no contiene un PDF válido", map[string]string{"archivo": "El contenido no es PDF"})
	}
	if int64(len(contenido)) > limite {
		return domain.Documento{}, domain.Unprocessable("El archivo supera el tamaño permitido", map[string]string{"archivo": "El tamaño máximo para este tipo es de " + tamanoEnMB(limite)})
	}
	if strings.TrimSpace(nombre) == "" {
		nombre = "documento.pdf"
	}
	return s.documentos.Guardar(ctx, postulanteID, domain.Documento{
		Tipo: tipo, Nombre: nombre, Tamano: int64(len(contenido)), Contenido: contenido,
	}, limiteTotalDocumentos)
}

func tamanoEnMB(bytes int64) string {
	if bytes == limiteOrganizacion {
		return "4 MB"
	}
	return "2 MB"
}

func (s *DocumentoService) Descargar(ctx context.Context, postulanteID, documentoID string) (domain.Documento, error) {
	return s.documentos.ObtenerPropio(ctx, postulanteID, documentoID)
}

func (s *DocumentoService) Eliminar(ctx context.Context, postulanteID, documentoID string) error {
	return s.documentos.EliminarPropioNoReferenciado(ctx, postulanteID, documentoID)
}
