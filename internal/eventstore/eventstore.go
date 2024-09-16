package eventstore

import (
	"context"
	_ "embed"
	"fmt"
	"github.com/driftbase/auth/internal/sverror"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"sync"
	"time"
)

type EventStore struct {
	PushTimeout     time.Duration
	maxRetries      int
	tenants         []string
	lastTenantQuery time.Time
	tenantMutex     sync.Mutex
	db              *pgxpool.Pool
}

func (es *EventStore) withTxn(ctx context.Context, txOptions pgx.TxOptions, fn func(tx pgx.Tx, ctx context.Context) error) error {
	tx, err := es.db.BeginTx(ctx, txOptions)
	if err != nil {
		return sverror.NewInternalError("error.failed.to.begin.transaction", err)
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback(ctx)
			panic(p)
		} else if err != nil {
			if rbErr := tx.Rollback(ctx); rbErr != nil {
				err = sverror.NewInternalError("error.failed.to.rollback.transaction",
					fmt.Errorf("rollback.error: %w, internal.error: %w", rbErr, err),
				)
			}
		}
	}()

	err = fn(tx, ctx)
	if err != nil {
		return err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return sverror.NewInternalError("error.failed.to.commit.transaction", err)
	}

	return nil
}
