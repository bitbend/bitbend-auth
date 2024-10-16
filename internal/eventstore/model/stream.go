package model

import (
	"time"
)

type Stream struct {
	TenantId       string    `db:"tenant_id" json:"tenant_id"`
	Id             string    `db:"id" json:"id"`
	StreamType     string    `db:"stream_type" json:"stream_type"`
	StreamVersion  int       `db:"stream_version" json:"stream_version"`
	StreamSequence int64     `db:"stream_sequence" json:"stream_sequence"`
	StreamOwner    string    `db:"stream_owner" json:"stream_owner"`
	CreatedAt      time.Time `db:"created_at" json:"created_at"`
}
