package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/domain"
)

// EvaluacionRepo implementa domain.BancoEvaluacion sobre eval_preguntas, eval_config y
// eval_intentos.
type EvaluacionRepo struct {
	db *pgxpool.Pool
}

var _ domain.BancoEvaluacion = (*EvaluacionRepo)(nil)

// NewEvaluacionRepo crea el repositorio del banco de preguntas.
func NewEvaluacionRepo(db *pgxpool.Pool) *EvaluacionRepo {
	return &EvaluacionRepo{db: db}
}

// CrearPreguntas inserta todas las preguntas en una transacción: si una falla, no queda
// ninguna.
func (r *EvaluacionRepo) CrearPreguntas(ctx context.Context, preguntas []domain.Pregunta) ([]domain.Pregunta, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("iniciando la carga de preguntas: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	creadas := make([]domain.Pregunta, 0, len(preguntas))
	for _, p := range preguntas {
		opciones, err := json.Marshal(p.Opciones)
		if err != nil {
			return nil, fmt.Errorf("codificando las opciones: %w", err)
		}
		if err := tx.QueryRow(ctx, `
INSERT INTO eval_preguntas (enunciado, opciones, respuesta_correcta)
VALUES ($1, $2, $3)
RETURNING id`, p.Enunciado, opciones, p.RespuestaCorrecta).Scan(&p.ID); err != nil {
			return nil, fmt.Errorf("guardando la pregunta: %w", err)
		}
		creadas = append(creadas, p)
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("confirmando la carga de preguntas: %w", err)
	}
	return creadas, nil
}

// ListarPreguntas devuelve una página del banco. La búsqueda usa position() en vez de
// LIKE para que % y _ del texto buscado no se traten como comodines.
func (r *EvaluacionRepo) ListarPreguntas(ctx context.Context, f domain.FiltroPreguntas) ([]domain.Pregunta, int, error) {
	const donde = `WHERE $1 = '' OR position(lower($1) in lower(enunciado)) > 0`

	var total int
	if err := r.db.QueryRow(ctx, `SELECT count(*) FROM eval_preguntas `+donde, f.Buscar).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("contando las preguntas: %w", err)
	}
	rows, err := r.db.Query(ctx, `
SELECT id, enunciado, opciones, respuesta_correcta FROM eval_preguntas `+donde+`
ORDER BY id LIMIT $2 OFFSET $3`, f.Buscar, f.PorPagina, (f.Pagina-1)*f.PorPagina)
	if err != nil {
		return nil, 0, fmt.Errorf("listando las preguntas: %w", err)
	}
	defer rows.Close()

	preguntas := []domain.Pregunta{}
	for rows.Next() {
		var (
			p        domain.Pregunta
			opciones []byte
		)
		if err := rows.Scan(&p.ID, &p.Enunciado, &opciones, &p.RespuestaCorrecta); err != nil {
			return nil, 0, fmt.Errorf("leyendo una pregunta: %w", err)
		}
		if err := json.Unmarshal(opciones, &p.Opciones); err != nil {
			return nil, 0, fmt.Errorf("decodificando las opciones de la pregunta %d: %w", p.ID, err)
		}
		preguntas = append(preguntas, p)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("recorriendo las preguntas: %w", err)
	}
	return preguntas, total, nil
}

// ObtenerPregunta devuelve una pregunta; false si no existe.
func (r *EvaluacionRepo) ObtenerPregunta(ctx context.Context, id int) (domain.Pregunta, bool, error) {
	var (
		p        domain.Pregunta
		opciones []byte
	)
	err := r.db.QueryRow(ctx,
		`SELECT id, enunciado, opciones, respuesta_correcta FROM eval_preguntas WHERE id = $1`, id,
	).Scan(&p.ID, &p.Enunciado, &opciones, &p.RespuestaCorrecta)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Pregunta{}, false, nil
	}
	if err != nil {
		return domain.Pregunta{}, false, fmt.Errorf("leyendo la pregunta: %w", err)
	}
	if err := json.Unmarshal(opciones, &p.Opciones); err != nil {
		return domain.Pregunta{}, false, fmt.Errorf("decodificando las opciones de la pregunta %d: %w", id, err)
	}
	return p, true, nil
}

// ActualizarPregunta reemplaza enunciado, opciones y respuesta correcta.
func (r *EvaluacionRepo) ActualizarPregunta(ctx context.Context, p domain.Pregunta) (bool, error) {
	opciones, err := json.Marshal(p.Opciones)
	if err != nil {
		return false, fmt.Errorf("codificando las opciones: %w", err)
	}
	tag, err := r.db.Exec(ctx, `
UPDATE eval_preguntas SET enunciado = $2, opciones = $3, respuesta_correcta = $4
WHERE id = $1`, p.ID, p.Enunciado, opciones, p.RespuestaCorrecta)
	if err != nil {
		return false, fmt.Errorf("actualizando la pregunta: %w", err)
	}
	return tag.RowsAffected() == 1, nil
}

// EliminarPregunta borra una pregunta del banco.
func (r *EvaluacionRepo) EliminarPregunta(ctx context.Context, id int) (bool, error) {
	tag, err := r.db.Exec(ctx, `DELETE FROM eval_preguntas WHERE id = $1`, id)
	if err != nil {
		return false, fmt.Errorf("eliminando la pregunta: %w", err)
	}
	return tag.RowsAffected() == 1, nil
}

// ContarPreguntas devuelve cuántas preguntas tiene el banco.
func (r *EvaluacionRepo) ContarPreguntas(ctx context.Context) (int, error) {
	var n int
	if err := r.db.QueryRow(ctx, `SELECT count(*) FROM eval_preguntas`).Scan(&n); err != nil {
		return 0, fmt.Errorf("contando las preguntas: %w", err)
	}
	return n, nil
}

// HayIntentos indica si algún postulante ya inició la evaluación.
func (r *EvaluacionRepo) HayIntentos(ctx context.Context) (bool, error) {
	var hay bool
	if err := r.db.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM eval_intentos)`).Scan(&hay); err != nil {
		return false, fmt.Errorf("consultando los intentos de evaluación: %w", err)
	}
	return hay, nil
}

// ObtenerConfig lee la fila única de configuración; false si nunca se guardó.
func (r *EvaluacionRepo) ObtenerConfig(ctx context.Context) (domain.ConfigEvaluacion, bool, error) {
	var c domain.ConfigEvaluacion
	err := r.db.QueryRow(ctx,
		`SELECT duracion_minutos, num_preguntas FROM eval_config WHERE id = 1`,
	).Scan(&c.DuracionMinutos, &c.NumPreguntas)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ConfigEvaluacion{}, false, nil
	}
	if err != nil {
		return domain.ConfigEvaluacion{}, false, fmt.Errorf("leyendo la configuración de la evaluación: %w", err)
	}
	return c, true, nil
}

// GuardarConfig crea o reemplaza la configuración (la tabla solo admite la fila id = 1).
func (r *EvaluacionRepo) GuardarConfig(ctx context.Context, c domain.ConfigEvaluacion) error {
	_, err := r.db.Exec(ctx, `
INSERT INTO eval_config (id, duracion_minutos, num_preguntas)
VALUES (1, $1, $2)
ON CONFLICT (id) DO UPDATE
SET duracion_minutos = EXCLUDED.duracion_minutos,
    num_preguntas = EXCLUDED.num_preguntas`, c.DuracionMinutos, c.NumPreguntas)
	if err != nil {
		return fmt.Errorf("guardando la configuración de la evaluación: %w", err)
	}
	return nil
}

// --- Intentos de la evaluación (domain.IntentosEvaluacion) ---

var _ domain.IntentosEvaluacion = (*EvaluacionRepo)(nil)

// preguntaIntentoJSON es cómo se guarda una pregunta del intento en eval_intentos.preguntas.
type preguntaIntentoJSON struct {
	PreguntaID int   `json:"pregunta_id"`
	Orden      []int `json:"orden"`
}

func preguntasAJSON(ps []domain.PreguntaDelIntento) ([]byte, error) {
	items := make([]preguntaIntentoJSON, len(ps))
	for i, p := range ps {
		items[i] = preguntaIntentoJSON{PreguntaID: p.PreguntaID, Orden: p.Orden}
	}
	return json.Marshal(items)
}

func preguntasDesdeJSON(datos []byte) ([]domain.PreguntaDelIntento, error) {
	var items []preguntaIntentoJSON
	if err := json.Unmarshal(datos, &items); err != nil {
		return nil, err
	}
	ps := make([]domain.PreguntaDelIntento, len(items))
	for i, it := range items {
		ps[i] = domain.PreguntaDelIntento{PreguntaID: it.PreguntaID, Orden: it.Orden}
	}
	return ps, nil
}

// PreguntasParaSorteo devuelve el id y la cantidad de opciones de cada pregunta del banco,
// sin traer los textos ni la respuesta correcta.
func (r *EvaluacionRepo) PreguntasParaSorteo(ctx context.Context) ([]domain.PreguntaSorteable, error) {
	rows, err := r.db.Query(ctx, `SELECT id, jsonb_array_length(opciones) FROM eval_preguntas ORDER BY id`)
	if err != nil {
		return nil, fmt.Errorf("listando las preguntas para el sorteo: %w", err)
	}
	defer rows.Close()
	var ps []domain.PreguntaSorteable
	for rows.Next() {
		var p domain.PreguntaSorteable
		if err := rows.Scan(&p.ID, &p.NumOpciones); err != nil {
			return nil, fmt.Errorf("leyendo una pregunta del sorteo: %w", err)
		}
		ps = append(ps, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("recorriendo las preguntas del sorteo: %w", err)
	}
	return ps, nil
}

// PostulacionDelPostulante devuelve el estado de la postulación y si superó la Fase de Méritos.
func (r *EvaluacionRepo) PostulacionDelPostulante(ctx context.Context, postulanteID string) (domain.PostulacionEvaluacion, bool, error) {
	var p domain.PostulacionEvaluacion
	err := r.db.QueryRow(ctx,
		`SELECT id::text, estado, supero_merito FROM postulaciones WHERE postulante_id = $1`, postulanteID,
	).Scan(&p.ID, &p.Estado, &p.SuperoMerito)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.PostulacionEvaluacion{}, false, nil
	}
	if err != nil {
		return domain.PostulacionEvaluacion{}, false, fmt.Errorf("consultando la postulación para la evaluación: %w", err)
	}
	return p, true, nil
}

// CrearIntento inserta el intento si la postulación no tiene uno. La restricción UNIQUE de
// postulacion_id hace que dos llamadas simultáneas no creen dos: la que pierde recibe el
// intento ya guardado.
func (r *EvaluacionRepo) CrearIntento(ctx context.Context, i domain.IntentoEvaluacion) (domain.IntentoEvaluacion, bool, error) {
	preguntas, err := preguntasAJSON(i.Preguntas)
	if err != nil {
		return domain.IntentoEvaluacion{}, false, fmt.Errorf("codificando las preguntas del intento: %w", err)
	}
	err = r.db.QueryRow(ctx, `
INSERT INTO eval_intentos (postulacion_id, preguntas, inicio, cierre)
VALUES ($1, $2, $3, $4)
ON CONFLICT (postulacion_id) DO NOTHING
RETURNING id::text`, i.PostulacionID, preguntas, i.Inicio, i.Cierre).Scan(&i.ID)
	if err == nil {
		return i, true, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return domain.IntentoEvaluacion{}, false, fmt.Errorf("creando el intento de evaluación: %w", err)
	}
	existente, ok, err := r.ObtenerIntento(ctx, i.PostulacionID)
	if err != nil {
		return domain.IntentoEvaluacion{}, false, err
	}
	if !ok {
		return domain.IntentoEvaluacion{}, false, fmt.Errorf("el intento de la postulación %s desapareció mientras se creaba", i.PostulacionID)
	}
	return existente, false, nil
}

// ObtenerIntento devuelve el intento de la postulación; false si no tiene.
func (r *EvaluacionRepo) ObtenerIntento(ctx context.Context, postulacionID string) (domain.IntentoEvaluacion, bool, error) {
	var (
		i         domain.IntentoEvaluacion
		preguntas []byte
	)
	err := r.db.QueryRow(ctx, `
SELECT id::text, postulacion_id::text, preguntas, inicio, cierre, finalizado_at, nota::float8
FROM eval_intentos WHERE postulacion_id = $1`, postulacionID,
	).Scan(&i.ID, &i.PostulacionID, &preguntas, &i.Inicio, &i.Cierre, &i.FinalizadoEn, &i.Nota)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.IntentoEvaluacion{}, false, nil
	}
	if err != nil {
		return domain.IntentoEvaluacion{}, false, fmt.Errorf("leyendo el intento de evaluación: %w", err)
	}
	if i.Preguntas, err = preguntasDesdeJSON(preguntas); err != nil {
		return domain.IntentoEvaluacion{}, false, fmt.Errorf("decodificando las preguntas del intento %s: %w", i.ID, err)
	}
	return i, true, nil
}
