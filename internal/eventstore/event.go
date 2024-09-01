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
	Aggregate() *Aggregate
	Creator() string
	Type() EventType
	Version() uint
}

type Command interface {
	action
	Payload() any
	UniqueConstraints() []*UniqueConstraint
}

type Event interface {
	action
	CorrelationId() *string
	CausationId() *string
	Position() float64
	CreatedAt() time.Time
	UnmarshalData(ptr any) error
}
