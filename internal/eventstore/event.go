package eventstore

import (
	"github.com/shopspring/decimal"
	"time"
)

type Event interface {
	GetAggregate() *Aggregate
	GetId() string
	GetType() EventType
	GetPayloadBytes() []byte
	UnmarshalPayload(ptr any) error
	GetCreator() *string
	GetCorrelationId() *string
	GetCausationId() *string
	GetPosition() decimal.Decimal
	GetCreatedAt() time.Time
}
