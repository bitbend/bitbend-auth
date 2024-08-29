package database

import (
	"database/sql"
	"time"
)

type UniqueConstraint struct {
	Id          string         `db:"id" json:"id"`
	TenantId    sql.NullString `db:"tenant_id" json:"tenant_id"`
	UniqueType  string         `db:"unique_type" json:"unique_type"`
	UniqueValue string         `db:"unique_value" json:"unique_value"`
	CreatedAt   time.Time      `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time      `db:"updated_at" json:"updated_at"`
}
