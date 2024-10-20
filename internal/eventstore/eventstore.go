package eventstore

import (
	"github.com/driftbase/auth/internal/database"
)

type EventStore struct {
	db *database.Db
}
