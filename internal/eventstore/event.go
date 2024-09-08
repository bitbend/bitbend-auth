package eventstore

import (
	"time"
)

type EventType string

func (et EventType) String() string {
	return string(et)
}

type action interface {
	GetAggregate() *Aggregate
	GetCreator() string
	GetEventType() EventType
	GetCorrelationId() *string
	GetCausationId() *string
}

type Command interface {
	action
	GetPayload() any
	GetUniqueConstraints() []*UniqueConstraint
}

type Event interface {
	action
	GetPosition() float64
	GetCreatedAt() time.Time
	UnmarshalData(ptr any) error
}
