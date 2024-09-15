package database

import (
	"database/sql"
	"time"
)

type Event struct {
	TenantId          string         `db:"tenant_id" json:"tenant_id"`
	AggregateType     string         `db:"aggregate_type" json:"aggregate_type"`
	AggregateVersion  int            `db:"aggregate_version" json:"aggregate_version"`
	AggregateId       string         `db:"aggregate_id" json:"aggregate_id"`
	AggregateSequence int64          `db:"aggregate_sequence" json:"aggregate_sequence"`
	EventType         string         `db:"event_type" json:"event_type"`
	Data              []byte         `db:"data" json:"data"`
	Creator           sql.NullString `db:"creator" json:"creator"`
	ResourceOwner     sql.NullString `db:"resource_owner" json:"resource_owner"`
	CorrelationId     sql.NullString `db:"correlation_id" json:"correlation_id"`
	CausationId       sql.NullString `db:"causation_id" json:"causation_id"`
	GlobalPosition    float64        `db:"global_position" json:"global_position"`
	CreatedAt         time.Time      `db:"created_at" json:"created_at"`
}
