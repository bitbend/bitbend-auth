package eventstore

import (
	"time"
)

type ReadModel struct {
	TenantId          TenantId
	AggregateId       AggregateId
	Events            []Event
	Owner             string
	Position          float64
	ProcessedSequence uint64
	CreatedAt         time.Time
	ChangedAt         time.Time
}

func (rm *ReadModel) AppendEvents(events ...Event) {
	rm.Events = append(rm.Events, events...)
}

func (rm *ReadModel) Reduce() error {
	if len(rm.Events) == 0 {
		return nil
	}

	if rm.TenantId == "" {
		rm.TenantId = rm.Events[0].GetAggregate().TenantId
	}

	if rm.AggregateId == "" {
		rm.AggregateId = rm.Events[0].GetAggregate().Id
	}

	if rm.Owner == "" {
		rm.Owner = rm.Events[0].GetAggregate().Owner
	}

	if rm.CreatedAt.IsZero() {
		rm.CreatedAt = rm.Events[0].GetCreatedAt()
	}

	rm.ChangedAt = rm.Events[len(rm.Events)-1].GetCreatedAt()
	rm.ProcessedSequence = rm.Events[len(rm.Events)-1].GetAggregate().Sequence
	rm.Position = rm.Events[len(rm.Events)-1].GetPosition()

	rm.Events = rm.Events[0:0]

	return nil
}
