package eventstore

import (
	"github.com/shopspring/decimal"
	"time"
)

type ReadModel struct {
	TenantId      TenantId
	AggregateId   AggregateId
	Events        []Event
	ResourceOwner string
	Sequence      uint64
	Position      decimal.Decimal
	CreatedAt     time.Time
	UpdatedAt     time.Time
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

	if rm.ResourceOwner == "" {
		rm.ResourceOwner = rm.Events[0].GetAggregate().ResourceOwner
	}

	rm.Sequence = rm.Events[len(rm.Events)-1].GetAggregate().Sequence

	rm.Position = rm.Events[len(rm.Events)-1].GetPosition()

	if rm.CreatedAt.IsZero() {
		rm.CreatedAt = rm.Events[0].GetCreatedAt()
	}

	rm.UpdatedAt = rm.Events[len(rm.Events)-1].GetCreatedAt()

	rm.Events = rm.Events[0:0]

	return nil
}
