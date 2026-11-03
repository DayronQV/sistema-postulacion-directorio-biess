package domain

import (
	"context"
	"time"
)

// EstadoCivil es el estado civil que informa el Registro Civil.
type EstadoCivil string

const (
	EstadoSoltero      EstadoCivil = "Soltero/a"
	EstadoCasado       EstadoCivil = "Casado/a"
	EstadoUnionDeHecho EstadoCivil = "Unión de hecho"
	EstadoDivorciado   EstadoCivil = "Divorciado/a"
	EstadoViudo        EstadoCivil = "Viudo/a"
)

// ConPareja indica si el estado civil exige mostrar los datos de la pareja
// (casado/a o unión de hecho, según docs/especificacion.md).
func (e EstadoCivil) ConPareja() bool {
	return e == EstadoCasado || e == EstadoUnionDeHecho
}

// Pareja son los datos del cónyuge o conviviente.
type Pareja struct {
	NombresApellidos string
	Cedula           string
}

// Ciudadano es la información que devuelve el Registro Civil para una cédula.
type Ciudadano struct {
	Cedula          string
	Nombres         string
	Apellidos       string
	Genero          string
	FechaNacimiento time.Time
	Nacionalidad    string
	EstadoCivil     EstadoCivil
	Pareja          *Pareja // solo si EstadoCivil.ConPareja()

	// Datos con los que se comprueba la identidad en el registro y en la
	// recuperación de contraseña: basta con acertar uno de los dos.
	FechaExpedicion time.Time
	CodigoDactilar  string // 10 caracteres alfanuméricos en mayúscula
}

// RegistroCivil es el puerto hacia el Registro Civil del Ecuador. Hoy se simula con
// un adaptador mock (internal/integration/registrocivil); si algún día hubiera
// acceso real, solo se cambia el adaptador.
type RegistroCivil interface {
	// ConsultarCiudadano devuelve los datos de la cédula, o un error que cumple
	// errors.Is(err, ErrNotFound) si no existe.
	//
	// Seguridad: quien use este puerto en un flujo público (registro, recuperación de
	// contraseña) no debe revelar si la cédula existe. Un ErrNotFound se trata igual que
	// una fecha de expedición o un código dactilar incorrectos: mismo 401 y mismo mensaje
	// ("la información ingresada es incorrecta"), y cuenta como intento fallido.
	ConsultarCiudadano(ctx context.Context, cedula string) (Ciudadano, error)
}
