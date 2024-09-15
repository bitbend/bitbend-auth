package eventstore

import (
	"time"
)

type WriteModel struct {
	TenantId      TenantId
	AggregateId   AggregateId
	Events        []Event
	ResourceOwner string
	Sequence      uint64
	UpdatedAt     time.Time
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

	if wm.ResourceOwner == "" {
		wm.ResourceOwner = wm.Events[0].GetAggregate().ResourceOwner
	}

	wm.Sequence = wm.Events[len(wm.Events)-1].GetAggregate().Sequence

	wm.UpdatedAt = wm.Events[len(wm.Events)-1].GetCreatedAt()

	wm.Events = nil
	wm.Events = []Event{}

	return nil
}
