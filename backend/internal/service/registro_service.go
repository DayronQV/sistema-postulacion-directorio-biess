// Package service contiene los casos de uso. Solo depende de los puertos de
// internal/domain: no conoce HTTP, SQL ni Redis.
package service

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"regexp"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"golang.org/x/crypto/bcrypt"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/domain"
)

// Reglas del primer paso del registro (docs/especificacion.md, EF-11).
const (
	MaxIntentosRegistro     = 3
	VentanaIntentosRegistro = 10 * time.Minute
	BloqueoRegistro         = 10 * time.Minute
	TTLTicketRegistro       = 10 * time.Minute
	MinPassword             = 8
	MaxPasswordBytes        = 72 // límite de bcrypt
)

// Mensajes literales del requerimiento.
const (
	MsgIdentidadIncorrecta = "la información ingresada es incorrecta"
	MsgRegistroBloqueado   = "registro fallido, inténtelo nuevamente en 10 minutos"
	MsgTicketInvalido      = "La validación de identidad expiró. Vuelva a validar su identidad"
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

	ok, err := comprobarIdentidad(ctx, s.registroCivil, cedula, fecha, codigo)
	if err != nil {
		return "", err
	}
	if !ok {
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
		return "", domain.ConflictoCedula()
	}
	return s.tickets.Emitir(ctx, cedula, TTLTicketRegistro)
}

// RegistroInput son los datos del segundo paso del registro.
type RegistroInput struct {
	Ticket   string
	Email    string
	Password string
}

// Registrar crea la cuenta del postulante cuya identidad ya validó el ticket.
// Devuelve el id de la cuenta. La contraseña se guarda solo como hash bcrypt y el
// ticket se borra al terminar, para que no se pueda reutilizar.
//
// Los datos se validan antes de leer el ticket: si el correo o la contraseña no
// sirven, el postulante los corrige sin repetir la validación de identidad.
func (s *RegistroService) Registrar(ctx context.Context, in RegistroInput) (string, error) {
	email := strings.ToLower(strings.TrimSpace(in.Email))
	fields := map[string]string{}
	if msg := validarEmail(email); msg != "" {
		fields["email"] = msg
	}
	if msg := validarPassword(in.Password); msg != "" {
		fields["password"] = msg
	}
	if len(fields) > 0 {
		return "", domain.Validation("Datos inválidos", fields)
	}

	ticket := strings.TrimSpace(in.Ticket)
	if ticket == "" {
		return "", domain.Unauthorized(MsgTicketInvalido)
	}
	cedula, err := s.tickets.Cedula(ctx, ticket)
	switch {
	case errors.Is(err, domain.ErrNotFound):
		return "", domain.Unauthorized(MsgTicketInvalido)
	case err != nil:
		return "", err
	}

	if existe, err := s.postulantes.ExisteCedula(ctx, cedula); err != nil {
		return "", err
	} else if existe {
		return "", domain.ConflictoCedula()
	}
	if existe, err := s.postulantes.ExisteEmail(ctx, email); err != nil {
		return "", err
	} else if existe {
		return "", domain.ConflictoEmail()
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("hasheando la contraseña: %w", err)
	}
	id, err := s.postulantes.Crear(ctx, domain.NuevoPostulante{
		Cedula: cedula, Email: email, PasswordHash: string(hash),
	})
	if err != nil {
		return "", err
	}

	// La cuenta ya existe: si el borrado falla no se informa como error, porque un
	// segundo uso del ticket choca igual con la cédula registrada (409) y el ticket
	// expira solo en 10 minutos.
	_ = s.tickets.Eliminar(ctx, ticket)
	return id, nil
}

// validarEmail devuelve el mensaje de error del correo, o "" si es válido.
func validarEmail(email string) string {
	if email == "" {
		return "Campo obligatorio"
	}
	addr, err := mail.ParseAddress(email)
	// ParseAddress acepta "Nombre <correo>"; aquí solo vale la dirección sola.
	if err != nil || addr.Address != email || !strings.Contains(email[strings.LastIndex(email, "@")+1:], ".") {
		return "Correo electrónico no válido"
	}
	return ""
}

// validarPassword exige al menos 8 caracteres con letras y números. El mínimo cuenta
// caracteres Unicode, no bytes ("á" es un carácter y dos bytes). El máximo sí cuenta
// bytes: bcrypt solo usa los primeros 72, así que una contraseña más larga se
// rechaza en lugar de truncarse en silencio.
func validarPassword(p string) string {
	if utf8.RuneCountInString(p) < MinPassword {
		return "La contraseña debe tener al menos 8 caracteres"
	}
	if len(p) > MaxPasswordBytes {
		return "La contraseña es demasiado larga (máximo 72 bytes)"
	}
	var letra, numero bool
	for _, r := range p {
		switch {
		case unicode.IsLetter(r):
			letra = true
		case unicode.IsDigit(r):
			numero = true
		}
	}
	if !letra || !numero {
		return "La contraseña debe combinar letras y números"
	}
	return ""
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

// comprobarIdentidad consulta el Registro Civil y compara el dato enviado. Una cédula
// que no existe devuelve false, igual que un dato incorrecto, para no revelar qué
// cédulas existen. Solo devuelve error si falla la consulta.
func comprobarIdentidad(ctx context.Context, rc domain.RegistroCivil, cedula string, fecha time.Time, codigo string) (bool, error) {
	ciudadano, err := rc.ConsultarCiudadano(ctx, cedula)
	switch {
	case errors.Is(err, domain.ErrNotFound):
		return false, nil
	case err != nil:
		return false, err
	}
	return coincide(ciudadano, fecha, codigo), nil
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
