package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/domain"
)

// Mensajes de la postulación.
const (
	MsgNacionalidadNoEcuatoriana = "Solo pueden postularse personas de nacionalidad ecuatoriana"
	MsgPostulacionNoEditable     = "La postulación ya no se puede modificar"
)

// PostulacionService implementa los casos de uso de la postulación: obtenerla (y
// crearla la primera vez) y guardarla completa en una sola transacción.
type PostulacionService struct {
	postulaciones domain.Postulaciones
	registroCivil domain.RegistroCivil
	tx            domain.TxManager
	secciones     []domain.SeccionPostulacion
	catalogos     domain.Catalogos // opcional: ver ConCatalogos
	ahora         func() time.Time
}

// NewPostulacionService crea el servicio. secciones son los repositorios de las
// secciones con filas: Guardar los recorre en orden dentro de la misma transacción, así
// que agregar una sección nueva no requiere cambiar este servicio.
func NewPostulacionService(
	postulaciones domain.Postulaciones,
	registroCivil domain.RegistroCivil,
	tx domain.TxManager,
	secciones ...domain.SeccionPostulacion,
) *PostulacionService {
	return &PostulacionService{
		postulaciones: postulaciones,
		registroCivil: registroCivil,
		tx:            tx,
		secciones:     secciones,
		ahora:         time.Now,
	}
}

// Obtener devuelve la postulación del postulante con todas sus secciones. Si todavía
// no tiene una, la crea en estado borrador con el código siguiente y la información
// personal del Registro Civil.
func (s *PostulacionService) Obtener(ctx context.Context, postulanteID string) (domain.Postulacion, error) {
	p, err := s.asegurar(ctx, postulanteID)
	if err != nil {
		return domain.Postulacion{}, err
	}
	for _, sec := range s.secciones {
		if err := sec.Cargar(ctx, p.ID, &p.Formulario); err != nil {
			return domain.Postulacion{}, err
		}
	}
	return p, nil
}

// asegurar devuelve la postulación del postulante y la crea si no existe. Si dos
// peticiones la crean a la vez, solo una inserta (postulante_id es único) y la otra
// lee la que ganó.
func (s *PostulacionService) asegurar(ctx context.Context, postulanteID string) (domain.Postulacion, error) {
	p, err := s.postulaciones.ObtenerPorPostulante(ctx, postulanteID)
	if err == nil {
		return p, nil
	}
	if !errors.Is(err, domain.ErrNotFound) {
		return domain.Postulacion{}, err
	}

	cedula, err := s.postulaciones.CedulaDePostulante(ctx, postulanteID)
	if err != nil {
		return domain.Postulacion{}, err
	}
	ciudadano, err := s.registroCivil.ConsultarCiudadano(ctx, cedula)
	if err != nil {
		// Un postulante registrado debería existir en el Registro Civil: si no, es una
		// inconsistencia del servidor y no un "no encontrado" para el cliente.
		return domain.Postulacion{}, fmt.Errorf("consultando el Registro Civil: %v", err)
	}
	if _, err := s.postulaciones.Crear(ctx, postulanteID, personalDe(ciudadano), s.ahora().Year()); err != nil {
		return domain.Postulacion{}, err
	}
	return s.postulaciones.ObtenerPorPostulante(ctx, postulanteID)
}

func personalDe(c domain.Ciudadano) domain.InformacionPersonal {
	p := domain.InformacionPersonal{
		Cedula:          c.Cedula,
		Nombres:         c.Nombres,
		Apellidos:       c.Apellidos,
		Genero:          c.Genero,
		FechaNacimiento: c.FechaNacimiento,
		Nacionalidad:    c.Nacionalidad,
		EstadoCivil:     c.EstadoCivil,
	}
	if c.EstadoCivil.ConPareja() {
		p.Pareja = c.Pareja
	}
	return p
}

// Guardar reemplaza el estado completo del formulario en una sola transacción:
// actualiza la postulación y cada sección con filas, o no escribe nada. Solo se edita en
// borrador (409 si está enviada o cerrada). Si la nacionalidad no es ecuatoriana, la
// postulación queda cerrada y se responde 422.
func (s *PostulacionService) Guardar(ctx context.Context, postulanteID string, f domain.Formulario) (domain.Postulacion, error) {
	f.Familiares = limpiarFamiliares(f.Familiares)
	f.Referencias = limpiarReferencias(f.Referencias)
	f.Formacion = limpiarFormacion(f.Formacion)
	actual, err := s.asegurar(ctx, postulanteID)
	if err != nil {
		return domain.Postulacion{}, err
	}
	// Quien no es ecuatoriano no puede continuar: se omite la validación del formulario
	// para que un cuerpo inválido no oculte el cierre (422 en lugar de 400). La transacción
	// de abajo cierra la postulación o responde 409 si ya no es un borrador.
	if domain.EsNacionalidadEcuatoriana(actual.Personal.Nacionalidad) {
		if err := validarFormularioEn(f, s.ahora()); err != nil {
			return domain.Postulacion{}, err
		}
		if err := s.validarGeografia(ctx, f.Domicilio); err != nil {
			return domain.Postulacion{}, err
		}
	}

	var cerrada bool
	err = s.tx.Ejecutar(ctx, func(ctx context.Context, tx domain.Tx) error {
		// La fila queda bloqueada hasta el final: una edición concurrente o un envío
		// simultáneo esperan en lugar de pisarse.
		p, err := s.postulaciones.BloquearPorPostulante(ctx, tx, postulanteID)
		if err != nil {
			return err
		}
		if p.Estado != domain.EstadoBorrador {
			return domain.Conflict(MsgPostulacionNoEditable, nil)
		}
		if !domain.EsNacionalidadEcuatoriana(p.Personal.Nacionalidad) {
			cerrada = true
			return s.postulaciones.Cerrar(ctx, tx, p.ID)
		}
		if err := s.postulaciones.Actualizar(ctx, tx, p.ID, f); err != nil {
			return err
		}
		for _, sec := range s.secciones {
			if err := sec.Reemplazar(ctx, tx, p.ID, f); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return domain.Postulacion{}, err
	}
	if cerrada {
		return domain.Postulacion{}, domain.Unprocessable(MsgNacionalidadNoEcuatoriana, nil)
	}
	return s.Obtener(ctx, postulanteID)
}

// validarFormulario aplica las reglas del servidor sobre los campos de la propia
// postulación. La validación de cada sección con filas la hace su propio código.
func validarFormulario(f domain.Formulario) error {
	return validarFormularioEn(f, time.Now())
}

func validarFormularioEn(f domain.Formulario, ahora time.Time) error {
	fields := map[string]string{}
	validarDomicilioFormato(f.Domicilio, fields)
	if f.Cargo != "" && !domain.CargoValido(f.Cargo) {
		fields["cargo"] = "Cargo no válido"
	}
	validarFamiliares(f, fields)
	validarReferencias(f, fields)
	validarFormacion(f.Formacion, ahora, fields)
	if len(fields) > 0 {
		return domain.Validation("Datos inválidos", fields)
	}
	return nil
}
