package repository

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
)

// StartTransaction return a tx
func StartTransaction(db *pgxpool.Pool, ctx context.Context) (pgx.Tx, error) {
	tx, err := db.Begin(ctx)
	if err != nil {
		return nil, err
	}

	return tx, nil
}

// DeferRollback rollback the transaction if it's not already closed
func DeferRollback(tx pgx.Tx, e echo.Context) {
	if err := tx.Rollback(e.Request().Context()); err != nil && err != pgx.ErrTxClosed {
		e.Logger().Error("Failed to rollback transaction", err)
	}
}

// CommitTransaction commit the transaction
func CommitTransaction(tx pgx.Tx, e echo.Context) error {
	if err := tx.Commit(e.Request().Context()); err != nil {
		e.Logger().Error("Failed to commit transaction", err)
		return err
	}
	return nil
}
