package eventstore

import (
	"context"
	_ "embed"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"sort"
	"sync"
	"time"
)

type eventTypeInterceptors struct {
	eventMapper func(Event) (Event, error)
}

var (
	eventInterceptors map[EventType]eventTypeInterceptors
	eventTypes        []string
	aggregateTypes    []string
	eventTypeMapping  = map[EventType]AggregateType{}
)

type EventStore struct {
	PushTimeout     time.Duration
	maxRetries      int
	tenants         []string
	lastTenantQuery time.Time
	tenantMutex     sync.Mutex
	pool            *pgxpool.Pool
}

func (es *EventStore) withTxn(ctx context.Context, txOptions pgx.TxOptions, fn func(tx pgx.Tx, ctx context.Context) error) error {
	tx, err := es.pool.BeginTx(ctx, txOptions)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback(ctx)
			panic(p)
		} else if err != nil {
			if rbErr := tx.Rollback(ctx); rbErr != nil {
				err = fmt.Errorf("failed to rollback transaction: %v, original error: %w", rbErr, err)
			}
		}
	}()

	err = fn(tx, ctx)
	if err != nil {
		return err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (es *EventStore) mapEvents(events []Event) (mappedEvents []Event, err error) {
	mappedEvents = make([]Event, len(events))
	for i, event := range events {
		mappedEvents[i], err = es.mapEventLocked(event)
		if err != nil {
			return nil, err
		}
	}
	return mappedEvents, nil
}

func (es *EventStore) mapEvent(event Event) (Event, error) {
	return es.mapEventLocked(event)
}

func (es *EventStore) mapEventLocked(event Event) (Event, error) {
	interceptors, ok := eventInterceptors[event.GetEventType()]
	if !ok || interceptors.eventMapper == nil {
		return EventBaseFromEvent(event), nil
	}
	return interceptors.eventMapper(event)
}

func RegisterEventMapper(aggregateType AggregateType, eventType EventType, mapper func(Event) (Event, error)) {
	if mapper == nil || eventType == "" {
		return
	}

	appendEventType(eventType)
	appendAggregateType(aggregateType)

	if eventInterceptors == nil {
		eventInterceptors = make(map[EventType]eventTypeInterceptors)
	}

	interceptor := eventInterceptors[eventType]
	interceptor.eventMapper = mapper
	eventInterceptors[eventType] = interceptor
	eventTypeMapping[eventType] = aggregateType
}

func appendEventType(typ EventType) {
	i := sort.SearchStrings(eventTypes, string(typ))
	if i < len(eventTypes) && eventTypes[i] == string(typ) {
		return
	}
	eventTypes = append(eventTypes[:i], append([]string{string(typ)}, eventTypes[i:]...)...)
}

func appendAggregateType(typ AggregateType) {
	i := sort.SearchStrings(aggregateTypes, string(typ))
	if len(aggregateTypes) > i && aggregateTypes[i] == string(typ) {
		return
	}
	aggregateTypes = append(aggregateTypes[:i], append([]string{string(typ)}, aggregateTypes[i:]...)...)
}
