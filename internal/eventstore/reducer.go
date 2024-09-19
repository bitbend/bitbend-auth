package eventstore

import (
	"context"
	"github.com/driftbase/auth/internal/database"
	"github.com/jackc/pgx/v5"
	"strconv"
)

type Reducer interface {
	Query() *SearchQueryBuilder
	AppendEvents(...Event)
	Reduce() error
}

func (es *EventStore) Reduce(ctx context.Context, reducer Reducer) error {
	stmt, args, err := reducer.Query().ToSql()
	if err != nil {
		return err
	}

	rows, err := es.db.Query(ctx, stmt, args)
	if err != nil {
		return err
	}

	eventRows, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[database.Event])
	if err != nil {
		return err
	}

	for _, eventRow := range eventRows {
		unmappedEvent := &event{
			aggregate: NewAggregate(
				TenantId(eventRow.TenantId),
				AggregateType(eventRow.AggregateType),
				AggregateVersion("v"+strconv.Itoa(eventRow.AggregateVersion)),
				AggregateId(eventRow.AggregateId),
				eventRow.ResourceOwner.String,
				uint64(eventRow.AggregateSequence),
			),
			eventType: EventType(eventRow.EventType),
			payload:   eventRow.Payload,
			createdAt: eventRow.CreatedAt,
		}
		if eventRow.CorrelationId.Valid {
			correlationId := eventRow.CorrelationId.String
			unmappedEvent.correlationId = &correlationId
		}
		if eventRow.CausationId.Valid {
			causationId := eventRow.CausationId.String
			unmappedEvent.causationId = &causationId
		}

		mappedEvent, err := es.mapEvent(unmappedEvent)
		if err != nil {
			return err
		}

		reducer.AppendEvents(mappedEvent)
	}

	return reducer.Reduce()
}
