package database

import (
	"database/sql"
	"time"
)

type AuditLog struct {
	Id        string         `db:"id" json:"id"`
	TenantId  sql.NullString `db:"tenant_id" json:"tenant_id"`
	Name      string         `db:"name" json:"name"`
	Data      []byte         `db:"data" json:"data"`
	CreatedAt time.Time      `db:"created_at" json:"created_at"`
}
