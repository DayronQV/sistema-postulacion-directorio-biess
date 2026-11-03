package service

import (
	"bytes"
	"context"
	"strings"
	"time"

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
	fases      domain.Fases // opcional: ver ConFases
	ahora      func() time.Time
}

func NewDocumentoService(documentos domain.Documentos) *DocumentoService {
	return &DocumentoService{documentos: documentos, ahora: time.Now}
}

// ConFases indica las fases del concurso: subir o eliminar un documento solo se permite con
// la fase Postulación publicada, con el formulario habilitado y dentro de sus fechas, igual
// que guardar o enviar la postulación (EF-13). Sin fases no hay restricción.
func (s *DocumentoService) ConFases(f domain.Fases) *DocumentoService {
	s.fases = f
	return s
}

// ConReloj reemplaza el reloj (para las pruebas).
func (s *DocumentoService) ConReloj(ahora func() time.Time) *DocumentoService {
	s.ahora = ahora
	return s
}

// verificarPeriodo responde 403 si hoy no se puede modificar la postulación, y con ella sus
// documentos.
func (s *DocumentoService) verificarPeriodo(ctx context.Context) error {
	abierto, err := formularioAbierto(ctx, s.fases, s.ahora())
	if err != nil {
		return err
	}
	if !abierto {
		return domain.Forbidden(domain.MsgPeriodoCerrado)
	}
	return nil
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
	if err := s.verificarPeriodo(ctx); err != nil {
		return domain.Documento{}, err
	}
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
	if err := s.verificarPeriodo(ctx); err != nil {
		return err
	}
	return s.documentos.EliminarPropioNoReferenciado(ctx, postulanteID, documentoID)
}
