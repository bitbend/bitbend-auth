package eventstore

import (
	"github.com/driftbase/auth/internal/database"
)

type EventStore struct {
	db *database.Db
}

func NewEventStore(db *database.Db) *EventStore {
	return &EventStore{
		db: db,
	}
}
