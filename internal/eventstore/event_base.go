package eventstore

import (
	"encoding/json"
	"time"
)

var _ Event = (*EventBase)(nil)

type EventBase struct {
	Id                            EventId
	Type                          EventType
	Version                       EventVersion
	Aggregate                     *Aggregate
	previousAggregateSequence     uint64
	previousAggregateTypeSequence uint64
	Data                          []byte
	Creator                       string
	CorrelationId                 *string
	CausationId                   *string
	Position                      float64
	CreatedAt                     time.Time
}

func (eb *EventBase) SetCorrelationId(correlationId string) {
	eb.CorrelationId = &correlationId
}

func (eb *EventBase) SetCausationId(causationId string) {
	eb.CausationId = &causationId
}

func (eb *EventBase) GetId() EventId {
	return eb.Id
}

func (eb *EventBase) GetAggregate() *Aggregate {
	return eb.Aggregate
}

func (eb *EventBase) GetType() EventType {
	return eb.Type
}

func (eb *EventBase) GetVersion() EventVersion {
	return eb.Version
}

func (eb *EventBase) GetData() []byte {
	return eb.Data
}

func (eb *EventBase) GetCreator() string {
	return eb.Creator
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

func (eb *EventBase) Unmarshal(ptr any) error {
	return json.Unmarshal(eb.Data, ptr)
}
