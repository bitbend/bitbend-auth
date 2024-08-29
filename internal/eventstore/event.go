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

type Event interface {
	GetId() EventId
	GetAggregate() *Aggregate
	GetType() EventType
	GetVersion() EventVersion
	GetData() []byte
	GetCreator() string
	GetCorrelationId() *string
	GetCausationId() *string
	GetPosition() float64
	GetCreatedAt() time.Time
	Unmarshal(ptr any) error
}
