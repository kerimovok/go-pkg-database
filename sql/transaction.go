package sql

import (
	"context"
	"fmt"

	"gorm.io/gorm"
)

// TransactionFunc defines the function signature for transaction operations
type TransactionFunc func(*gorm.DB) error

// WithTransaction executes a function within a database transaction
func WithTransaction(db *gorm.DB, fn TransactionFunc) error {
	return WithTransactionContext(context.Background(), db, fn)
}

// WithTransactionContext executes a function within a database transaction with context
func WithTransactionContext(ctx context.Context, db *gorm.DB, fn TransactionFunc) error {
	if db == nil {
		return fmt.Errorf("database connection is nil")
	}

	tx := db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r) // Re-panic after cleanup
		}
	}()

	if err := fn(tx); err != nil {
		if rollbackErr := tx.Rollback().Error; rollbackErr != nil {
			return fmt.Errorf("transaction failed: %w, rollback failed: %v", err, rollbackErr)
		}
		return fmt.Errorf("transaction failed: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// WithNestedTransaction handles nested transactions using savepoints
func WithNestedTransaction(db *gorm.DB, fn TransactionFunc) error {
	return WithNestedTransactionContext(context.Background(), db, fn)
}

// WithNestedTransactionContext handles nested transactions using savepoints with context
func WithNestedTransactionContext(ctx context.Context, db *gorm.DB, fn TransactionFunc) error {
	if db == nil {
		return fmt.Errorf("database connection is nil")
	}

	// Check if we're already in a transaction
	if db.Statement != nil && db.Statement.DB != nil {
		// We're in a transaction, use savepoint
		savepoint := fmt.Sprintf("sp_%d", ctx.Value("savepoint_id"))
		if err := db.SavePoint(savepoint).Error; err != nil {
			return fmt.Errorf("failed to create savepoint: %w", err)
		}

		defer func() {
			if r := recover(); r != nil {
				db.RollbackTo(savepoint)
				panic(r)
			}
		}()

		if err := fn(db); err != nil {
			if rollbackErr := db.RollbackTo(savepoint).Error; rollbackErr != nil {
				return fmt.Errorf("nested transaction failed: %w, savepoint rollback failed: %v", err, rollbackErr)
			}
			return fmt.Errorf("nested transaction failed: %w", err)
		}

		return nil
	}

	// Not in a transaction, start a new one
	return WithTransactionContext(ctx, db, fn)
}
