package eventstore

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/oklog/ulid/v2"
	"github.com/stephenafamo/bob/dialect/psql"
	"github.com/stephenafamo/bob/dialect/psql/im"
	"time"
)

func (es *EventStore) Push(ctx context.Context, commands []Command) ([]Event, error) {
	var err error
	events, err := commandsToEvents(commands)
	if err != nil {
		return nil, err
	}

	err = es.WithTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted}, func(ctx context.Context, tx pgx.Tx) error {
		err = insertStreams(ctx, tx, events)
		if err != nil {
			return err
		}

		events, err = insertEvents(ctx, tx, events)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	mappedEvents, err := es.mapEvents(events)
	if err != nil {
		return nil, err
	}

	return mappedEvents, nil
}

func insertStreams(ctx context.Context, tx pgx.Tx, events []Event) error {
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

		key := fmt.Sprintf("%s:%s", aggregate.TenantId, aggregate.Id)

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
			"created_at",
		),
	)

	for _, event := range events {
		query.Apply(
			im.Values(
				psql.Arg(
					event.GetAggregate().TenantId,
					event.GetId(),
					event.GetAggregate().Id,
					event.GetAggregate().Type,
					event.GetAggregate().Version,
					event.GetAggregate().Sequence,
					event.GetAggregate().Owner,
					event.GetType(),
					event.GetPayload(),
					event.GetCreator(),
					event.GetCorrelationId(),
					event.GetCausationId(),
					event.GetCreatedAt(),
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

func commandToEvent(command Command) (Event, error) {
	payload, err := json.Marshal(command.GetPayload())
	if err != nil {
		return nil, err
	}

	return &EventBase{
		Aggregate:     command.GetAggregate(),
		Id:            ulid.Make().String(),
		Type:          command.GetType(),
		Payload:       payload,
		Creator:       command.GetCreator(),
		CorrelationId: command.GetCorrelationId(),
		CausationId:   command.GetCausationId(),
		CreatedAt:     time.Now(),
	}, nil
}

func commandsToEvents(commands []Command) ([]Event, error) {
	events := make([]Event, len(commands))
	for _, command := range commands {
		event, err := commandToEvent(command)
		if err != nil {
			return nil, err
		}
		events = append(events, event)
	}

	return events, nil
}
