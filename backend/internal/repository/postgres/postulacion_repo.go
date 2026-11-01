package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/domain"
)

// Códigos de error de PostgreSQL que indican datos inválidos del cliente.
const (
	pgForeignKeyViolation = "23503"
	pgInvalidTextRepr     = "22P02"
)

// PostulacionRepo implementa domain.Postulaciones sobre la tabla postulaciones.
type PostulacionRepo struct {
	db *pgxpool.Pool
}

var _ domain.Postulaciones = (*PostulacionRepo)(nil)

// NewPostulacionRepo crea el repositorio de postulaciones.
func NewPostulacionRepo(db *pgxpool.Pool) *PostulacionRepo {
	return &PostulacionRepo{db: db}
}

// selectPostulacion lista las columnas de postulaciones en el orden de scanPostulacion.
// Los textos opcionales vuelven como cadena vacía para no manejar NULL en Go.
const selectPostulacion = `
SELECT id, codigo, postulante_id, estado, COALESCE(cargo, ''),
       nombres, apellidos, genero, fecha_nacimiento, nacionalidad, estado_civil,
       COALESCE(pareja_nombres, ''), COALESCE(pareja_cedula, ''),
       pais_id, provincia_id, canton_id,
       COALESCE(provincia_texto, ''), COALESCE(ciudad_texto, ''),
       COALESCE(telefono_celular, ''), COALESCE(email_secundario, ''),
       COALESCE(calle_principal, ''), COALESCE(numero, ''), COALESCE(calle_secundaria, ''),
       COALESCE(sector, ''), COALESCE(referencia, ''), COALESCE(telefono_domicilio, ''),
       tiene_familiares, tiene_publicaciones,
       COALESCE(respaldo_organizacion_nombre, ''), respaldo_organizacion_doc_id::text,
       consentimiento_at, created_at, updated_at, enviada_at,
       (SELECT email FROM postulantes WHERE postulantes.id = postulante_id),
       (SELECT cedula FROM postulantes WHERE postulantes.id = postulante_id)
FROM postulaciones`

func scanPostulacion(row pgx.Row) (domain.Postulacion, error) {
	var (
		p             domain.Postulacion
		estadoCivil   string
		parejaNombres string
		parejaCedula  string
		f             = &p.Formulario
		d             = &f.Domicilio
	)
	err := row.Scan(
		&p.ID, &p.Codigo, &p.PostulanteID, &p.Estado, &f.Cargo,
		&p.Personal.Nombres, &p.Personal.Apellidos, &p.Personal.Genero,
		&p.Personal.FechaNacimiento, &p.Personal.Nacionalidad, &estadoCivil,
		&parejaNombres, &parejaCedula,
		&d.PaisID, &d.ProvinciaID, &d.CantonID,
		&d.ProvinciaTexto, &d.CiudadTexto,
		&d.TelefonoCelular, &d.EmailSecundario,
		&d.CallePrincipal, &d.Numero, &d.CalleSecundaria,
		&d.Sector, &d.Referencia, &d.TelefonoDomicilio,
		&f.TieneFamiliares, &f.TienePublicaciones,
		&f.RespaldoOrganizacion.Nombre, &f.RespaldoOrganizacion.DocumentoID,
		&p.ConsentimientoAt, &p.CreadaEn, &p.ActualizadaEn, &p.EnviadaEn,
		&p.Personal.Email, &p.Personal.Cedula,
	)
	if err != nil {
		return domain.Postulacion{}, err
	}
	p.Personal.EstadoCivil = domain.EstadoCivil(estadoCivil)
	if parejaNombres != "" {
		p.Personal.Pareja = &domain.Pareja{NombresApellidos: parejaNombres, Cedula: parejaCedula}
	}
	f.Consentimiento = p.ConsentimientoAt != nil
	return p, nil
}

// CedulaDePostulante devuelve la cédula del postulante, o ErrNotFound.
func (r *PostulacionRepo) CedulaDePostulante(ctx context.Context, postulanteID string) (string, error) {
	var cedula string
	err := r.db.QueryRow(ctx, `SELECT cedula FROM postulantes WHERE id = $1`, postulanteID).Scan(&cedula)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", domain.NotFound("Postulante no encontrado")
	}
	if err != nil {
		return "", fmt.Errorf("consultando la cédula del postulante: %w", err)
	}
	return cedula, nil
}

// Crear inserta la postulación en borrador. El código sale de la secuencia
// postulacion_codigo_seq: único y creciente (BIESS-AAAA-NNNNNN). Si el postulante ya
// tiene una postulación no inserta nada y devuelve false.
func (r *PostulacionRepo) Crear(ctx context.Context, postulanteID string, p domain.InformacionPersonal, anio int) (bool, error) {
	var parejaNombres, parejaCedula *string
	if p.Pareja != nil {
		parejaNombres, parejaCedula = &p.Pareja.NombresApellidos, &p.Pareja.Cedula
	}
	tag, err := r.db.Exec(ctx, `
INSERT INTO postulaciones
    (codigo, postulante_id, nombres, apellidos, genero, fecha_nacimiento, nacionalidad,
     estado_civil, pareja_nombres, pareja_cedula)
VALUES
    (format('BIESS-%s-%s', $1::int, lpad(nextval('postulacion_codigo_seq')::text, 6, '0')),
     $2, $3, $4, $5, $6, $7, $8, $9, $10)
ON CONFLICT (postulante_id) DO NOTHING`,
		anio, postulanteID, p.Nombres, p.Apellidos, p.Genero, p.FechaNacimiento,
		p.Nacionalidad, string(p.EstadoCivil), parejaNombres, parejaCedula)
	if err != nil {
		return false, fmt.Errorf("creando la postulación: %w", err)
	}
	return tag.RowsAffected() == 1, nil
}

// ObtenerPorPostulante devuelve la postulación con su formulario base, o ErrNotFound.
func (r *PostulacionRepo) ObtenerPorPostulante(ctx context.Context, postulanteID string) (domain.Postulacion, error) {
	p, err := scanPostulacion(r.db.QueryRow(ctx, selectPostulacion+` WHERE postulante_id = $1`, postulanteID))
	return completar(p, err, "consultando la postulación")
}

// BloquearPorPostulante lee la postulación dentro de la transacción y bloquea su fila
// hasta el final de la transacción.
func (r *PostulacionRepo) BloquearPorPostulante(ctx context.Context, tx domain.Tx, postulanteID string) (domain.Postulacion, error) {
	ptx, err := pgxTx(tx)
	if err != nil {
		return domain.Postulacion{}, err
	}
	p, err := scanPostulacion(ptx.QueryRow(ctx, selectPostulacion+` WHERE postulante_id = $1 FOR UPDATE`, postulanteID))
	return completar(p, err, "bloqueando la postulación")
}

func completar(p domain.Postulacion, err error, contexto string) (domain.Postulacion, error) {
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Postulacion{}, domain.NotFound("Postulación no encontrada")
	}
	if err != nil {
		return domain.Postulacion{}, fmt.Errorf("%s: %w", contexto, err)
	}
	return p, nil
}

// msgReferenciaInvalida es el mensaje de un identificador que no existe o no se puede
// usar en esta postulación.
const msgReferenciaInvalida = "Referencia no válida"

// Actualizar escribe los campos de la postulación. Una referencia que no existe (país,
// provincia, cantón), un documento que no es de esta postulación o un identificador mal
// formado es un dato inválido del cliente, no un error del servidor.
func (r *PostulacionRepo) Actualizar(ctx context.Context, tx domain.Tx, postulacionID string, f domain.Formulario) error {
	ptx, err := pgxTx(tx)
	if err != nil {
		return err
	}
	if err := r.comprobarRespaldo(ctx, ptx, postulacionID, f.RespaldoOrganizacion.DocumentoID); err != nil {
		return err
	}
	d := f.Domicilio
	_, err = ptx.Exec(ctx, `
UPDATE postulaciones SET
    cargo = NULLIF($2, ''),
    pais_id = $3, provincia_id = $4, canton_id = $5,
    provincia_texto = NULLIF($6, ''), ciudad_texto = NULLIF($7, ''),
    telefono_celular = NULLIF($8, ''), email_secundario = NULLIF($9, ''),
    calle_principal = NULLIF($10, ''), numero = NULLIF($11, ''), calle_secundaria = NULLIF($12, ''),
    sector = NULLIF($13, ''), referencia = NULLIF($14, ''), telefono_domicilio = NULLIF($15, ''),
    tiene_familiares = $16, tiene_publicaciones = $17,
    respaldo_organizacion_nombre = NULLIF($18, ''), respaldo_organizacion_doc_id = $19::uuid,
    consentimiento_at = CASE WHEN $20 THEN COALESCE(consentimiento_at, now()) ELSE NULL END,
    updated_at = now()
WHERE id = $1`,
		postulacionID, f.Cargo,
		d.PaisID, d.ProvinciaID, d.CantonID,
		d.ProvinciaTexto, d.CiudadTexto,
		d.TelefonoCelular, d.EmailSecundario,
		d.CallePrincipal, d.Numero, d.CalleSecundaria,
		d.Sector, d.Referencia, d.TelefonoDomicilio,
		f.TieneFamiliares, f.TienePublicaciones,
		f.RespaldoOrganizacion.Nombre, f.RespaldoOrganizacion.DocumentoID,
		f.Consentimiento)
	if err != nil {
		if esReferenciaInvalida(err) {
			return domain.Validation("Datos inválidos", map[string]string{"domicilio": msgReferenciaInvalida})
		}
		return fmt.Errorf("actualizando la postulación: %w", err)
	}
	return nil
}

// comprobarRespaldo exige que el documento de respaldo de la organización exista, sea de
// tipo organizacion y pertenezca a esta misma postulación: la clave foránea sola
// aceptaría el documento de otro postulante.
func (r *PostulacionRepo) comprobarRespaldo(ctx context.Context, tx pgx.Tx, postulacionID string, documentoID *string) error {
	if documentoID == nil {
		return nil
	}
	var propio bool
	err := tx.QueryRow(ctx, `
SELECT EXISTS (
    SELECT 1 FROM documentos
    WHERE id = $1::uuid AND postulacion_id = $2 AND tipo = 'organizacion')`,
		*documentoID, postulacionID).Scan(&propio)
	if err != nil {
		if esReferenciaInvalida(err) {
			propio = false
		} else {
			return fmt.Errorf("comprobando el documento de respaldo: %w", err)
		}
	}
	if !propio {
		return domain.Validation("Datos inválidos", map[string]string{"respaldo_organizacion.documento_id": msgReferenciaInvalida})
	}
	return nil
}

// esReferenciaInvalida indica un error de PostgreSQL causado por un dato del cliente: una
// clave foránea que no existe o un identificador mal formado.
func esReferenciaInvalida(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && (pgErr.Code == pgForeignKeyViolation || pgErr.Code == pgInvalidTextRepr)
}

// Cerrar pasa la postulación a estado "cerrada".
func (r *PostulacionRepo) Cerrar(ctx context.Context, tx domain.Tx, postulacionID string) error {
	ptx, err := pgxTx(tx)
	if err != nil {
		return err
	}
	if _, err := ptx.Exec(ctx,
		`UPDATE postulaciones SET estado = 'cerrada', updated_at = now() WHERE id = $1`, postulacionID,
	); err != nil {
		return fmt.Errorf("cerrando la postulación: %w", err)
	}
	return nil
}
