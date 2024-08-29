package database

import (
	"database/sql"
	"time"
)

type Snapshot struct {
	Id                string         `db:"id" json:"id"`
	TenantId          sql.NullString `db:"tenant_id" json:"tenant_id"`
	AggregateType     string         `db:"aggregate_type" json:"aggregate_type"`
	AggregateVersion  int            `db:"aggregate_version" json:"aggregate_version"`
	AggregateId       string         `db:"aggregate_id" json:"aggregate_id"`
	AggregateSequence int64          `db:"aggregate_sequence" json:"aggregate_sequence"`
	Data              []byte         `db:"data" json:"data"`
	CreatedAt         time.Time      `db:"created_at" json:"created_at"`
	UpdatedAt         time.Time      `db:"updated_at" json:"updated_at"`
}
