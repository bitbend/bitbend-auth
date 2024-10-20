package eventstore

import (
	"context"
	"github.com/jackc/pgx/v5"
)

type Session struct {
	es          *EventStore
	startStream bool
	commands    []Command
}

func (es *EventStore) NewSession(ctx context.Context) (*Session, error) {
	return &Session{
		es:       es,
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

	mappedEvents, err := s.es.mapEvents(events)
	if err != nil {
		return nil, err
	}

	return mappedEvents, nil
}
