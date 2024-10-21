package database

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Db struct {
	pool *pgxpool.Pool
}

func (db *Db) Pool() *pgxpool.Pool {
	return db.pool
}

func (db *Db) WithTx(ctx context.Context, txOptions pgx.TxOptions, fn func(tx pgx.Tx) error) error {
	tx, err := db.Pool().BeginTx(ctx, txOptions)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback(ctx)
			panic(p)
		} else if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	err = fn(tx)
	if err != nil {
		return fmt.Errorf("failed transaction execution: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
