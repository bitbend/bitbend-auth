package eventstore

import (
	"context"
	_ "embed"
	"fmt"
	"github.com/driftbase/auth/internal/sverror"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
	"github.com/stephenafamo/bob/dialect/psql"
	"github.com/stephenafamo/bob/dialect/psql/sm"
	"sync"
	"time"
)

type EventStore struct {
	PushTimeout     time.Duration
	maxRetries      int
	tenantIds       []TenantId
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

func (es *EventStore) GetTenantIds(ctx context.Context, fromDb bool) ([]TenantId, error) {
	if !fromDb {
		return es.tenantIds, nil
	}

	stmt, args, err := psql.Select(
		sm.From("events"),
		sm.Columns("tenant_id"),
		sm.Distinct("tenant_id"),
	).Build()
	if err != nil {
		return nil, sverror.NewInternalError("error.failed.to.build.query", err)
	}

	rows, err := es.db.Query(ctx, stmt, args...)
	if err != nil {
		return nil, sverror.NewInternalError("error.failed.to.query", err)
	}

	tenantIdRows, err := pgx.CollectRows(rows, pgx.RowTo[string])
	if err != nil {
		return nil, sverror.NewInternalError("error.failed.to.collect.rows", err)
	}

	tenantIds := make([]TenantId, 0)
	for _, tenantIdRow := range tenantIdRows {
		tenantIds = append(tenantIds, TenantId(tenantIdRow))
	}

	return tenantIds, nil
}

func (es *EventStore) GetCurrentGlobalPosition(ctx context.Context) (*decimal.Decimal, error) {
	stmt, args, err := psql.Select(
		sm.From("events"),
		sm.Columns(
			psql.Raw("max(global_position)").As("global_position"),
		),
	).Build()
	if err != nil {
		return nil, sverror.NewInternalError("error.failed.to.build.query", err)
	}

	row := es.db.QueryRow(ctx, stmt, args...)

	globalPosition := new(decimal.Decimal)
	err = row.Scan(globalPosition)
	if err != nil {
		return nil, sverror.NewInternalError("error.failed.to.scan.row", err)
	}

	return globalPosition, nil
}
