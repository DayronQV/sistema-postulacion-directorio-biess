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
	if f.TieneFamiliares && len(f.Familiares) == 0 {
		falta("familiares", "Agregue al menos un familiar o responda que no tiene")
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
	if f.TienePublicaciones && len(f.Publicaciones) == 0 {
		falta("publicaciones", "Agregue al menos una publicación o responda que no tiene")
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
	if err := s.verificarPeriodo(ctx); err != nil {
		return domain.Postulacion{}, err
	}
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
	correo := correoDeEnvio(actual)

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
		return s.mailer.Enviar(ctx, correo)
	})
	if err != nil {
		return domain.Postulacion{}, err
	}
	return s.Obtener(ctx, postulanteID)
}

// correoDeEnvio arma el correo de confirmación: agradece la postulación, recuerda el código
// único y explica el próximo paso (imprimir el formulario que genera el sistema, firmarlo y
// entregarlo).
func correoDeEnvio(p domain.Postulacion) domain.Mensaje {
	var b strings.Builder
	linea := func(formato string, args ...any) { fmt.Fprintf(&b, formato+"\n", args...) }

	linea("Hola %s,", p.Personal.Nombres)
	linea("")
	linea("¡Muchas gracias por postular al Concurso de Méritos y Oposición para la selección de los Miembros del Directorio del BIESS! Nos alegra contar con su participación.")
	linea("")
	linea("Recibimos su postulación correctamente. Su código único es: %s", p.Codigo)
	linea("")
	linea("Próximo paso: ingrese al sistema, imprima el formulario de postulación que se generó con sus datos, fírmelo y entréguelo junto con los requisitos en el lugar, fecha y horario que establezca el BIESS.")
	linea("")
	linea("Le deseamos muchos éxitos en el proceso.")
	linea("")
	linea("Atentamente,")
	linea("Concurso Directorio BIESS")

	return domain.Mensaje{
		Para:   p.Personal.Email,
		Asunto: "¡Gracias por postular! Postulación " + p.Codigo + " recibida",
		Texto:  b.String(),
	}
}
