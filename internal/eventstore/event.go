package eventstore

import (
	"time"
)

type EventId string

func (ei EventId) String() string {
	return string(ei)
}

type EventType string

func (et EventType) String() string {
	return string(et)
}

type EventVersion uint

func (ev EventVersion) Unit() uint {
	return uint(ev)
}

type Event struct {
	Id            EventId
	Aggregate     *Aggregate
	Type          EventType
	Version       EventVersion
	Data          []byte
	Creator       string
	CorrelationId *string
	CausationId   *string
	Position      float64
	CreatedAt     time.Time
}
