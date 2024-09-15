package eventstore

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/shopspring/decimal"
	"strings"
	"time"
)

const pushPlaceholder = "($%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, extract(epoch from clock_timestamp()), statement_timestamp())"

//go:embed push.sql
var pushStmt string

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

			events, err = insertEvents(ctx, tx, sequences, commands)
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

func insertEvents(ctx context.Context, tx pgx.Tx, sequences []*latestSequence, commands []Command) ([]Event, error) {
	events, placeholders, args, err := mapCommands(commands, sequences)
	if err != nil {
		return nil, err
	}

	rows, err := tx.Query(ctx, fmt.Sprintf(pushStmt, strings.Join(placeholders, ", ")), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for i := 0; rows.Next(); i++ {
		err = rows.Scan(&events[i].(*event).createdAt, &events[i].(*event).position)
		if err != nil {
			return nil, err
		}
	}

	if err := rows.Err(); err != nil {
		var pgError *pgconn.PgError
		if errors.As(err, &pgError) {
			if pgError.Code == "40001" {
				return nil, fmt.Errorf("transaction conflict: %w", err)
			}
		}
		return nil, fmt.Errorf("failed to insert events: %w", err)
	}

	return events, nil
}

func mapCommands(commands []Command, sequences []*latestSequence) (events []Event, placeholders []string, args []any, err error) {
	events = make([]Event, 0)
	args = make([]any, 0)
	placeholders = make([]string, 0)

	for i, command := range commands {
		sequence := searchSequenceByCommand(sequences, command)
		if sequence == nil {
			return nil, nil, nil, nil
		}
		sequence.aggregate.Sequence++

		var commandEvent Event
		commandEvent, err = commandToEvent(sequence, command)
		if err != nil {
			return nil, nil, nil, err
		}
		events = append(events, commandEvent)

		placeholders = append(placeholders, fmt.Sprintf(pushPlaceholder,
			len(args)+1,
			len(args)+2,
			len(args)+3,
			len(args)+4,
			len(args)+5,
			len(args)+6,
			len(args)+7,
			len(args)+8,
			len(args)+9,
			len(args)+10,
			len(args)+11,
		))

		aggregateVersion, err := events[i].(*event).aggregate.Version.Int()
		if err != nil {
			return nil, nil, nil, err
		}
		args = append(args,
			events[i].(*event).aggregate.TenantId.String(),
			events[i].(*event).aggregate.Type.String(),
			aggregateVersion,
			events[i].(*event).aggregate.Id.String(),
			events[i].(*event).aggregate.Sequence,
			events[i].(*event).eventType.String(),
			events[i].(*event).payload,
			events[i].(*event).aggregate.ResourceOwner,
			events[i].(*event).creator,
			events[i].(*event).correlationId,
			events[i].(*event).causationId,
		)
	}

	return events, placeholders, args, nil
}

var _ Event = (*event)(nil)

type event struct {
	aggregate     *Aggregate
	eventType     EventType
	payload       []byte
	creator       string
	correlationId *string
	causationId   *string
	position      decimal.Decimal
	createdAt     time.Time
}

func (e *event) GetAggregate() *Aggregate {
	return e.aggregate
}

func (e *event) GetCreator() string {
	return e.creator
}

func (e *event) GetEventType() EventType {
	return e.eventType
}

func (e *event) GetCorrelationId() *string {
	return e.correlationId
}

func (e *event) GetCausationId() *string {
	return e.causationId
}

func (e *event) GetPosition() decimal.Decimal {
	return e.position
}

func (e *event) GetCreatedAt() time.Time {
	return e.createdAt
}

func (e *event) UnmarshalData(ptr any) error {
	return json.Unmarshal(e.payload, ptr)
}

func commandToEvent(sequence *latestSequence, command Command) (_ *event, err error) {
	var payload []byte
	if command.GetPayload() != nil {
		payload, err = json.Marshal(command.GetPayload())
		if err != nil {
			return nil, fmt.Errorf("failed to marshal payload: %w", err)
		}
	}
	return &event{
		aggregate:     sequence.aggregate,
		eventType:     command.GetEventType(),
		payload:       payload,
		creator:       command.GetCreator(),
		correlationId: command.GetCorrelationId(),
		causationId:   command.GetCausationId(),
	}, nil
}
