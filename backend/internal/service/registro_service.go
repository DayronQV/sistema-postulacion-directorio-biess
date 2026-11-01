// Package service contiene los casos de uso. Solo depende de los puertos de
// internal/domain: no conoce HTTP, SQL ni Redis.
package service

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"time"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/domain"
)

// Reglas del primer paso del registro (docs/especificacion.md, EF-11).
const (
	MaxIntentosRegistro     = 3
	VentanaIntentosRegistro = 10 * time.Minute
	BloqueoRegistro         = 10 * time.Minute
	TTLTicketRegistro       = 10 * time.Minute
)

// Mensajes literales del requerimiento.
const (
	MsgIdentidadIncorrecta = "la información ingresada es incorrecta"
	MsgRegistroBloqueado   = "registro fallido, inténtelo nuevamente en 10 minutos"
	MsgCedulaRegistrada    = "La cédula ya se encuentra registrada"
)

// Formatos aceptados para la fecha de expedición: ISO (como el resto de fechas de la
// API) y DD/MM/AAAA (como la escribe el postulante).
var formatosFecha = []string{"2006-01-02", "02/01/2006"}

var codigoDactilarRe = regexp.MustCompile(`^[A-Z0-9]{10}$`)

// IdentidadInput son los datos del primer paso del registro. Se envía exactamente uno
// de FechaExpedicion y CodigoDactilar.
type IdentidadInput struct {
	Cedula          string
	FechaExpedicion string
	CodigoDactilar  string
}

// RegistroService implementa el registro de postulantes.
type RegistroService struct {
	registroCivil domain.RegistroCivil
	intentos      domain.Intentos
	tickets       domain.TicketsRegistro
	postulantes   domain.Postulantes
}

// NewRegistroService crea el servicio con sus dependencias inyectadas.
func NewRegistroService(
	registroCivil domain.RegistroCivil,
	intentos domain.Intentos,
	tickets domain.TicketsRegistro,
	postulantes domain.Postulantes,
) *RegistroService {
	return &RegistroService{
		registroCivil: registroCivil,
		intentos:      intentos,
		tickets:       tickets,
		postulantes:   postulantes,
	}
}

// ValidarIdentidad comprueba la cédula contra el Registro Civil y, si coincide,
// devuelve un ticket para completar el registro.
//
// Al tercer fallo la cédula queda bloqueada 10 minutos y, mientras dure, cualquier
// intento responde LOCKED aunque los datos sean correctos. Una cédula que no existe
// en el Registro Civil cuenta como un dato incorrecto, para no revelar qué cédulas
// existen. Que la cédula ya tenga cuenta solo se informa a quien acertó su identidad.
func (s *RegistroService) ValidarIdentidad(ctx context.Context, in IdentidadInput) (string, error) {
	cedula := strings.TrimSpace(in.Cedula)
	if err := domain.ValidarCedula(cedula); err != nil {
		return "", err
	}
	fecha, codigo, err := parseDatoIdentidad(in)
	if err != nil {
		return "", err
	}

	restante, err := s.intentos.Bloqueo(ctx, cedula)
	if err != nil {
		return "", err
	}
	if restante > 0 {
		return "", domain.Locked(MsgRegistroBloqueado, restante)
	}

	ciudadano, err := s.registroCivil.ConsultarCiudadano(ctx, cedula)
	switch {
	case errors.Is(err, domain.ErrNotFound):
		return "", s.fallo(ctx, cedula)
	case err != nil:
		return "", err
	}
	if !coincide(ciudadano, fecha, codigo) {
		return "", s.fallo(ctx, cedula)
	}

	if err := s.intentos.Reiniciar(ctx, cedula); err != nil {
		return "", err
	}
	existe, err := s.postulantes.ExisteCedula(ctx, cedula)
	if err != nil {
		return "", err
	}
	if existe {
		return "", domain.Conflict(MsgCedulaRegistrada, map[string]string{"cedula": MsgCedulaRegistrada})
	}
	return s.tickets.Emitir(ctx, cedula, TTLTicketRegistro)
}

// fallo registra un intento fallido y bloquea la cédula al llegar al máximo.
func (s *RegistroService) fallo(ctx context.Context, cedula string) error {
	n, err := s.intentos.RegistrarFallo(ctx, cedula, VentanaIntentosRegistro)
	if err != nil {
		return err
	}
	if n >= MaxIntentosRegistro {
		if err := s.intentos.Bloquear(ctx, cedula, BloqueoRegistro); err != nil {
			return err
		}
		return domain.Locked(MsgRegistroBloqueado, BloqueoRegistro)
	}
	return domain.Unauthorized(MsgIdentidadIncorrecta)
}

// parseDatoIdentidad valida que venga exactamente uno de los dos datos y su formato.
// Devuelve la fecha (o el valor cero) y el código dactilar en mayúscula (o "").
func parseDatoIdentidad(in IdentidadInput) (time.Time, string, error) {
	fechaTxt := strings.TrimSpace(in.FechaExpedicion)
	codigo := strings.ToUpper(strings.TrimSpace(in.CodigoDactilar))

	switch {
	case fechaTxt == "" && codigo == "":
		msg := "Ingrese la fecha de expedición o el código dactilar"
		return time.Time{}, "", domain.Validation("Datos inválidos", map[string]string{
			"fecha_expedicion": msg, "codigo_dactilar": msg,
		})
	case fechaTxt != "" && codigo != "":
		msg := "Ingrese solo uno: la fecha de expedición o el código dactilar"
		return time.Time{}, "", domain.Validation("Datos inválidos", map[string]string{
			"fecha_expedicion": msg, "codigo_dactilar": msg,
		})
	case codigo != "":
		if !codigoDactilarRe.MatchString(codigo) {
			return time.Time{}, "", domain.Validation("Datos inválidos", map[string]string{
				"codigo_dactilar": "El código dactilar debe tener 10 letras o números",
			})
		}
		return time.Time{}, codigo, nil
	}

	for _, layout := range formatosFecha {
		if f, err := time.Parse(layout, fechaTxt); err == nil {
			return f, "", nil
		}
	}
	return time.Time{}, "", domain.Validation("Datos inválidos", map[string]string{
		"fecha_expedicion": "Fecha no válida (DD/MM/AAAA)",
	})
}

// coincide compara el dato enviado con el del Registro Civil. Las fechas se comparan
// por día, sin importar la hora ni la zona.
func coincide(c domain.Ciudadano, fecha time.Time, codigo string) bool {
	if codigo != "" {
		return codigo == strings.ToUpper(c.CodigoDactilar)
	}
	y1, m1, d1 := fecha.Date()
	y2, m2, d2 := c.FechaExpedicion.Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}
