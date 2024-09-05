package eventstore

import (
	"encoding/json"
	"time"
)

var _ Event = (*EventBase)(nil)

type EventBase struct {
	Aggregate                     *Aggregate `json:"-"`
	EventType                     EventType  `json:"-"`
	previousAggregateSequence     uint64
	previousAggregateTypeSequence uint64
	Data                          []byte    `json:"-"`
	Creator                       string    `json:"-"`
	CorrelationId                 *string   `json:"-"`
	CausationId                   *string   `json:"-"`
	Position                      float64   `json:"-"`
	CreatedAt                     time.Time `json:"-"`
}

func (eb *EventBase) SetCorrelationId(correlationId string) *EventBase {
	eb.CorrelationId = &correlationId
	return eb
}

func (eb *EventBase) SetCausationId(causationId string) *EventBase {
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
