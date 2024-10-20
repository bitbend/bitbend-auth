package eventstore

import "context"

type Reducer interface {
	Query() *QueryBuilder
	AppendEvents(events ...Event)
	Reduce() error
}

func (es *EventStore) Reduce(ctx context.Context, reducer Reducer) error {
	stmt, args, err := reducer.Query().ToSQL()
	if err != nil {
		return err
	}

	rows, err := es.db.Pool.Query(ctx, stmt, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		event := new(EventBase)
		if err := rows.Scan(
			&event.Aggregate.TenantId,
			&event.Id,
			&event.Aggregate.Id,
			&event.Aggregate.Type,
			&event.Aggregate.Version,
			&event.Aggregate.Sequence,
			&event.Aggregate.Owner,
			&event.Payload,
			&event.CorrelationId,
			&event.CausationId,
			&event.CreatedAt,
		); err != nil {
			return err
		}

		mappedEvent, err := es.mapEvent(event)
		if err != nil {
			return err
		}

		reducer.AppendEvents(mappedEvent)
	}

	return reducer.Reduce()
}
