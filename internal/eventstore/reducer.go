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
		unmappedEvent := &EventBase{
			Aggregate: NewAggregate(
				TenantId(eventRow.TenantId),
				AggregateType(eventRow.StreamType),
				AggregateVersion("v"+strconv.Itoa(eventRow.StreamVersion)),
				AggregateId(eventRow.StreamId),
				eventRow.StreamOwner,
				uint64(eventRow.StreamSequence),
			),
			EventType: EventType(eventRow.EventType),
			Payload:   eventRow.Payload,
			CreatedAt: eventRow.CreatedAt,
		}
		if eventRow.Creator.Valid {
			creator := eventRow.Creator.String
			unmappedEvent.Creator = &creator
		}
		if eventRow.CorrelationId.Valid {
			correlationId := eventRow.CorrelationId.String
			unmappedEvent.CorrelationId = &correlationId
		}
		if eventRow.CausationId.Valid {
			causationId := eventRow.CausationId.String
			unmappedEvent.CausationId = &causationId
		}

		mappedEvent, err := es.mapEvent(unmappedEvent)
		if err != nil {
			return err
		}

		reducer.AppendEvents(mappedEvent)
	}

	return reducer.Reduce()
}
