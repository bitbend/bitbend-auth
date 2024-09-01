package eventstore

import (
	"encoding/json"
	"time"
)

var _ Event = (*EventBase)(nil)

type EventBase struct {
	Agg                           *Aggregate
	Typ                           EventType
	previousAggregateSequence     uint64
	previousAggregateTypeSequence uint64
	Data                          []byte
	User                          string
	Correlation                   *string
	Causation                     *string
	Pos                           float64
	Create                        time.Time
}

func (eb *EventBase) WithCorrelationId(correlationId string) *EventBase {
	eb.Correlation = &correlationId
	return eb
}

func (eb *EventBase) WithCausationId(causationId string) *EventBase {
	eb.Causation = &causationId
	return eb
}

func (eb *EventBase) Aggregate() *Aggregate {
	return eb.Agg
}

func (eb *EventBase) Creator() string {
	return eb.User
}

func (eb *EventBase) Type() EventType {
	return eb.Typ
}

func (eb *EventBase) Version() uint {
	return eb.Agg.Version.Uint()
}

func (eb *EventBase) CorrelationId() *string {
	return eb.Correlation
}

func (eb *EventBase) CausationId() *string {
	return eb.Causation
}

func (eb *EventBase) Position() float64 {
	return eb.Pos
}

func (eb *EventBase) CreatedAt() time.Time {
	return eb.Create
}

func (eb *EventBase) UnmarshalData(ptr any) error {
	return json.Unmarshal(eb.Data, ptr)
}
