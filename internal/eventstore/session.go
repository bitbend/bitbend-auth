package eventstore

import (
	"context"
	"github.com/jackc/pgx/v5"
)

type Session struct {
	es       *EventStore
	commands []Command
}

func (es *EventStore) NewSession(ctx context.Context) (*Session, error) {
	return &Session{
		es:       es,
		commands: make([]Command, 0),
	}, nil
}

func (s *Session) Push(commands ...Command) *Session {
	s.commands = append(s.commands, commands...)
	return s
}

func (s *Session) SaveChanges(ctx context.Context) ([]Event, error) {
	var events []Event
	err := s.es.db.WithTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted}, func(tx pgx.Tx) error {
		sequences, err := fetchLatestSequences(ctx, tx, s.commands)
		if err != nil {
			return err
		}

		events, err = commandsToEvents(s.commands, sequences)
		if err != nil {
			return err
		}

		if err = createStreams(ctx, tx, events); err != nil {
			return err
		}

		events, err = createEvents(ctx, tx, events)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	mappedEvents, err := s.es.mapEvents(events)
	if err != nil {
		return nil, err
	}

	s.es.notify(mappedEvents)

	return mappedEvents, nil
}

func (s *Session) Reduce(ctx context.Context, reducer Reducer) error {
	stmt, args, err := reducer.Query().ToSQL()
	if err != nil {
		return err
	}

	rows, err := s.es.db.Pool().Query(ctx, stmt, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		event := new(EventBase)
		if err = rows.Scan(
			&event.Aggregate.TenantId,
			&event.Id,
			&event.Aggregate.Id,
			&event.Aggregate.Type,
			&event.Aggregate.Version,
			&event.Aggregate.Sequence,
			&event.Aggregate.Owner,
			&event.Payload,
			&event.Creator,
			&event.CorrelationId,
			&event.CausationId,
			&event.CreatedAt,
		); err != nil {
			return err
		}

		mappedEvent, err := s.es.mapEvent(event)
		if err != nil {
			return err
		}

		reducer.AppendEvents(mappedEvent)
	}
	if err = rows.Err(); err != nil {
		return err
	}

	return reducer.Reduce()
}
