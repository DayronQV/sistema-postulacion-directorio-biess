package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/domain"
)

// Resultados de la validación del cargo con el Historial Laboral
// (docs/requerimiento-biess.md, ítems 1.4.1 a 1.4.3).
const (
	MsgCargoAfiliadosValido = "Postulante afiliado al IESS"
	MsgCargoJubiladosValido = "Postulante acredita jubilación"
	MsgCargoEmpleadorValido = "Postulante es empleador"

	MsgCargoAfiliadosNoElegible = "Su postulación para el cargo de Miembro Principal y Alterno en Representación del Sector de Afiliados ante el Directorio del BIESS, NO puede continuar, por cuanto no mantiene la calidad de afiliado al IESS"
	MsgCargoJubiladosNoElegible = "Su postulación para el cargo de Miembro Principal y Alterno en Representación del Sector de los Jubilados ante el Directorio del BIESS, NO puede continuar, por cuanto no acredita jubilación"
	MsgCargoEmpleadorNoElegible = "Su postulación para el cargo de Miembro Principal y Alterno en Representación del Sector Empleador ante el Directorio del BIESS, NO puede continuar, por cuanto no es empleador"
)

// maxNombreOrganizacion es la longitud máxima, en caracteres, del nombre de la
// organización de respaldo.
const maxNombreOrganizacion = 200

// ConHistorialLaboral activa la validación del cargo con el Historial Laboral dado. Sin
// él, guardar o validar un cargo falla con un error interno en lugar de omitir la
// regla, así que el cableado de producción siempre debe llamarlo. Devuelve el mismo
// servicio para encadenarlo al constructor.
func (s *PostulacionService) ConHistorialLaboral(h domain.HistorialLaboral) *PostulacionService {
	s.historial = h
	return s
}

// evaluarCargo consulta el Historial Laboral y decide si la cédula cumple la calidad que
// exige el cargo. No modifica la postulación.
func (s *PostulacionService) evaluarCargo(ctx context.Context, cedula, cargo string) (domain.ResultadoCargo, error) {
	if s.historial == nil {
		return domain.ResultadoCargo{}, errors.New("el Historial Laboral no está configurado")
	}
	calidad, err := s.historial.ConsultarCalidad(ctx, cedula)
	if err != nil {
		return domain.ResultadoCargo{}, fmt.Errorf("consultando el Historial Laboral: %w", err)
	}
	res := domain.ResultadoCargo{Cargo: cargo}
	switch cargo {
	case domain.CargoAfiliados:
		res.Elegible, res.Mensaje = calidad.EsAfiliado, MsgCargoAfiliadosNoElegible
		if res.Elegible {
			res.Mensaje = MsgCargoAfiliadosValido
		}
	case domain.CargoJubilados:
		res.Elegible, res.Mensaje = calidad.EsJubilado, MsgCargoJubiladosNoElegible
		if res.Elegible {
			res.Mensaje = MsgCargoJubiladosValido
		}
	case domain.CargoEmpleador:
		res.Elegible, res.Mensaje = calidad.EsEmpleador, MsgCargoEmpleadorNoElegible
		if res.Elegible {
			res.Mensaje = MsgCargoEmpleadorValido
		}
	default:
		return domain.ResultadoCargo{}, domain.Validation("Datos inválidos", map[string]string{"cargo": "Cargo no válido"})
	}
	return res, nil
}

// ValidarCargo comprueba con el Historial Laboral si el postulante cumple la calidad del
// cargo. Si la cumple, devuelve el resultado y no escribe nada (el cargo se guarda con
// PUT /postulacion). Si no, cierra la postulación y responde 422 con el mensaje del
// requerimiento; lo mismo ocurre si la nacionalidad no es ecuatoriana. 409 si la
// postulación ya no es un borrador.
func (s *PostulacionService) ValidarCargo(ctx context.Context, postulanteID, cargo string) (domain.ResultadoCargo, error) {
	if !domain.CargoValido(cargo) {
		return domain.ResultadoCargo{}, domain.Validation("Datos inválidos", map[string]string{"cargo": "Cargo no válido"})
	}
	if err := s.verificarPeriodo(ctx); err != nil {
		return domain.ResultadoCargo{}, err
	}
	actual, err := s.asegurar(ctx, postulanteID)
	if err != nil {
		return domain.ResultadoCargo{}, err
	}
	if !domain.EsNacionalidadEcuatoriana(actual.Personal.Nacionalidad) {
		return domain.ResultadoCargo{}, s.cerrar(ctx, postulanteID, MsgNacionalidadNoEcuatoriana)
	}
	res, err := s.evaluarCargo(ctx, actual.Personal.Cedula, cargo)
	if err != nil {
		return domain.ResultadoCargo{}, err
	}
	if !res.Elegible {
		return domain.ResultadoCargo{}, s.cerrar(ctx, postulanteID, res.Mensaje)
	}
	if actual.Estado != domain.EstadoBorrador {
		return domain.ResultadoCargo{}, domain.Conflict(MsgPostulacionNoEditable, nil)
	}
	return res, nil
}

// cerrar pasa el borrador a "cerrada" y devuelve siempre un error: 422 con el motivo, o
// 409 si la postulación ya no es un borrador.
func (s *PostulacionService) cerrar(ctx context.Context, postulanteID, motivo string) error {
	err := s.tx.Ejecutar(ctx, func(ctx context.Context, tx domain.Tx) error {
		p, err := s.postulaciones.BloquearPorPostulante(ctx, tx, postulanteID)
		if err != nil {
			return err
		}
		if p.Estado != domain.EstadoBorrador {
			return domain.Conflict(MsgPostulacionNoEditable, nil)
		}
		return s.postulaciones.Cerrar(ctx, tx, p.ID)
	})
	if err != nil {
		return err
	}
	return domain.Unprocessable(motivo, nil)
}

// limpiarRespaldo quita los espacios sobrantes del nombre de la organización y trata un
// documento_id vacío como ausente.
func limpiarRespaldo(r domain.RespaldoOrganizacion) domain.RespaldoOrganizacion {
	r.Nombre = strings.TrimSpace(r.Nombre)
	if r.DocumentoID != nil {
		id := strings.TrimSpace(*r.DocumentoID)
		r.DocumentoID = &id
		if id == "" {
			r.DocumentoID = nil
		}
	}
	return r
}

// validarRespaldo agrega a fields los errores del respaldo de la organización. Sin nombre
// ni documento significa que responde "No" y no se exige nada; con cualquiera de los dos
// (respuesta "Sí") se exigen ambos y un cargo elegido. Que el documento exista, sea de
// tipo organizacion y pertenezca a esta postulación lo comprueba el repositorio.
func validarRespaldo(f domain.Formulario, fields map[string]string) {
	r := f.RespaldoOrganizacion
	if r.Nombre == "" && r.DocumentoID == nil {
		return
	}
	if f.Cargo == "" {
		fields["cargo"] = "Seleccione un cargo para indicar la organización de respaldo"
	}
	switch {
	case r.Nombre == "":
		fields["respaldo_organizacion.nombre"] = msgObligatorio
	case utf8.RuneCountInString(r.Nombre) > maxNombreOrganizacion:
		fields["respaldo_organizacion.nombre"] = fmt.Sprintf("Máximo %d caracteres", maxNombreOrganizacion)
	}
	if r.DocumentoID == nil {
		fields["respaldo_organizacion.documento_id"] = msgObligatorio
	}
}
