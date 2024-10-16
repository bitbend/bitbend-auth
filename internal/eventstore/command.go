package eventstore

type Command interface {
	GetAggregate() *Aggregate
	GetType() EventType
	GetPayload() any
	GetCreator() *string
	GetCorrelationId() *string
	GetCausationId() *string
}
