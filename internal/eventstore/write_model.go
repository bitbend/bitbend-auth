package eventstore

import (
	"github.com/shopspring/decimal"
	"time"
)

var _ Reducer = (*WriteModel)(nil)

type WriteModel struct {
	TenantId          string          `json:"-"`
	AggregateId       string          `json:"-"`
	AggregateSequence int64           `json:"-"`
	Owner             string          `json:"-"`
	Events            []Event         `json:"-"`
	Position          decimal.Decimal `json:"-"`
	CreatedAt         time.Time       `json:"-"`
	UpdatedAt         time.Time       `json:"-"`
}

func (wm *WriteModel) Query() *QueryBuilder {
	return NewQuery()
}

func (wm *WriteModel) AppendEvents(events ...Event) {
	wm.Events = append(wm.Events, events...)
}

func (wm *WriteModel) Reduce() error {
	if len(wm.Events) == 0 {
		return nil
	}

	if wm.TenantId == "" {
		wm.TenantId = wm.Events[0].GetAggregate().TenantId
	}

	if wm.AggregateId == "" {
		wm.AggregateId = wm.Events[0].GetAggregate().Id
	}

	if wm.AggregateSequence == 0 {
		wm.AggregateSequence = wm.Events[len(wm.Events)-1].GetAggregate().Sequence
	}

	if wm.Owner == "" {
		wm.Owner = wm.Events[0].GetAggregate().Owner
	}

	wm.Position = wm.Events[len(wm.Events)-1].GetPosition()
	wm.CreatedAt = wm.Events[0].GetCreatedAt()
	wm.UpdatedAt = wm.Events[len(wm.Events)-1].GetCreatedAt()

	wm.Events = wm.Events[0:0]

	return nil
}
