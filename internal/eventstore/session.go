package eventstore

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
)

type Session struct {
	es          *EventStore
	tx          pgx.Tx
	startStream bool
	commands    []Command
}

func (es *EventStore) NewSession(ctx context.Context) (*Session, error) {
	return &Session{
		es:       es,
		commands: make([]Command, 0),
	}, nil
}

func (es *EventStore) NewSessionTx(ctx context.Context) (*Session, error) {
	tx, err := es.db.Pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		return nil, fmt.Errorf("error starting session transaction: %v", err)
	}

	return &Session{
		es:       es,
		tx:       tx,
		commands: make([]Command, 0),
	}, nil
}

func (s *Session) StartStream(commands ...Command) *Session {
	s.startStream = true
	s.commands = append(s.commands, commands...)
	return s
}

func (s *Session) AppendEvents(commands ...Command) *Session {
	s.commands = append(s.commands, commands...)
	return s
}

func (s *Session) SaveChanges(ctx context.Context) ([]Event, error) {
	var events []Event

	if s.tx != nil {
		sequences, err := fetchLatestSequences(ctx, s.tx, s.commands)
		if err != nil {
			return nil, err
		}

		events, err = commandsToEvents(s.commands, sequences)
		if err != nil {
			return nil, err
		}

		if s.startStream {
			if err = createStreams(ctx, s.tx, events); err != nil {
				return nil, err
			}
		}

		events, err = createEvents(ctx, s.tx, events)
		if err != nil {
			return nil, err
		}

		defer func() {
			if p := recover(); p != nil {
				_ = s.tx.Rollback(ctx)
				panic(p)
			} else if err != nil {
				_ = s.tx.Rollback(ctx)
			}
		}()

		if err = s.tx.Commit(ctx); err != nil {
			return nil, fmt.Errorf("failed to commit session transaction: %w", err)
		}
	} else {
		err := s.es.db.WithTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted}, func(tx pgx.Tx) error {
			sequences, err := fetchLatestSequences(ctx, tx, s.commands)
			if err != nil {
				return err
			}

			events, err = commandsToEvents(s.commands, sequences)
			if err != nil {
				return err
			}

			if s.startStream {
				if err = createStreams(ctx, tx, events); err != nil {
					return err
				}
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
	}

	mappedEvents, err := s.es.mapEvents(events)
	if err != nil {
		return nil, err
	}

	return mappedEvents, nil
}

func (s *Session) Reduce(ctx context.Context, reducer Reducer) error {
	stmt, args, err := reducer.Query().ToSQL()
	if err != nil {
		return err
	}

	var rows pgx.Rows
	if s.tx != nil {
		rows, err = s.tx.Query(ctx, stmt, args...)
		if err != nil {
			return err
		}
	} else {
		rows, err = s.es.db.Pool.Query(ctx, stmt, args...)
		if err != nil {
			return err
		}
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

	return reducer.Reduce()
}
