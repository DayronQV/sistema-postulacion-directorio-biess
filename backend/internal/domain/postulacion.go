package domain

import (
	"context"
	"time"
)

// Estados de una postulación (columna postulaciones.estado).
const (
	EstadoBorrador = "borrador"
	EstadoEnviada  = "enviada"
	EstadoCerrada  = "cerrada"
)

// NacionalidadEcuatoriana es la única nacionalidad que puede postularse
// (docs/api.md: PUT /postulacion responde 422 si es otra).
const NacionalidadEcuatoriana = "Ecuatoriana"

// Cargos a los que se puede postular (postulaciones.cargo).
const (
	CargoAfiliados = "afiliados"
	CargoJubilados = "jubilados"
	CargoEmpleador = "empleador"
)

// CargoValido indica si cargo es uno de los tres cargos del concurso.
func CargoValido(cargo string) bool {
	return cargo == CargoAfiliados || cargo == CargoJubilados || cargo == CargoEmpleador
}

// InformacionPersonal es la información copiada del Registro Civil al crear la
// postulación. El postulante no la edita.
type InformacionPersonal struct {
	Cedula          string
	Email           string // correo del registro (correo 1 del domicilio); solo lectura
	Nombres         string
	Apellidos       string
	Genero          string
	FechaNacimiento time.Time
	Nacionalidad    string
	EstadoCivil     EstadoCivil
	Pareja          *Pareja
}

// Domicilio es la información domiciliaria y de contacto. Todos los campos son
// opcionales mientras la postulación es un borrador.
type Domicilio struct {
	PaisID            *int
	ProvinciaID       *int
	CantonID          *int
	ProvinciaTexto    string
	CiudadTexto       string
	TelefonoCelular   string
	EmailSecundario   string
	CallePrincipal    string
	Numero            string
	CalleSecundaria   string
	Sector            string
	Referencia        string
	TelefonoDomicilio string
}

// RespaldoOrganizacion es la organización que respalda al postulante y su documento.
type RespaldoOrganizacion struct {
	Nombre      string
	DocumentoID *string
}

// Familiar es un familiar que trabaja en el IESS o el BIESS.
type Familiar struct {
	NombresApellidos string
	Parentesco       string
	Institucion      string // "IESS" o "BIESS"
	Area             string
	Cargo            string
}

// ReferenciaPersonal es una persona de contacto declarada por el postulante.
type ReferenciaPersonal struct {
	NombresApellidos string
	Relacion         string
	Telefono         string
	Correo           string
}

// Formulario es el estado completo del formulario que se guarda con PUT
// /postulacion. Contiene los campos de la propia postulación y una lista por cada
// sección con filas. Cada issue de sección agrega aquí su lista y su repositorio
// (SeccionPostulacion); la transacción no cambia.
type Formulario struct {
	Domicilio            Domicilio
	TieneFamiliares      bool
	Familiares           []Familiar
	Referencias          []ReferenciaPersonal
	Formacion            []FormacionAcademica
	Experiencia          []ExperienciaProfesional
	DuracionExperiencia  Duracion
	Publicaciones        []PublicacionAcademica
	Capacitaciones       []Capacitacion
	Inhabilidades        Inhabilidades
	Elegible             bool // lo calcula el servidor a partir de Inhabilidades
	Cargo                string // vacío mientras no lo haya elegido
	RespaldoOrganizacion RespaldoOrganizacion
	TienePublicaciones   bool
	Consentimiento       bool
}

// Postulacion es la postulación de un postulante con su estado actual.
type Postulacion struct {
	ID               string
	Codigo           string
	PostulanteID     string
	Estado           string
	Personal         InformacionPersonal
	Formulario       Formulario
	ConsentimientoAt *time.Time
	CreadaEn         time.Time
	ActualizadaEn    time.Time
	EnviadaEn        *time.Time
	SuperoMerito     bool // lo marca el administrador (EF-13); habilita la evaluación (EF-14)
}

// Tx representa una transacción abierta. Es opaca para el dominio y el servicio: solo
// el repositorio que la creó sabe qué hay dentro (en PostgreSQL, un pgx.Tx), así el
// servicio no depende de pgx.
type Tx any

// TxManager es el puerto de la unidad de trabajo: abre una transacción, ejecuta fn y
// hace COMMIT si fn devuelve nil o ROLLBACK si devuelve un error (o hay un pánico).
// El error de fn se devuelve tal cual, para poder compararlo con errors.Is.
type TxManager interface {
	Ejecutar(ctx context.Context, fn func(ctx context.Context, tx Tx) error) error
}

// Postulaciones es el puerto de la tabla postulaciones. Los métodos con Tx
// participan de la transacción abierta por TxManager.
type Postulaciones interface {
	// CedulaDePostulante devuelve la cédula del postulante, o ErrNotFound.
	CedulaDePostulante(ctx context.Context, postulanteID string) (string, error)
	// Crear inserta la postulación en borrador con un código nuevo (BIESS-AAAA-NNNNNN).
	// Si el postulante ya tiene una, no crea otra y devuelve false.
	Crear(ctx context.Context, postulanteID string, personal InformacionPersonal, anio int) (bool, error)
	// ObtenerPorPostulante devuelve la postulación con su formulario base, o ErrNotFound.
	ObtenerPorPostulante(ctx context.Context, postulanteID string) (Postulacion, error)
	// BloquearPorPostulante lee la postulación dentro de la transacción y bloquea su
	// fila (SELECT ... FOR UPDATE) hasta el COMMIT o ROLLBACK, o devuelve ErrNotFound.
	BloquearPorPostulante(ctx context.Context, tx Tx, postulanteID string) (Postulacion, error)
	// Actualizar escribe los campos de la postulación (sin las secciones con filas).
	Actualizar(ctx context.Context, tx Tx, postulacionID string, f Formulario) error
	// Cerrar pasa la postulación a estado "cerrada".
	Cerrar(ctx context.Context, tx Tx, postulacionID string) error
	// MarcarEnviada pasa la postulación a estado "enviada" y guarda enviada_at.
	MarcarEnviada(ctx context.Context, tx Tx, postulacionID string) error
}

// SeccionPostulacion es el puerto de una sección con filas (familiares, formación,
// experiencia, ...). Cada sección lo implementa en su propio archivo del repositorio y
// se agrega a la lista que recibe PostulacionService: la transacción las recorre sin
// conocer ninguna en particular.
type SeccionPostulacion interface {
	// Cargar completa en f las filas de la sección, leyéndolas fuera de transacción.
	Cargar(ctx context.Context, postulacionID string, f *Formulario) error
	// Reemplazar borra las filas de la sección y escribe las de f, dentro de la
	// transacción tx. Cualquier error provoca el ROLLBACK de todo el guardado.
	Reemplazar(ctx context.Context, tx Tx, postulacionID string, f Formulario) error
}
