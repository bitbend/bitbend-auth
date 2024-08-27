package eventstore

import (
	"time"
)

type WriteModel struct {
	TenantId          TenantId
	AggregateId       AggregateId
	Events            []Event
	Owner             string
	Creator           string
	ProcessedSequence uint64
	ChangedAt         time.Time
}

func (wm *WriteModel) AppendEvents(events ...Event) {
	wm.Events = append(wm.Events, events...)
}

func (wm *WriteModel) Reduce() error {
	if len(wm.Events) == 0 {
		return nil
	}

	if wm.TenantId == "" {
		wm.TenantId = wm.Events[0].Aggregate.TenantId
	}

	if wm.AggregateId == "" {
		wm.AggregateId = wm.Events[0].Aggregate.Id
	}

	if wm.Owner == "" {
		wm.Owner = wm.Events[0].Aggregate.Owner
	}

	if wm.Creator == "" {
		wm.Creator = wm.Events[0].Aggregate.Creator
	}

	wm.ProcessedSequence = wm.Events[len(wm.Events)-1].Aggregate.Sequence
	wm.ChangedAt = wm.Events[len(wm.Events)-1].CreatedAt

	wm.Events = nil
	wm.Events = []Event{}

	return nil
}
