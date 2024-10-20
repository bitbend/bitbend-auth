package eventstore

import (
	"github.com/shopspring/decimal"
	"time"
)

type ReadModel struct {
	TenantId    string          `json:"-"`
	AggregateId string          `json:"-"`
	Owner       string          `json:"-"`
	Events      []Event         `json:"-"`
	Position    decimal.Decimal `json:"-"`
	CreatedAt   time.Time       `json:"-"`
	UpdatedAt   time.Time       `json:"-"`
}

func (rm *ReadModel) Query() *QueryBuilder {
	return NewQuery()
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

	rm.Position = rm.Events[len(rm.Events)-1].GetPosition()
	rm.CreatedAt = rm.Events[0].GetCreatedAt()
	rm.UpdatedAt = rm.Events[len(rm.Events)-1].GetCreatedAt()

	rm.Events = rm.Events[0:0]

	return nil
}
