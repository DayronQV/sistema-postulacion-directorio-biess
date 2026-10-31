// Package registrocivil contiene los adaptadores del puerto domain.RegistroCivil.
// Hoy solo existe el simulador (Mock): el Registro Civil real solo se consulta con un
// convenio del Estado y no hay acceso para este proyecto.
package registrocivil

import (
	"context"
	"time"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/domain"
)

// Cédulas de prueba. Los demás módulos pueden usarlas en sus pruebas. Todas son
// válidas por módulo 10 y no pertenecen a personas reales.
const (
	CedulaFlujoFeliz        = "1712345675" // soltera, ecuatoriana (afiliada)
	CedulaCasada            = "0912345675" // casada, con datos de la pareja (jubilada)
	CedulaUnionDeHecho      = "1312345679" // unión de hecho, con datos de la pareja (empleador)
	CedulaExtranjera        = "0103456786" // colombiana: bloqueo por nacionalidad
	CedulaSinCalidad        = "1709876542" // ni afiliado, ni jubilado, ni empleador
	CedulaBloqueoPorIntento = "0924681356" // pruebas del bloqueo por intentos fallidos
	CedulaRecuperacion      = "1802468130" // pruebas de recuperación de contraseña
)

// Mock es un Registro Civil simulado con ciudadanos de prueba fijos.
type Mock struct {
	ciudadanos map[string]domain.Ciudadano
}

var _ domain.RegistroCivil = (*Mock)(nil)

// NewMock crea el simulador con los 7 ciudadanos de prueba.
func NewMock() *Mock {
	lista := []domain.Ciudadano{
		{
			Cedula: CedulaFlujoFeliz, Nombres: "Carlos Andrés", Apellidos: "Mendoza Villacís",
			Genero: "Masculino", FechaNacimiento: fecha(1985, 3, 14),
			Nacionalidad: "Ecuatoriana", EstadoCivil: domain.EstadoSoltero,
			FechaExpedicion: fecha(2019, 5, 20), CodigoDactilar: "V1234V5678",
		},
		{
			Cedula: CedulaCasada, Nombres: "María Fernanda", Apellidos: "Zambrano Cedeño",
			Genero: "Femenino", FechaNacimiento: fecha(1978, 11, 2),
			Nacionalidad: "Ecuatoriana", EstadoCivil: domain.EstadoCasado,
			Pareja:          &domain.Pareja{NombresApellidos: "Luis Alberto Párraga Mera", Cedula: "0913579249"},
			FechaExpedicion: fecha(2017, 8, 11), CodigoDactilar: "E3344I5566",
		},
		{
			Cedula: CedulaUnionDeHecho, Nombres: "José Luis", Apellidos: "Cevallos Intriago",
			Genero: "Masculino", FechaNacimiento: fecha(1982, 7, 25),
			Nacionalidad: "Ecuatoriana", EstadoCivil: domain.EstadoUnionDeHecho,
			Pareja:          &domain.Pareja{NombresApellidos: "Ana Lucía Moreira Zambrano", Cedula: "1313579243"},
			FechaExpedicion: fecha(2021, 2, 3), CodigoDactilar: "V2468E1357",
		},
		{
			Cedula: CedulaExtranjera, Nombres: "Andrés Felipe", Apellidos: "Rojas Gómez",
			Genero: "Masculino", FechaNacimiento: fecha(1990, 1, 30),
			Nacionalidad: "Colombiana", EstadoCivil: domain.EstadoSoltero,
			FechaExpedicion: fecha(2022, 10, 17), CodigoDactilar: "V9876V5432",
		},
		{
			Cedula: CedulaSinCalidad, Nombres: "Sofía Elizabeth", Apellidos: "Paredes Naranjo",
			Genero: "Femenino", FechaNacimiento: fecha(1993, 9, 8),
			Nacionalidad: "Ecuatoriana", EstadoCivil: domain.EstadoSoltero,
			FechaExpedicion: fecha(2020, 6, 29), CodigoDactilar: "I1122V3344",
		},
		{
			Cedula: CedulaBloqueoPorIntento, Nombres: "Diego Armando", Apellidos: "Quiroz Macías",
			Genero: "Masculino", FechaNacimiento: fecha(1988, 12, 12),
			Nacionalidad: "Ecuatoriana", EstadoCivil: domain.EstadoSoltero,
			FechaExpedicion: fecha(2018, 12, 5), CodigoDactilar: "E5566V7788",
		},
		{
			Cedula: CedulaRecuperacion, Nombres: "Gabriela Alejandra", Apellidos: "Tello Freire",
			Genero: "Femenino", FechaNacimiento: fecha(1975, 4, 19),
			Nacionalidad: "Ecuatoriana", EstadoCivil: domain.EstadoDivorciado,
			FechaExpedicion: fecha(2016, 3, 22), CodigoDactilar: "V3141V5926",
		},
	}

	m := &Mock{ciudadanos: make(map[string]domain.Ciudadano, len(lista))}
	for _, c := range lista {
		m.ciudadanos[c.Cedula] = c
	}
	return m
}

// ConsultarCiudadano devuelve los datos de la cédula de prueba, o un error
// domain.ErrNotFound si no es una de las 7 del simulador.
func (m *Mock) ConsultarCiudadano(ctx context.Context, cedula string) (domain.Ciudadano, error) {
	if err := ctx.Err(); err != nil {
		return domain.Ciudadano{}, err
	}
	c, ok := m.ciudadanos[cedula]
	if !ok {
		return domain.Ciudadano{}, domain.NotFound("La cédula no existe en el Registro Civil")
	}
	// Copia de la pareja para que quien llama no pueda modificar los datos del simulador.
	if c.Pareja != nil {
		pareja := *c.Pareja
		c.Pareja = &pareja
	}
	return c, nil
}

func fecha(anio int, mes time.Month, dia int) time.Time {
	return time.Date(anio, mes, dia, 0, 0, 0, 0, time.UTC)
}
