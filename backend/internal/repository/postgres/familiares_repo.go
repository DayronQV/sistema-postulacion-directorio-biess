package postgres

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/domain"
)

// FamiliaresRepo implementa domain.SeccionPostulacion para postulacion_familiares. Es
// la sección de referencia: las demás (formación, experiencia, ...) siguen el mismo
// molde en su propio archivo y se agregan al cableado sin tocar la transacción.
type FamiliaresRepo struct {
	db *pgxpool.Pool
}

var _ domain.SeccionPostulacion = (*FamiliaresRepo)(nil)

// NewFamiliaresRepo crea el repositorio de familiares.
func NewFamiliaresRepo(db *pgxpool.Pool) *FamiliaresRepo {
	return &FamiliaresRepo{db: db}
}

// Cargar completa f.Familiares con las filas de la postulación en el orden de captura.
// La tabla no tiene una columna de orden, así que ese orden vive en el id: Reemplazar lo
// genera creciente (ver idOrdenado) y aquí se ordena por él.
func (r *FamiliaresRepo) Cargar(ctx context.Context, postulacionID string, f *domain.Formulario) error {
	rows, err := r.db.Query(ctx, `
SELECT nombres_apellidos, parentesco, institucion, area, cargo
FROM postulacion_familiares
WHERE postulacion_id = $1
ORDER BY id`, postulacionID)
	if err != nil {
		return fmt.Errorf("consultando los familiares: %w", err)
	}
	defer rows.Close()

	f.Familiares = nil
	for rows.Next() {
		var fam domain.Familiar
		if err := rows.Scan(&fam.NombresApellidos, &fam.Parentesco, &fam.Institucion, &fam.Area, &fam.Cargo); err != nil {
			return fmt.Errorf("leyendo un familiar: %w", err)
		}
		f.Familiares = append(f.Familiares, fam)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("leyendo los familiares: %w", err)
	}
	return nil
}

// Reemplazar borra los familiares de la postulación y escribe los de f, dentro de tx.
func (r *FamiliaresRepo) Reemplazar(ctx context.Context, tx domain.Tx, postulacionID string, f domain.Formulario) error {
	ptx, err := pgxTx(tx)
	if err != nil {
		return err
	}
	if _, err := ptx.Exec(ctx, `DELETE FROM postulacion_familiares WHERE postulacion_id = $1`, postulacionID); err != nil {
		return fmt.Errorf("borrando los familiares: %w", err)
	}
	if !f.TieneFamiliares {
		return nil
	}
	ahora := time.Now()
	for i, fam := range f.Familiares {
		id, err := idOrdenado(ahora, i)
		if err != nil {
			return err
		}
		if _, err := ptx.Exec(ctx, `
INSERT INTO postulacion_familiares (id, postulacion_id, nombres_apellidos, parentesco, institucion, area, cargo)
VALUES ($1, $2, $3, $4, $5, $6, $7)`,
			id, postulacionID, fam.NombresApellidos, fam.Parentesco, fam.Institucion, fam.Area, fam.Cargo,
		); err != nil {
			return fmt.Errorf("guardando un familiar: %w", err)
		}
	}
	return nil
}

// idOrdenado genera un UUID (versión 7) cuyo orden de bytes sigue el orden de captura:
// los 48 bits altos son los milisegundos de ahora, que comparten todas las filas de un
// mismo guardado, y los 12 bits siguientes son la posición de la fila (0 a 4095). Postgres
// compara los uuid byte a byte, así que ORDER BY id devuelve las filas como se capturaron
// sin agregar una columna al esquema. El resto de los bits son aleatorios.
func idOrdenado(ahora time.Time, posicion int) (string, error) {
	if posicion < 0 || posicion > 0xfff {
		return "", fmt.Errorf("posición de familiar fuera de rango: %d", posicion)
	}
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", fmt.Errorf("generando el id de un familiar: %w", err)
	}
	var ms [8]byte
	binary.BigEndian.PutUint64(ms[:], uint64(ahora.UnixMilli()))
	copy(b[:6], ms[2:])
	b[6] = 0x70 | byte(posicion>>8)
	b[7] = byte(posicion)
	b[8] = 0x80 | b[8]&0x3f
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16]), nil
}
