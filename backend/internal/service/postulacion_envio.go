package service

import (
	"context"
	"fmt"
	"strings"
	"time"

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
	// El correo se arma antes de bloquear la fila: consulta los catálogos geográficos.
	correo := correoDeEnvio(actual, s.ubicacion(ctx, actual.Formulario.Domicilio))

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

// ubicacionCorreo son los nombres de país, provincia y ciudad del domicilio.
type ubicacionCorreo struct {
	pais, provincia, ciudad string
}

// ubicacion traduce los ids de catálogo del domicilio a nombres (fuera de Ecuador,
// provincia y ciudad son texto libre). Es de mejor esfuerzo: si el catálogo no responde,
// el correo sale igual con lo que haya, porque no debe impedir el envío.
func (s *PostulacionService) ubicacion(ctx context.Context, d domain.Domicilio) ubicacionCorreo {
	u := ubicacionCorreo{provincia: d.ProvinciaTexto, ciudad: d.CiudadTexto}
	if s.catalogos == nil || d.PaisID == nil {
		return u
	}
	nombre := func(items []domain.ItemCatalogo, err error, id *int) string {
		if err != nil || id == nil {
			return ""
		}
		for _, i := range items {
			if i.ID == *id {
				return i.Nombre
			}
		}
		return ""
	}
	paises, err := s.catalogos.Paises(ctx)
	u.pais = nombre(paises, err, d.PaisID)
	if d.ProvinciaID != nil {
		provincias, err := s.catalogos.Provincias(ctx, *d.PaisID)
		u.provincia = nombre(provincias, err, d.ProvinciaID)
		if d.CantonID != nil {
			cantones, err := s.catalogos.Cantones(ctx, *d.ProvinciaID)
			u.ciudad = nombre(cantones, err, d.CantonID)
		}
	}
	return u
}

// Nombres de los cargos y etiquetas de los valores guardados, como los muestra el formulario.
var (
	nombresCargo = map[string]string{
		domain.CargoAfiliados: "Miembro Principal y Alterno en Representación del Sector de Afiliados ante el Directorio del BIESS",
		domain.CargoJubilados: "Miembro Principal y Alterno en Representación del Sector de los Jubilados ante el Directorio del BIESS",
		domain.CargoEmpleador: "Miembro Principal y Alterno en Representación del Sector Empleador ante el Directorio del BIESS",
	}
	etiquetas = map[string]string{
		"tercer_nivel": "Tercer Nivel de Grado", "especialidad": "Cuarto Nivel - Especialidad Académica",
		"maestria": "Cuarto Nivel - Maestría Académica", "phd": "Cuarto Nivel PhD - Doctorado",
		"nacional": "Nacional", "extranjera": "Extranjera",
		"finanzas": "Finanzas", "mercado_capitales": "Mercado de Capitales",
		"administracion_empresas": "Administración de Empresas", "economia": "Economía",
		"derecho": "Derecho", "afines": "Materias afines",
		"general": "General", "direccion": "Dirección",
		"publicacion": "Publicación", "investigacion": "Trabajo de investigación",
		"curso": "Curso", "seminario": "Seminario", "taller": "Taller",
	}
)

func etiqueta(v string) string {
	if e, ok := etiquetas[v]; ok {
		return e
	}
	return valor(v)
}

func valor(v string) string {
	if strings.TrimSpace(v) == "" {
		return "—"
	}
	return v
}

func siNo(v bool) string {
	if v {
		return "Sí"
	}
	return "No"
}

// fecha convierte AAAA-MM-DD en DD/MM/AAAA; otro formato se deja tal cual.
func fecha(iso string) string {
	if t, err := time.Parse(time.DateOnly, iso); err == nil {
		return t.Format("02/01/2006")
	}
	return valor(iso)
}

func duracion(d domain.Duracion) string {
	return fmt.Sprintf("%d años, %d meses, %d días", d.Anios, d.Meses, d.Dias)
}

// correoDeEnvio arma el correo con el formulario completo de la postulación, sección por
// sección (requerimiento 1.10.3: se envía con todos los campos llenos por el postulante).
func correoDeEnvio(p domain.Postulacion, u ubicacionCorreo) domain.Mensaje {
	f := p.Formulario
	per := p.Personal
	var b strings.Builder
	linea := func(formato string, args ...any) { fmt.Fprintf(&b, formato+"\n", args...) }
	titulo := func(t string) { linea(""); linea("%s", t) }

	linea("Hola %s %s,", per.Nombres, per.Apellidos)
	linea("")
	linea("Recibimos su postulación al Concurso de Méritos y Oposición del Directorio del BIESS.")
	linea("Código único: %s", p.Codigo)

	titulo("1. INFORMACIÓN PERSONAL")
	linea("- Nombres y apellidos: %s %s", per.Nombres, per.Apellidos)
	linea("- Cédula: %s", per.Cedula)
	linea("- Género: %s", valor(per.Genero))
	linea("- Fecha de nacimiento: %s", per.FechaNacimiento.Format("02/01/2006"))
	linea("- Nacionalidad: %s", valor(per.Nacionalidad))
	linea("- Estado civil: %s", valor(string(per.EstadoCivil)))
	if per.Pareja != nil {
		linea("- Pareja: %s, cédula %s", per.Pareja.NombresApellidos, per.Pareja.Cedula)
	}

	d := f.Domicilio
	titulo("2. INFORMACIÓN DOMICILIARIA Y DE CONTACTO")
	linea("- País: %s", valor(u.pais))
	linea("- Provincia / Estado: %s", valor(u.provincia))
	linea("- Ciudad: %s", valor(u.ciudad))
	linea("- Teléfono celular: %s", valor(d.TelefonoCelular))
	linea("- Correo electrónico 1: %s", valor(per.Email))
	linea("- Correo electrónico 2: %s", valor(d.EmailSecundario))
	linea("- Dirección: %s %s y %s, sector %s", valor(d.CallePrincipal), valor(d.Numero),
		valor(d.CalleSecundaria), valor(d.Sector))
	linea("- Referencia: %s", valor(d.Referencia))
	linea("- Teléfono de domicilio: %s", valor(d.TelefonoDomicilio))

	titulo("3. INFORMACIÓN ADICIONAL")
	linea("- ¿Tiene familiares trabajando en el IESS o BIESS?: %s", siNo(f.TieneFamiliares))
	for i, fam := range f.Familiares {
		linea("  %d. %s (%s), %s, %s / %s", i+1, fam.NombresApellidos, fam.Parentesco,
			fam.Institucion, valor(fam.Area), valor(fam.Cargo))
	}

	titulo("4. SELECCIÓN DEL CARGO")
	linea("- %s", valor(nombresCargo[f.Cargo]))
	if f.RespaldoOrganizacion.Nombre != "" {
		linea("- Organización que respalda: %s", f.RespaldoOrganizacion.Nombre)
	}

	titulo("5. FORMACIÓN ACADÉMICA")
	for i, fo := range f.Formacion {
		linea("  %d. %s (%s), %s, %s. Área: %s. SENESCYT %s del %s", i+1, fo.Titulo,
			etiqueta(fo.Nivel), fo.Institucion, etiqueta(fo.Tipo), etiqueta(fo.Area),
			valor(fo.SenescytNumero), fecha(fo.SenescytFecha))
	}

	titulo("6. EXPERIENCIA PROFESIONAL")
	for i, e := range f.Experiencia {
		linea("  %d. %s en %s (%s), del %s al %s: %s. Actividades: %s. %s", i+1, e.Cargo,
			e.Institucion, etiqueta(e.Tipo), fecha(e.FechaInicio), fecha(e.FechaFin),
			duracion(e.Duracion), etiqueta(e.Actividades), e.Descripcion)
	}
	linea("- Tiempo total de experiencia: %s", duracion(f.DuracionExperiencia))

	titulo("7. PUBLICACIONES")
	linea("- ¿Cuenta con publicaciones e investigaciones académicas?: %s", siNo(f.TienePublicaciones))
	for i, pub := range f.Publicaciones {
		linea("  %d. %s (%s), %s, %s. Relacionada con su ejercicio o formación: %s", i+1,
			pub.Titulo, etiqueta(pub.Tipo), valor(pub.Medio), fecha(pub.Fecha), siNo(pub.Relacionada))
	}

	titulo("8. CAPACITACIÓN")
	for i, c := range f.Capacitaciones {
		linea("  %d. %s (%s), %s, del %s al %s, %d horas", i+1, c.Nombre, etiqueta(c.Tipo),
			c.Institucion, fecha(c.FechaInicio), fecha(c.FechaFin), c.Horas)
	}

	titulo("9. REFERENCIAS PERSONALES")
	for i, r := range f.Referencias {
		linea("  %d. %s (%s), teléfono %s, correo %s", i+1, r.NombresApellidos, valor(r.Relacion),
			valor(r.Telefono), valor(r.Correo))
	}

	titulo("10. INHABILIDADES")
	for n := 1; n <= domain.TotalPreguntasInhabilidades; n++ {
		linea("  Pregunta %d: %s", n, siNo(f.Inhabilidades.Respuestas[n]))
	}
	if f.Inhabilidades.Respuestas[domain.PreguntaRemovido] {
		r := f.Inhabilidades.Removido
		linea("  Detalle de la pregunta %d: %s, %s, %s. Motivo: %s", domain.PreguntaRemovido,
			r.Institucion, valor(r.Cargo), fecha(r.FechaFin), r.Motivo)
	}

	titulo("CONSENTIMIENTO")
	linea("- Aceptó el consentimiento y la declaración de los 5 numerales.")
	linea("")
	linea("Imprima el formulario desde el sistema, fírmelo y entréguelo con los requisitos en el lugar, fecha y horario establecidos por el IESS.")

	return domain.Mensaje{
		Para:   per.Email,
		Asunto: "Postulación recibida " + p.Codigo,
		Texto:  b.String(),
	}
}
