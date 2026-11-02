package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/domain"
	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/http/middleware"
)

// PostulacionUseCase es lo que el handler necesita del servicio de postulación. Se
// define aquí para poder probar el handler con un servicio falso.
type PostulacionUseCase interface {
	Obtener(ctx context.Context, postulanteID string) (domain.Postulacion, error)
	Guardar(ctx context.Context, postulanteID string, f domain.Formulario) (domain.Postulacion, error)
	Enviar(ctx context.Context, postulanteID string) (domain.Postulacion, error)
}

// PostulacionHandler atiende las rutas /postulacion. Las rutas pasan por RequireAuth y
// RequireRole(postulante), que guardan la identidad en el contexto.
type PostulacionHandler struct {
	postulacion PostulacionUseCase
}

// NewPostulacionHandler crea el handler de postulación.
func NewPostulacionHandler(postulacion PostulacionUseCase) *PostulacionHandler {
	return &PostulacionHandler{postulacion: postulacion}
}

// postulanteID devuelve el id del postulante autenticado. Si la ruta no pasó por
// RequireAuth responde 401 y devuelve false.
func postulanteID(c *gin.Context) (string, bool) {
	claims, ok := middleware.ClaimsFrom(c)
	if !ok || claims.UsuarioID == "" {
		middleware.Fail(c, domain.Unauthorized("Debe iniciar sesión para continuar"))
		return "", false
	}
	return claims.UsuarioID, true
}

type domicilioDTO struct {
	PaisID            *int   `json:"pais_id"`
	ProvinciaID       *int   `json:"provincia_id"`
	CantonID          *int   `json:"canton_id"`
	ProvinciaTexto    string `json:"provincia_texto"`
	CiudadTexto       string `json:"ciudad_texto"`
	TelefonoCelular   string `json:"telefono_celular"`
	EmailSecundario   string `json:"email_secundario"`
	CallePrincipal    string `json:"calle_principal"`
	Numero            string `json:"numero"`
	CalleSecundaria   string `json:"calle_secundaria"`
	Sector            string `json:"sector"`
	Referencia        string `json:"referencia"`
	TelefonoDomicilio string `json:"telefono_domicilio"`
}

type familiarDTO struct {
	NombresApellidos string `json:"nombres_apellidos"`
	Parentesco       string `json:"parentesco"`
	Institucion      string `json:"institucion"`
	Area             string `json:"area"`
	Cargo            string `json:"cargo"`
}

type referenciaDTO struct {
	NombresApellidos string `json:"nombres_apellidos"`
	Relacion         string `json:"relacion"`
	Telefono         string `json:"telefono"`
	Correo           string `json:"correo"`
}

type formacionDTO struct {
	Nivel          string `json:"nivel"`
	Institucion    string `json:"institucion"`
	Tipo           string `json:"tipo"`
	Titulo         string `json:"titulo"`
	Area           string `json:"area"`
	SenescytNumero string `json:"senescyt_numero"`
	SenescytFecha  string `json:"senescyt_fecha"`
	DocumentoID    string `json:"documento_id"`
}

type duracionDTO struct {
	Anios int `json:"anios"`
	Meses int `json:"meses"`
	Dias  int `json:"dias"`
}

type experienciaDTO struct {
	Tipo        string      `json:"tipo"`
	Cargo       string      `json:"cargo"`
	Institucion string      `json:"institucion"`
	FechaInicio string      `json:"fecha_inicio"`
	FechaFin    string      `json:"fecha_fin"`
	Actividades string      `json:"actividades"`
	Descripcion string      `json:"descripcion"`
	DocumentoID string      `json:"documento_id"`
	Duracion    duracionDTO `json:"duracion"`
}

type publicacionDTO struct {
	Titulo      string `json:"titulo"`
	Tipo        string `json:"tipo"`
	Medio       string `json:"medio"`
	Fecha       string `json:"fecha"`
	Relacionada bool   `json:"relacionada"`
	DocumentoID string `json:"documento_id"`
}

type capacitacionDTO struct {
	Tipo        string `json:"tipo"`
	Nombre      string `json:"nombre"`
	Institucion string `json:"institucion"`
	FechaInicio string `json:"fecha_inicio"`
	FechaFin    string `json:"fecha_fin"`
	Horas       int    `json:"horas"`
	DocumentoID string `json:"documento_id"`
}

type inhabilidadesDTO struct {
	respuestas map[int]bool
	removido   removidoDTO
	invalidos  map[string]string // campos mal escritos al leer el JSON; los reporta Guardar
}

type removidoDTO struct {
	Institucion string `json:"institucion"`
	Cargo       string `json:"cargo"`
	FechaFin    string `json:"fecha_fin"`
	Motivo      string `json:"motivo"`
}

// UnmarshalJSON lee las claves "1" a "13" (respuesta Sí/No) y "removido_detalle". Una clave
// sin respuesta (null) cuenta como pregunta sin responder y cualquier otra clave se ignora.
func (d *inhabilidadesDTO) UnmarshalJSON(datos []byte) error {
	var crudo map[string]json.RawMessage
	if err := json.Unmarshal(datos, &crudo); err != nil {
		return err
	}
	d.respuestas = map[int]bool{}
	for n := 1; n <= domain.TotalPreguntasInhabilidades; n++ {
		valor, ok := crudo[strconv.Itoa(n)]
		if !ok || string(valor) == "null" {
			continue
		}
		var respuesta bool
		if err := json.Unmarshal(valor, &respuesta); err != nil {
			d.marcarInvalido("inhabilidades."+strconv.Itoa(n), "Debe ser true o false")
			continue
		}
		d.respuestas[n] = respuesta
	}
	if valor, ok := crudo["removido_detalle"]; ok && string(valor) != "null" {
		if err := json.Unmarshal(valor, &d.removido); err != nil {
			d.marcarInvalido("inhabilidades.removido_detalle", "Tipo de dato no válido")
		}
	}
	return nil
}

// marcarInvalido anota un campo mal escrito. No se devuelve como error del decodificador
// porque la ruta del campo que arma encoding/json depende de la versión de Go.
func (d *inhabilidadesDTO) marcarInvalido(campo, mensaje string) {
	if d.invalidos == nil {
		d.invalidos = map[string]string{}
	}
	d.invalidos[campo] = mensaje
}

// MarshalJSON escribe solo las preguntas respondidas y, si la 7 es Sí, su detalle.
func (d inhabilidadesDTO) MarshalJSON() ([]byte, error) {
	salida := make(map[string]any, len(d.respuestas)+1)
	for n, respuesta := range d.respuestas {
		salida[strconv.Itoa(n)] = respuesta
	}
	if d.respuestas[domain.PreguntaRemovido] {
		salida["removido_detalle"] = d.removido
	}
	return json.Marshal(salida)
}

type respaldoDTO struct {
	Nombre      string  `json:"nombre"`
	DocumentoID *string `json:"documento_id"`
}

// formularioDTO es el cuerpo de PUT /postulacion y la parte editable de la respuesta.
// Cada issue de sección agrega aquí su campo. Los campos que todavía no existen en este
// DTO (formacion, experiencia, ...) se ignoran.
type formularioDTO struct {
	Domicilio            domicilioDTO      `json:"domicilio"`
	TieneFamiliares      bool              `json:"tiene_familiares"`
	Familiares           []familiarDTO     `json:"familiares"`
	Referencias          []referenciaDTO   `json:"referencias"`
	Formacion            []formacionDTO    `json:"formacion"`
	Experiencia          []experienciaDTO  `json:"experiencia"`
	DuracionExperiencia  duracionDTO       `json:"duracion_experiencia_total"`
	Publicaciones        []publicacionDTO  `json:"publicaciones"`
	Capacitaciones       []capacitacionDTO `json:"capacitaciones"`
	Inhabilidades        inhabilidadesDTO  `json:"inhabilidades"`
	Cargo                string            `json:"cargo"`
	RespaldoOrganizacion respaldoDTO       `json:"respaldo_organizacion"`
	TienePublicaciones   bool              `json:"tiene_publicaciones"`
	Consentimiento       bool              `json:"consentimiento"`
}

type parejaDTO struct {
	NombresApellidos string `json:"nombres_apellidos"`
	Cedula           string `json:"cedula"`
}

type personalDTO struct {
	Cedula          string     `json:"cedula"`
	Email           string     `json:"email"`
	Nombres         string     `json:"nombres"`
	Apellidos       string     `json:"apellidos"`
	Genero          string     `json:"genero"`
	FechaNacimiento string     `json:"fecha_nacimiento"`
	Nacionalidad    string     `json:"nacionalidad"`
	EstadoCivil     string     `json:"estado_civil"`
	Pareja          *parejaDTO `json:"pareja,omitempty"`
}

type postulacionResponse struct {
	ID                  string      `json:"id"`
	Codigo              string      `json:"codigo"`
	Estado              string      `json:"estado"`
	InformacionPersonal personalDTO `json:"informacion_personal"`
	formularioDTO
	Elegible  bool      `json:"elegible"` // solo lectura: lo calcula el servidor
	UpdatedAt time.Time `json:"updated_at"`
}

func (d formularioDTO) dominio() domain.Formulario {
	dom := d.Domicilio
	f := domain.Formulario{
		Domicilio: domain.Domicilio{
			PaisID: dom.PaisID, ProvinciaID: dom.ProvinciaID, CantonID: dom.CantonID,
			ProvinciaTexto: dom.ProvinciaTexto, CiudadTexto: dom.CiudadTexto,
			TelefonoCelular: dom.TelefonoCelular, EmailSecundario: dom.EmailSecundario,
			CallePrincipal: dom.CallePrincipal, Numero: dom.Numero, CalleSecundaria: dom.CalleSecundaria,
			Sector: dom.Sector, Referencia: dom.Referencia, TelefonoDomicilio: dom.TelefonoDomicilio,
		},
		TieneFamiliares:      d.TieneFamiliares,
		Cargo:                d.Cargo,
		RespaldoOrganizacion: domain.RespaldoOrganizacion{Nombre: d.RespaldoOrganizacion.Nombre, DocumentoID: d.RespaldoOrganizacion.DocumentoID},
		TienePublicaciones:   d.TienePublicaciones,
		Consentimiento:       d.Consentimiento,
		Inhabilidades: domain.Inhabilidades{
			Respuestas: d.Inhabilidades.respuestas,
			Removido:   domain.RemovidoDetalle(d.Inhabilidades.removido),
		},
	}
	for _, fam := range d.Familiares {
		f.Familiares = append(f.Familiares, domain.Familiar(fam))
	}
	for _, referencia := range d.Referencias {
		f.Referencias = append(f.Referencias, domain.ReferenciaPersonal(referencia))
	}
	for _, fila := range d.Formacion {
		f.Formacion = append(f.Formacion, domain.FormacionAcademica(fila))
	}
	for _, fila := range d.Experiencia {
		f.Experiencia = append(f.Experiencia, domain.ExperienciaProfesional{
			Tipo: fila.Tipo, Cargo: fila.Cargo, Institucion: fila.Institucion,
			FechaInicio: fila.FechaInicio, FechaFin: fila.FechaFin, Actividades: fila.Actividades,
			Descripcion: fila.Descripcion, DocumentoID: fila.DocumentoID,
		})
	}
	for _, fila := range d.Publicaciones {
		f.Publicaciones = append(f.Publicaciones, domain.PublicacionAcademica(fila))
	}
	for _, fila := range d.Capacitaciones {
		f.Capacitaciones = append(f.Capacitaciones, domain.Capacitacion(fila))
	}
	return f
}

func respuesta(p domain.Postulacion) postulacionResponse {
	f := p.Formulario
	d := f.Domicilio
	r := postulacionResponse{
		ID:     p.ID,
		Codigo: p.Codigo,
		Estado: p.Estado,
		InformacionPersonal: personalDTO{
			Cedula:          p.Personal.Cedula,
			Email:           p.Personal.Email,
			Nombres:         p.Personal.Nombres,
			Apellidos:       p.Personal.Apellidos,
			Genero:          p.Personal.Genero,
			FechaNacimiento: p.Personal.FechaNacimiento.Format(time.DateOnly),
			Nacionalidad:    p.Personal.Nacionalidad,
			EstadoCivil:     string(p.Personal.EstadoCivil),
		},
		formularioDTO: formularioDTO{
			Domicilio: domicilioDTO{
				PaisID: d.PaisID, ProvinciaID: d.ProvinciaID, CantonID: d.CantonID,
				ProvinciaTexto: d.ProvinciaTexto, CiudadTexto: d.CiudadTexto,
				TelefonoCelular: d.TelefonoCelular, EmailSecundario: d.EmailSecundario,
				CallePrincipal: d.CallePrincipal, Numero: d.Numero, CalleSecundaria: d.CalleSecundaria,
				Sector: d.Sector, Referencia: d.Referencia, TelefonoDomicilio: d.TelefonoDomicilio,
			},
			TieneFamiliares:      f.TieneFamiliares,
			Familiares:           []familiarDTO{},
			Referencias:          []referenciaDTO{},
			Formacion:            []formacionDTO{},
			Experiencia:          []experienciaDTO{},
			DuracionExperiencia:  duracionDTO{Anios: f.DuracionExperiencia.Anios, Meses: f.DuracionExperiencia.Meses, Dias: f.DuracionExperiencia.Dias},
			Publicaciones:        []publicacionDTO{},
			Capacitaciones:       []capacitacionDTO{},
			Cargo:                f.Cargo,
			RespaldoOrganizacion: respaldoDTO{Nombre: f.RespaldoOrganizacion.Nombre, DocumentoID: f.RespaldoOrganizacion.DocumentoID},
			TienePublicaciones:   f.TienePublicaciones,
			Consentimiento:       f.Consentimiento,
			Inhabilidades: inhabilidadesDTO{
				respuestas: f.Inhabilidades.Respuestas,
				removido:   removidoDTO(f.Inhabilidades.Removido),
			},
		},
		Elegible:  f.Elegible,
		UpdatedAt: p.ActualizadaEn,
	}
	if pareja := p.Personal.Pareja; pareja != nil {
		r.InformacionPersonal.Pareja = &parejaDTO{NombresApellidos: pareja.NombresApellidos, Cedula: pareja.Cedula}
	}
	for _, fam := range f.Familiares {
		r.Familiares = append(r.Familiares, familiarDTO(fam))
	}
	for _, referencia := range f.Referencias {
		r.Referencias = append(r.Referencias, referenciaDTO(referencia))
	}
	for _, fila := range f.Formacion {
		r.Formacion = append(r.Formacion, formacionDTO(fila))
	}
	for _, fila := range f.Experiencia {
		r.Experiencia = append(r.Experiencia, experienciaDTO{
			Tipo: fila.Tipo, Cargo: fila.Cargo, Institucion: fila.Institucion,
			FechaInicio: fila.FechaInicio, FechaFin: fila.FechaFin, Actividades: fila.Actividades,
			Descripcion: fila.Descripcion, DocumentoID: fila.DocumentoID,
			Duracion: duracionDTO{Anios: fila.Duracion.Anios, Meses: fila.Duracion.Meses, Dias: fila.Duracion.Dias},
		})
	}
	for _, fila := range f.Publicaciones {
		r.Publicaciones = append(r.Publicaciones, publicacionDTO(fila))
	}
	for _, fila := range f.Capacitaciones {
		r.Capacitaciones = append(r.Capacitaciones, capacitacionDTO(fila))
	}
	return r
}

// Obtener atiende GET /postulacion: la postulación del usuario, creada en borrador la
// primera vez.
func (h *PostulacionHandler) Obtener(c *gin.Context) {
	id, ok := postulanteID(c)
	if !ok {
		return
	}
	p, err := h.postulacion.Obtener(c.Request.Context(), id)
	if err != nil {
		middleware.Fail(c, err)
		return
	}
	c.JSON(http.StatusOK, respuesta(p))
}

// Guardar atiende PUT /postulacion: reemplaza el estado completo del formulario en una
// transacción y devuelve la postulación guardada.
func (h *PostulacionHandler) Guardar(c *gin.Context) {
	id, ok := postulanteID(c)
	if !ok {
		return
	}
	var req formularioDTO
	if !middleware.BindJSON(c, &req) {
		return
	}
	if len(req.Inhabilidades.invalidos) > 0 {
		middleware.Fail(c, domain.Validation("Datos inválidos", req.Inhabilidades.invalidos))
		return
	}
	p, err := h.postulacion.Guardar(c.Request.Context(), id, req.dominio())
	if err != nil {
		middleware.Fail(c, err)
		return
	}
	c.JSON(http.StatusOK, respuesta(p))
}

// Enviar atiende POST /postulacion/enviar: valida que la postulación esté completa, la marca
// como enviada y manda el formulario al correo del postulante. Devuelve la postulación.
func (h *PostulacionHandler) Enviar(c *gin.Context) {
	id, ok := postulanteID(c)
	if !ok {
		return
	}
	p, err := h.postulacion.Enviar(c.Request.Context(), id)
	if err != nil {
		middleware.Fail(c, err)
		return
	}
	c.JSON(http.StatusOK, respuesta(p))
}
