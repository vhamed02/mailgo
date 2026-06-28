package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// TxManager handles database transactions
type TxManager struct {
	db *pgxpool.Pool
}

// NewTxManager creates a new transaction manager
func NewTxManager(db *pgxpool.Pool) *TxManager {
	return &TxManager{db: db}
}

// WithTransaction executes a function within a database transaction
// If the function returns an error, the transaction is rolled back
// Otherwise, the transaction is committed
func (tm *TxManager) WithTransaction(ctx context.Context, fn func(tx pgx.Tx) error) error {
	tx, err := tm.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Ensure transaction is finalized
	defer func() {
		if p := recover(); p != nil {
			// Panic occurred, rollback
			_ = tx.Rollback(ctx)
			panic(p) // Re-throw panic after rollback
		}
	}()

	// Execute the function
	if err := fn(tx); err != nil {
		// Function returned error, rollback
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			return fmt.Errorf("failed to rollback transaction (original error: %v): %w", err, rbErr)
		}
		return err
	}

	// Success, commit
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetDB returns the underlying database pool
func (tm *TxManager) GetDB() *pgxpool.Pool {
	return tm.db
}
