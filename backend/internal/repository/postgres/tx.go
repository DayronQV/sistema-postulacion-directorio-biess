package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/domain"
)

// rollbackTimeout es el plazo del ROLLBACK, que no depende del contexto de la
// petición: si la petición se canceló o venció, igual hay que deshacer la transacción.
const rollbackTimeout = 5 * time.Second

// TxManager implementa domain.TxManager con transacciones de pgx.
type TxManager struct {
	db *pgxpool.Pool
}

var _ domain.TxManager = (*TxManager)(nil)

// NewTxManager crea el administrador de transacciones sobre el pool.
func NewTxManager(db *pgxpool.Pool) *TxManager {
	return &TxManager{db: db}
}

// Ejecutar abre una transacción (BEGIN), llama a fn con ella y hace COMMIT si fn
// devuelve nil. Ante cualquier error de fn, o un pánico, hace ROLLBACK: no queda ninguna
// escritura parcial. El error de fn se devuelve sin envolver.
func (m *TxManager) Ejecutar(ctx context.Context, fn func(ctx context.Context, tx domain.Tx) error) error {
	tx, err := m.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("abriendo la transacción: %w", err)
	}

	committed := false
	defer func() {
		if committed {
			return
		}
		rbCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), rollbackTimeout)
		defer cancel()
		// Si el COMMIT falló o la conexión se cayó, la transacción ya está cerrada y el
		// servidor la deshizo por su cuenta; el error que importa es el original.
		_ = tx.Rollback(rbCtx)
	}()

	if err := fn(ctx, tx); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("confirmando la transacción: %w", err)
	}
	committed = true
	return nil
}

// pgxTx recupera el pgx.Tx que TxManager entregó dentro de domain.Tx.
func pgxTx(tx domain.Tx) (pgx.Tx, error) {
	ptx, ok := tx.(pgx.Tx)
	if !ok {
		return nil, fmt.Errorf("la transacción no es de pgx (%T)", tx)
	}
	return ptx, nil
}
