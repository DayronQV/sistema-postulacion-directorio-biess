package service

import (
	"context"
	"regexp"
	"strconv"
	"strings"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/domain"
)

// MsgMeritoSoloEnviada es el 409 de marcar el mérito de una postulación no enviada.
const MsgMeritoSoloEnviada = "Solo se puede marcar la Fase de Méritos de una postulación enviada"

// Un id de postulación es un UUID; con otro formato no hay nada que buscar (404).
var uuidRe = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

// ResultadoMerito es la postulación después de marcar su mérito.
type ResultadoMerito struct {
	ID            string
	SuperoMerito  bool
	EstadoVisible string // lo que ve el postulante: "Evaluación Técnica" si superó
}

const postulacionesAdminPorPagina = 20

// FiltroPostulacionesInput conserva los valores de la URL para validarlos en el servicio.
type FiltroPostulacionesInput struct {
	Cargo    string
	Estado   string
	Elegible string
	Pagina   int
}

// PaginaPostulaciones es una página de la nómina administrativa.
type PaginaPostulaciones struct {
	Items     []domain.ResumenPostulacionAdmin
	Total     int
	Pagina    int
	PorPagina int
}

// AdminPostulacionService reúne las acciones del administrador sobre las postulaciones.
type AdminPostulacionService struct {
	meritos       domain.Meritos
	postulaciones domain.PostulacionesAdmin
	documentos    domain.DocumentosAdmin
	secciones     []domain.SeccionPostulacion
}

// ConConsultas conecta la nómina, el detalle y las descargas del panel.
func (s *AdminPostulacionService) ConConsultas(postulaciones domain.PostulacionesAdmin, documentos domain.DocumentosAdmin, secciones ...domain.SeccionPostulacion) *AdminPostulacionService {
	s.postulaciones = postulaciones
	s.documentos = documentos
	s.secciones = secciones
	return s
}

// Listar devuelve la nómina paginada después de validar los filtros.
func (s *AdminPostulacionService) Listar(ctx context.Context, in FiltroPostulacionesInput) (PaginaPostulaciones, error) {
	f := domain.FiltroPostulacionesAdmin{
		Cargo:  strings.ToLower(strings.TrimSpace(in.Cargo)),
		Estado: strings.ToLower(strings.TrimSpace(in.Estado)),
		Pagina: in.Pagina, PorPagina: postulacionesAdminPorPagina,
	}
	if f.Pagina == 0 {
		f.Pagina = 1
	}
	campos := map[string]string{}
	if f.Cargo != "" && !domain.CargoValido(f.Cargo) {
		campos["cargo"] = "Cargo no válido"
	}
	if f.Estado != "" && f.Estado != domain.EstadoBorrador && f.Estado != domain.EstadoEnviada && f.Estado != domain.EstadoCerrada {
		campos["estado"] = "Estado no válido"
	}
	if f.Pagina < 1 {
		campos["pagina"] = "Debe ser mayor o igual a 1"
	}
	elegible := strings.ToLower(strings.TrimSpace(in.Elegible))
	if elegible != "" {
		valor, err := strconv.ParseBool(elegible)
		if err != nil || (elegible != "true" && elegible != "false") {
			campos["elegible"] = "Debe ser true o false"
		} else {
			f.Elegible = &valor
		}
	}
	if len(campos) > 0 {
		return PaginaPostulaciones{}, domain.Validation("Filtros inválidos", campos)
	}
	items, total, err := s.postulaciones.ListarAdmin(ctx, f)
	if err != nil {
		return PaginaPostulaciones{}, err
	}
	return PaginaPostulaciones{Items: items, Total: total, Pagina: f.Pagina, PorPagina: f.PorPagina}, nil
}

// Obtener carga el formulario completo y calcula la duración de sus experiencias.
func (s *AdminPostulacionService) Obtener(ctx context.Context, postulacionID string) (domain.Postulacion, error) {
	id := strings.TrimSpace(postulacionID)
	if !uuidRe.MatchString(id) {
		return domain.Postulacion{}, domain.NotFound("La postulación no existe")
	}
	p, err := s.postulaciones.ObtenerPorIDAdmin(ctx, id)
	if err != nil {
		return domain.Postulacion{}, err
	}
	for _, seccion := range s.secciones {
		if err := seccion.Cargar(ctx, p.ID, &p.Formulario); err != nil {
			return domain.Postulacion{}, err
		}
	}
	completarDuraciones(&p.Formulario)
	return p, nil
}

// DescargarDocumento obtiene un certificado del expediente indicado.
func (s *AdminPostulacionService) DescargarDocumento(ctx context.Context, postulacionID, documentoID string) (domain.Documento, error) {
	postulacionID = strings.TrimSpace(postulacionID)
	documentoID = strings.TrimSpace(documentoID)
	if !uuidRe.MatchString(postulacionID) || !uuidRe.MatchString(documentoID) {
		return domain.Documento{}, domain.NotFound("Documento no encontrado")
	}
	return s.documentos.ObtenerDePostulacion(ctx, postulacionID, documentoID)
}

// NewAdminPostulacionService crea el servicio con sus dependencias inyectadas.
func NewAdminPostulacionService(meritos domain.Meritos) *AdminPostulacionService {
	return &AdminPostulacionService{meritos: meritos}
}

// MarcarMerito registra si el postulante superó la Fase de Méritos (EF-13). Solo se marca
// una postulación enviada: 404 si no existe y 409 si está en borrador o cerrada. Que solo
// el administrador pueda hacerlo lo controla la ruta (grupo de escritura del panel).
func (s *AdminPostulacionService) MarcarMerito(ctx context.Context, postulacionID string, supero bool) (ResultadoMerito, error) {
	id := strings.TrimSpace(postulacionID)
	if !uuidRe.MatchString(id) {
		return ResultadoMerito{}, domain.NotFound("La postulación no existe")
	}
	estado, err := s.meritos.EstadoPorID(ctx, id)
	if err != nil {
		return ResultadoMerito{}, err
	}
	if estado != domain.EstadoEnviada {
		return ResultadoMerito{}, domain.Conflict(MsgMeritoSoloEnviada, nil)
	}
	actualizada, err := s.meritos.ActualizarMerito(ctx, id, supero)
	if err != nil {
		return ResultadoMerito{}, err
	}
	if !actualizada {
		// Cambió de estado entre la lectura y la escritura.
		return ResultadoMerito{}, domain.Conflict(MsgMeritoSoloEnviada, nil)
	}
	return ResultadoMerito{
		ID:            id,
		SuperoMerito:  supero,
		EstadoVisible: domain.EstadoVisible(domain.EstadoEnviada, supero),
	}, nil
}
