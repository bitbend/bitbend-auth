package eventstore

import (
	"encoding/json"
	"github.com/shopspring/decimal"
	"time"
)

type EventType string

var _ Event = (*EventBase)(nil)

type EventBase struct {
	Aggregate     *Aggregate      `json:"-"`
	Id            string          `json:"-"`
	Type          EventType       `json:"-"`
	Payload       []byte          `json:"-"`
	Creator       *string         `json:"-"`
	CorrelationId *string         `json:"-"`
	CausationId   *string         `json:"-"`
	Position      decimal.Decimal `json:"-"`
	CreatedAt     time.Time       `json:"-"`
}

func (eb *EventBase) GetAggregate() *Aggregate {
	return eb.Aggregate
}

func (eb *EventBase) GetId() string {
	return eb.Id
}

func (eb *EventBase) GetType() EventType {
	return eb.Type
}

func (eb *EventBase) GetPayload() []byte {
	return eb.Payload
}

func (eb *EventBase) UnmarshalPayload(ptr any) error {
	return json.Unmarshal(eb.Payload, ptr)
}

func (eb *EventBase) GetCreator() *string {
	return eb.Creator
}

func (eb *EventBase) GetCorrelationId() *string {
	return eb.CorrelationId
}

func (eb *EventBase) GetCausationId() *string {
	return eb.CausationId
}

func (eb *EventBase) GetPosition() decimal.Decimal {
	return eb.Position
}

func (eb *EventBase) GetCreatedAt() time.Time {
	return eb.CreatedAt
}

func EventBaseFromEvent(event Event) *EventBase {
	return &EventBase{
		Aggregate:     event.GetAggregate(),
		Id:            event.GetId(),
		Type:          event.GetType(),
		Payload:       event.GetPayload(),
		Creator:       event.GetCreator(),
		CorrelationId: event.GetCorrelationId(),
		CausationId:   event.GetCausationId(),
		Position:      event.GetPosition(),
		CreatedAt:     event.GetCreatedAt(),
	}
}
