package model

import (
	"github.com/shopspring/decimal"
	"time"
)

type Event struct {
	TenantId       string          `db:"tenant_id" json:"tenant_id"`
	Id             string          `db:"id" json:"id"`
	StreamId       string          `db:"stream_id" json:"stream_id"`
	StreamType     string          `db:"stream_type" json:"stream_type"`
	StreamVersion  int             `db:"stream_version" json:"stream_version"`
	StreamSequence int64           `db:"stream_sequence" json:"stream_sequence"`
	StreamOwner    string          `db:"stream_owner" json:"stream_owner"`
	EventType      string          `db:"event_type" json:"event_type"`
	Payload        []byte          `db:"payload" json:"payload"`
	Creator        string          `db:"creator" json:"creator"`
	CorrelationId  string          `db:"correlation_id" json:"correlation_id"`
	CausationId    string          `db:"causation_id" json:"causation_id"`
	GlobalPosition decimal.Decimal `db:"global_position" json:"global_position"`
	CreatedAt      time.Time       `db:"created_at" json:"created_at"`
}
