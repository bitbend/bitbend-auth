package eventstore

import (
	"github.com/shopspring/decimal"
	"time"
)

type common interface {
	GetAggregate() *Aggregate
	GetType() EventType
	GetCreator() *string
	GetCorrelationId() *string
	GetCausationId() *string
}

type Command interface {
	common
	GetPayload() any
}

type Event interface {
	common
	GetId() string
	GetPayloadBytes() []byte
	UnmarshalPayload(ptr any) error
	GetPosition() decimal.Decimal
	GetCreatedAt() time.Time
}
