package database

import "time"

type DeletedRecord struct {
	Id        string    `db:"id" json:"id"`
	TenantId  string    `db:"tenant_id" json:"tenant_id"`
	TableName string    `db:"table_name" json:"table_name"`
	RecordId  string    `db:"record_id" json:"record_id"`
	Data      []byte    `db:"data" json:"data"`
	DeletedAt time.Time `db:"deleted_at" json:"deleted_at"`
}
