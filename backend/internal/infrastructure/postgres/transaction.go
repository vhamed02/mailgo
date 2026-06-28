package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// TxKey is the context key used to carry an active pgx.Tx.
type TxKey struct{}

// DBTX is satisfied by both *pgxpool.Pool and pgx.Tx, allowing repositories
// to work transparently inside or outside a transaction.
type DBTX interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
}

// TxFromContext returns the active pgx.Tx when the caller is running inside
// WithTransaction, or nil otherwise.
func TxFromContext(ctx context.Context) pgx.Tx {
	tx, _ := ctx.Value(TxKey{}).(pgx.Tx)
	return tx
}

// GetDB returns the active pgx.Tx from ctx when available,
// otherwise returns the provided pool. Use this in every repository method.
func GetDB(ctx context.Context, pool *pgxpool.Pool) DBTX {
	if tx := TxFromContext(ctx); tx != nil {
		return tx
	}
	return pool
}

// TxManager handles database transactions.
type TxManager struct {
	db *pgxpool.Pool
}

// NewTxManager creates a new transaction manager.
func NewTxManager(db *pgxpool.Pool) *TxManager {
	return &TxManager{db: db}
}

// WithTransaction executes fn inside a database transaction.
// It injects the tx into the context so repositories pick it up automatically
// via GetDB. If fn returns an error the transaction is rolled back, otherwise
// committed.
func (tm *TxManager) WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	tx, err := tm.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback(ctx)
			panic(p)
		}
	}()

	txCtx := context.WithValue(ctx, TxKey{}, tx)

	if err := fn(txCtx); err != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			return fmt.Errorf("rollback failed (original: %v): %w", err, rbErr)
		}
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetPool returns the underlying database pool.
func (tm *TxManager) GetPool() *pgxpool.Pool {
	return tm.db
}
