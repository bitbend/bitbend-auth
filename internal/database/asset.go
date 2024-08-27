package database

import "time"

type Asset struct {
	Id          string    `db:"id" json:"id"`
	TenantId    string    `db:"tenant_id" json:"tenant_id"`
	AssetType   string    `db:"asset_type" json:"asset_type"`
	Name        string    `db:"name" json:"name"`
	ContentType string    `db:"content_type" json:"content_type"`
	Data        []byte    `db:"data" json:"data"`
	Size        int64     `db:"size" json:"size"`
	Hash        string    `db:"hash" json:"hash"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time `db:"updated_at" json:"updated_at"`
}
