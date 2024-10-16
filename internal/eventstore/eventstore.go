package eventstore

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type EventStore struct {
	tenantIds []string
	db        pgxpool.Pool
}

func (es *EventStore) WithTx(ctx context.Context, txOptions pgx.TxOptions, fn func(tx pgx.Tx) error) error {
	tx, err := es.db.BeginTx(ctx, txOptions)
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
