package model

import (
	"github.com/shopspring/decimal"
	"time"
)

type Field struct {
	TenantId         string          `db:"tenant_id" json:"tenant_id"`
	Id               string          `db:"id" json:"id"`
	StreamId         string          `db:"stream_id" json:"stream_id"`
	StreamType       string          `db:"stream_type" json:"stream_type"`
	StreamVersion    int             `db:"stream_version" json:"stream_version"`
	StreamSequence   int64           `db:"stream_sequence" json:"stream_sequence"`
	StreamOwner      string          `db:"stream_owner" json:"stream_owner"`
	FieldName        string          `db:"field_name" json:"field_name"`
	FieldValueText   string          `db:"field_value_text" json:"field_value_text"`
	FieldValueNumber decimal.Decimal `db:"field_value_number" json:"field_value_number"`
	FieldValueBool   bool            `db:"field_value_bool" json:"field_value_bool"`
	IsIndexed        bool            `db:"is_indexed" json:"is_indexed"`
	IsUnique         bool            `db:"is_unique" json:"is_unique"`
	CreatedAt        time.Time       `db:"created_at" json:"created_at"`
	UpdatedAt        time.Time       `db:"updated_at" json:"updated_at"`
}
