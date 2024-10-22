package eventstore

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/oklog/ulid/v2"
	"github.com/stephenafamo/bob"
	"github.com/stephenafamo/bob/dialect/psql"
	"github.com/stephenafamo/bob/dialect/psql/im"
	"github.com/stephenafamo/bob/dialect/psql/sm"
	"time"
)

func (es *EventStore) Push(ctx context.Context, commands ...Command) ([]Event, error) {
	var sequences []*latestSequence
	var events []Event
	var err error

retry:
	for i := 0; i <= 10; i++ {
		err = es.db.WithTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted}, func(tx pgx.Tx) error {
			sequences, err = fetchLatestSequences(ctx, tx, commands)
			if err != nil {
				return err
			}

			events, err = commandsToEvents(commands, sequences)
			if err != nil {
				return err
			}

			err = createStreams(ctx, tx, events)
			if err != nil {
				return err
			}

			events, err = createEvents(ctx, tx, events)
			if err != nil {
				return err
			}

			return nil
		})
		if err != nil {
			var pgErr *pgconn.PgError
			if !errors.As(err, &pgErr) || pgErr.ConstraintName != "events_stream_uk" || pgErr.SQLState() != "23505" {
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

type latestSequence struct {
	aggregate *Aggregate
}

func fetchLatestSequences(ctx context.Context, tx pgx.Tx, commands []Command) ([]*latestSequence, error) {
	var expr []bob.Expression
	commandMap := make(map[string]Event)

	query := psql.Select(
		sm.From("streams"),
		sm.Columns(
			"tenant_id",
			"id",
			"stream_type",
			"stream_version",
			"stream_sequence",
			"stream_owner",
		),
		sm.ForUpdate(),
	)

	for _, command := range commands {
		aggregate := command.GetAggregate()
		if aggregate == nil {
			continue
		}

		key := fmt.Sprintf("%s:%s:%s", aggregate.TenantId, aggregate.Id, aggregate.Type)

		if _, exists := commandMap[key]; !exists {
			expr = append(expr, psql.ArgGroup(aggregate.TenantId, aggregate.Id, aggregate.Type))
		}
	}

	query.Apply(
		sm.Where(
			psql.Group(
				psql.Quote("tenant_id"),
				psql.Quote("id"),
				psql.Quote("stream_type"),
			).In(expr...),
		),
	)

	stmt, args, err := query.Build()
	if err != nil {
		return nil, err
	}

	rows, err := tx.Query(ctx, stmt, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sequences []*latestSequence
	for rows.Next() {
		var sequence latestSequence
		if err = rows.Scan(
			&sequence.aggregate.TenantId,
			&sequence.aggregate.Id,
			&sequence.aggregate.Type,
			&sequence.aggregate.Version,
			&sequence.aggregate.Sequence,
			&sequence.aggregate.Owner,
		); err != nil {
			return nil, err
		}
		sequences = append(sequences, &sequence)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return sequences, nil
}

func createStreams(ctx context.Context, tx pgx.Tx, events []Event) error {
	eventMap := make(map[string]Event)

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
		im.OnConflict("tenant_id", "id").DoNothing(),
	)

	for _, event := range events {
		aggregate := event.GetAggregate()
		if aggregate == nil {
			continue
		}

		key := fmt.Sprintf("%s:%s:%s", aggregate.TenantId, aggregate.Id, aggregate.Type)

		if _, exists := eventMap[key]; !exists {
			eventMap[key] = event

			query.Apply(
				im.Values(
					psql.Arg(
						aggregate.TenantId,
						aggregate.Id,
						aggregate.Type,
						aggregate.Version,
						aggregate.Sequence,
						aggregate.Owner,
						aggregate.CreatedAt,
					),
				),
			)
		}
	}

	stmt, args, err := query.Build()
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, stmt, args...)
	if err != nil {
		return err
	}

	return nil
}

func createEvents(ctx context.Context, tx pgx.Tx, events []Event) ([]Event, error) {
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
			"created_at",
		),
	)

	for _, event := range events {
		query.Apply(
			im.Values(
				psql.Arg(
					event.(*EventBase).Aggregate.TenantId,
					event.(*EventBase).Id,
					event.(*EventBase).Aggregate.Id,
					event.(*EventBase).Aggregate.Type,
					event.(*EventBase).Aggregate.Version,
					event.(*EventBase).Aggregate.Sequence,
					event.(*EventBase).Aggregate.Owner,
					event.(*EventBase).Type,
					event.(*EventBase).Payload,
					event.(*EventBase).Creator,
					event.(*EventBase).CorrelationId,
					event.(*EventBase).CausationId,
					event.(*EventBase).CreatedAt,
				),
			),
		)
	}

	stmt, args, err := query.Build()
	if err != nil {
		return nil, err
	}

	_, err = tx.Exec(ctx, stmt, args...)
	if err != nil {
		return nil, err
	}

	return events, nil
}

func commandToEvent(command Command, sequences *latestSequence) (Event, error) {
	payload, err := json.Marshal(command.GetPayload())
	if err != nil {
		return nil, err
	}

	return &EventBase{
		Aggregate:     sequences.aggregate,
		Id:            ulid.Make().String(),
		Type:          command.GetType(),
		Payload:       payload,
		Creator:       command.GetCreator(),
		CorrelationId: command.GetCorrelationId(),
		CausationId:   command.GetCausationId(),
		CreatedAt:     time.Now(),
	}, nil
}

func searchSequenceByCommand(sequences []*latestSequence, command Command) *latestSequence {
	for _, sequence := range sequences {
		if sequence.aggregate.TenantId == command.GetAggregate().TenantId &&
			sequence.aggregate.Id == command.GetAggregate().Id &&
			sequence.aggregate.Type == command.GetAggregate().Type {
			return sequence
		}
	}
	return nil
}

func commandsToEvents(commands []Command, sequences []*latestSequence) ([]Event, error) {
	events := make([]Event, len(commands))
	for _, command := range commands {
		sequence := searchSequenceByCommand(sequences, command)
		event, err := commandToEvent(command, sequence)
		if err != nil {
			return nil, err
		}
		events = append(events, event)
	}

	return events, nil
}
