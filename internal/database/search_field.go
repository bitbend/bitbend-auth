package database

import "time"

type SearchField struct {
	Id               string    `db:"id" json:"id"`
	TenantId         string    `db:"tenant_id" json:"tenant_id"`
	AggregateType    string    `db:"aggregate_type" json:"aggregate_type"`
	AggregateVersion int       `db:"aggregate_version" json:"aggregate_version"`
	AggregateId      string    `db:"aggregate_id" json:"aggregate_id"`
	FieldName        string    `db:"field_name" json:"field_name"`
	FieldValue       string    `db:"field_value" json:"field_value"`
	CreatedAt        time.Time `db:"created_at" json:"created_at"`
	UpdatedAt        time.Time `db:"updated_at" json:"updated_at"`
}
