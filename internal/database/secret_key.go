package database

import (
	"time"
)

type SecretKey struct {
	TenantId    string    `db:"tenant_id" json:"tenant_id"`
	SecretType  string    `db:"secret_type" json:"secret_type"`
	SecretValue string    `db:"secret_value" json:"secret_value"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time `db:"updated_at" json:"updated_at"`
}
