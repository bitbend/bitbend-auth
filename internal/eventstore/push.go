package eventstore

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"github.com/driftbase/auth/internal/sverror"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/oklog/ulid/v2"
	"github.com/stephenafamo/bob/dialect/psql"
	"github.com/stephenafamo/bob/dialect/psql/im"
)

func (es *EventStore) Push(ctx context.Context, commands ...Command) (events []Event, err error) {
	var sequences []*latestSequence

	if es.PushTimeout > 0 {
		var cancel func()
		ctx, cancel = context.WithTimeout(ctx, es.PushTimeout)
		defer cancel()
	}

retry:
	for i := 0; i <= es.maxRetries; i++ {
		err = es.withTxn(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted}, func(tx pgx.Tx, ctx context.Context) error {
			sequences, err = latestSequences(ctx, tx, commands)
			if err != nil {
				return err
			}

			events, err = mapCommandsToEvents(commands, sequences)

			if err = insertStreams(ctx, tx, events); err != nil {
				return err
			}

			events, err = insertEvents(ctx, tx, events)
			if err != nil {
				return err
			}

			if err = handleUniqueConstraints(ctx, tx, commands); err != nil {
				return err
			}

			return nil
		})
		if err != nil {
			var pgErr *pgconn.PgError
			if !errors.As(err, &pgErr) || pgErr.ConstraintName != "events_pk" || pgErr.SQLState() != "23505" {
				break retry
			}
		}
	}

	mappedEvents, err := es.mapEvents(events)
	if err != nil {
		return nil, err
	}

	es.notify(mappedEvents)

	return mappedEvents, nil
}

func insertStreams(ctx context.Context, tx pgx.Tx, events []Event) error {
	streams := eventsToAggregate(events)

	query := psql.Insert(
		im.Into(
			"streams",
			"tenant_id",
			"id",
			"stream_type",
			"stream_version",
			"stream_sequence",
			"stream_owner",
			"created_at",
		),
		im.OnConflict("id", "stream_type", "stream_sequence").DoNothing(),
	)

	for _, stream := range streams {
		streamVersion, err := stream.Version.Int()
		if err != nil {
			return err
		}

		query.Apply(
			im.Values(
				psql.Arg(
					stream.TenantId.String(),
					stream.Id.String(),
					stream.Type.String(),
					streamVersion,
					stream.Sequence,
					stream.ResourceOwner,
				),
				psql.Raw("statement_timestamp()"),
			),
		)
	}

	stmt, args, err := query.Build()
	if err != nil {
		return err
	}

	rows, err := tx.Query(ctx, stmt, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	if err := rows.Err(); err != nil {
		var pgError *pgconn.PgError
		if errors.As(err, &pgError) {
			if pgError.Code == "40001" {
				return sverror.NewInternalError("error.transaction.conflict", err)
			}
		}
		return sverror.NewInternalError("error.failed.to.insert.streams", err)
	}

	return nil
}

func insertEvents(ctx context.Context, tx pgx.Tx, events []Event) ([]Event, error) {
	query := psql.Insert(
		im.Into(
			"events",
			"tenant_id",
			"id",
			"stream_id",
			"stream_type",
			"stream_version",
			"stream_sequence",
			"stream_owner",
			"event_type",
			"payload",
			"creator",
			"correlation_id",
			"causation_id",
			"global_position",
			"created_at",
		),
		im.Returning("global_position", "created_at"),
	)

	for i, _ := range events {
		aggregateVersion, err := events[i].(*EventBase).Aggregate.Version.Int()
		if err != nil {
			return nil, err
		}

		query.Apply(
			im.Values(
				psql.Arg(
					events[i].(*EventBase).Aggregate.TenantId.String(),
					ulid.Make(),
					events[i].(*EventBase).Aggregate.Id.String(),
					events[i].(*EventBase).Aggregate.Type.String(),
					aggregateVersion,
					events[i].(*EventBase).Aggregate.Sequence,
					events[i].(*EventBase).Aggregate.ResourceOwner,
					events[i].(*EventBase).EventType.String(),
					events[i].(*EventBase).Payload,
					events[i].(*EventBase).Creator,
					events[i].(*EventBase).CorrelationId,
					events[i].(*EventBase).CausationId,
				),
				psql.Raw("extract(epoch from clock_timestamp())"),
				psql.Raw("statement_timestamp()"),
			),
		)
	}

	stmt, args, err := query.Build()
	if err != nil {
		return nil, err
	}

	rows, err := tx.Query(ctx, stmt, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for i := 0; rows.Next(); i++ {
		err = rows.Scan(&events[i].(*EventBase).CreatedAt, &events[i].(*EventBase).Position)
		if err != nil {
			return nil, err
		}
	}

	if err := rows.Err(); err != nil {
		var pgError *pgconn.PgError
		if errors.As(err, &pgError) {
			if pgError.Code == "40001" {
				return nil, sverror.NewInternalError("error.transaction.conflict", err)
			}
		}
		return nil, sverror.NewInternalError("error.failed.to.insert.events", err)
	}

	return events, nil
}

func mapCommandsToEvents(commands []Command, sequences []*latestSequence) ([]Event, error) {
	events := make([]Event, 0)

	for _, command := range commands {
		sequence := searchSequenceByCommand(sequences, command)
		if sequence == nil {
			return nil, nil
		}
		sequence.aggregate.Sequence++

		var commandEvent Event
		commandEvent, err := mapCommandToEvent(sequence, command)
		if err != nil {
			return nil, err
		}
		events = append(events, commandEvent)
	}

	return events, nil
}

func eventsToAggregate(events []Event) []*Aggregate {
	uniqueAggregates := make(map[string]*Aggregate)
	var aggregates []*Aggregate

	for _, evt := range events {
		eventAggregate := evt.GetAggregate()
		if eventAggregate == nil {
			continue
		}

		aggregateKey := eventAggregate.Type.String() + ":" + eventAggregate.Id.String()

		if _, exists := uniqueAggregates[aggregateKey]; exists {
			continue
		}

		uniqueAggregates[aggregateKey] = eventAggregate

		aggregates = append(aggregates, eventAggregate)
	}

	return aggregates
}

func mapCommandToEvent(sequence *latestSequence, command Command) (*EventBase, error) {
	eventBase := &EventBase{
		Aggregate: sequence.aggregate,
		EventType: command.GetEventType(),
	}

	if command.GetPayload() != nil {
		payload, err := json.Marshal(command.GetPayload())
		if err != nil {
			return nil, sverror.NewInternalError("error.failed.to.marshal.command.payload", err)
		}
		eventBase.Payload = payload
	}

	if command.GetCreator() != nil {
		eventBase.Creator = command.GetCreator()
	}

	if command.GetCorrelationId() != nil {
		eventBase.CorrelationId = command.GetCorrelationId()
	}

	if command.GetCausationId() != nil {
		eventBase.CausationId = command.GetCausationId()
	}

	return eventBase, nil
}
