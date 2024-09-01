package eventstore

import (
	"encoding/json"
	"time"
)

var _ Event = (*EventBase)(nil)

type EventBase struct {
	Aggregate                     *Aggregate
	EventType                     EventType
	previousAggregateSequence     uint64
	previousAggregateTypeSequence uint64
	Data                          []byte
	Creator                       string
	CorrelationId                 *string
	CausationId                   *string
	Position                      float64
	CreatedAt                     time.Time
}

func (eb *EventBase) WithCorrelationId(correlationId string) *EventBase {
	eb.CorrelationId = &correlationId
	return eb
}

func (eb *EventBase) WithCausationId(causationId string) *EventBase {
	eb.CausationId = &causationId
	return eb
}

func (eb *EventBase) GetAggregate() *Aggregate {
	return eb.Aggregate
}

func (eb *EventBase) GetCreator() string {
	return eb.Creator
}

func (eb *EventBase) GetEventType() EventType {
	return eb.EventType
}

func (eb *EventBase) GetCorrelationId() *string {
	return eb.CorrelationId
}

func (eb *EventBase) GetCausationId() *string {
	return eb.CausationId
}

func (eb *EventBase) GetPosition() float64 {
	return eb.Position
}

func (eb *EventBase) GetCreatedAt() time.Time {
	return eb.CreatedAt
}

func (eb *EventBase) UnmarshalData(ptr any) error {
	return json.Unmarshal(eb.Data, ptr)
}
