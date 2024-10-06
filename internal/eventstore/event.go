package eventstore

import (
	"github.com/shopspring/decimal"
	"time"
)

type EventType string

func (et EventType) String() string {
	return string(et)
}

type action interface {
	GetAggregate() *Aggregate
	GetCreator() *string
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
	GetPosition() decimal.Decimal
	GetCreatedAt() time.Time
	UnmarshalPayload(ptr any) error
}
