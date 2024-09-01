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

type action interface {
	GetAggregate() *Aggregate
	GetCreator() string
	GetEventType() EventType
}

type Command interface {
	action
	GetPayload() any
	GetUniqueConstraints() []*UniqueConstraint
}

type Event interface {
	action
	GetCorrelationId() *string
	GetCausationId() *string
	GetPosition() float64
	GetCreatedAt() time.Time
	UnmarshalData(ptr any) error
}
