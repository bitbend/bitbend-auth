package eventstore

import (
	"time"
)

type EventBase struct {
	Id                            EventId
	Type                          EventType
	Version                       EventVersion
	Aggregate                     *Aggregate
	previousAggregateSequence     uint64
	previousAggregateTypeSequence uint64
	Data                          []byte
	Creator                       string
	Position                      float64
	CreatedAt                     time.Time
}
