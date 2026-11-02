package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/domain"
)

// Mensajes del envío de la postulación.
const (
	MsgPostulacionYaEnviada  = "La postulación ya fue enviada y no se puede enviar de nuevo"
	MsgPostulacionIncompleta = "La postulación está incompleta"
	msgSeccionIncompleta     = "Sección incompleta"
)

// ConMailer indica el correo con el que se envía el formulario al postulante. Sin él,
// Enviar cambia el estado igualmente pero no manda nada (por ejemplo, en pruebas).
func (s *PostulacionService) ConMailer(m domain.Mailer) *PostulacionService {
	s.mailer = m
	return s
}

// seccionesIncompletas devuelve, por cada sección obligatoria que falta, su clave en
// `fields` y el motivo. Sigue las mismas reglas con las que el formulario marca una sección
// como completa; Información personal viene del Registro Civil y no se evalúa.
func seccionesIncompletas(p domain.Postulacion) map[string]string {
	f := p.Formulario
	faltan := map[string]string{}
	falta := func(clave, motivo string) { faltan[clave] = motivo }

	d := f.Domicilio
	if d.PaisID == nil || d.TelefonoCelular == "" || (d.CantonID == nil && d.CiudadTexto == "") {
		falta("domicilio", "Indique país, ciudad y teléfono celular")
	}
	if f.Cargo == "" {
		falta("cargo", "Seleccione el cargo al que postula")
	}
	if len(f.Formacion) == 0 {
		falta("formacion", "Agregue al menos un título")
	}
	if len(f.Experiencia) == 0 {
		falta("experiencia", "Agregue al menos una experiencia profesional")
	}
	if len(f.Capacitaciones) == 0 {
		falta("capacitaciones", "Agregue al menos una capacitación")
	}
	if len(f.Referencias) == 0 {
		falta("referencias", "Agregue al menos una referencia personal")
	}
	if !f.Inhabilidades.Completa() {
		falta("inhabilidades", "Responda las 13 preguntas")
	}
	if !f.Consentimiento {
		falta("consentimiento", "Debe aceptar el consentimiento y la declaración")
	}
	for clave := range faltan {
		if faltan[clave] == "" {
			faltan[clave] = msgSeccionIncompleta
		}
	}
	return faltan
}

// Enviar cierra la edición de la postulación: comprueba que esté en borrador (409 si ya se
// envió o se cerró) y completa (422 con `fields` por sección), la marca "enviada" con su
// fecha y manda el formulario al correo del postulante, todo en una transacción: si el
// correo falla, la postulación sigue en borrador y se puede reintentar.
func (s *PostulacionService) Enviar(ctx context.Context, postulanteID string) (domain.Postulacion, error) {
	actual, err := s.Obtener(ctx, postulanteID)
	if err != nil {
		return domain.Postulacion{}, err
	}
	if actual.Estado != domain.EstadoBorrador {
		return domain.Postulacion{}, domain.Conflict(MsgPostulacionYaEnviada, nil)
	}
	if faltan := seccionesIncompletas(actual); len(faltan) > 0 {
		return domain.Postulacion{}, domain.Unprocessable(MsgPostulacionIncompleta, faltan)
	}

	err = s.tx.Ejecutar(ctx, func(ctx context.Context, tx domain.Tx) error {
		// La fila queda bloqueada: dos envíos simultáneos no pasan los dos.
		p, err := s.postulaciones.BloquearPorPostulante(ctx, tx, postulanteID)
		if err != nil {
			return err
		}
		if p.Estado != domain.EstadoBorrador {
			return domain.Conflict(MsgPostulacionYaEnviada, nil)
		}
		if err := s.postulaciones.MarcarEnviada(ctx, tx, p.ID); err != nil {
			return err
		}
		if s.mailer == nil {
			return nil
		}
		return s.mailer.Enviar(ctx, correoDeEnvio(actual))
	})
	if err != nil {
		return domain.Postulacion{}, err
	}
	return s.Obtener(ctx, postulanteID)
}

// correoDeEnvio arma el correo con el formulario completo de la postulación.
func correoDeEnvio(p domain.Postulacion) domain.Mensaje {
	f := p.Formulario
	var b strings.Builder
	linea := func(formato string, args ...any) { fmt.Fprintf(&b, formato+"\n", args...) }

	linea("Hola %s %s,", p.Personal.Nombres, p.Personal.Apellidos)
	linea("")
	linea("Recibimos su postulación al Concurso de Méritos y Oposición del Directorio del BIESS.")
	linea("Código: %s", p.Codigo)
	linea("Cargo: %s", f.Cargo)
	linea("")
	linea("Resumen del formulario")
	linea("- Cédula: %s", p.Personal.Cedula)
	linea("- Nacionalidad: %s", p.Personal.Nacionalidad)
	linea("- Teléfono celular: %s", f.Domicilio.TelefonoCelular)
	linea("- Formación académica: %d", len(f.Formacion))
	linea("- Experiencia profesional: %d (%d años, %d meses, %d días)", len(f.Experiencia),
		f.DuracionExperiencia.Anios, f.DuracionExperiencia.Meses, f.DuracionExperiencia.Dias)
	linea("- Publicaciones: %d", len(f.Publicaciones))
	linea("- Capacitaciones: %d", len(f.Capacitaciones))
	linea("- Referencias personales: %d", len(f.Referencias))
	linea("- Inhabilidades: 13 preguntas respondidas")
	linea("")
	linea("Imprima el formulario desde el sistema, fírmelo y entréguelo con los requisitos en el lugar, fecha y horario establecidos por el IESS.")

	return domain.Mensaje{
		Para:   p.Personal.Email,
		Asunto: "Postulación recibida " + p.Codigo,
		Texto:  b.String(),
	}
}
