package eventstore

import (
	"github.com/driftbase/auth/internal/sverror"
	"sort"
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

func RegisterEventMapper(aggregateType AggregateType, eventType EventType, eventMapper func(Event) (Event, error)) {
	if eventMapper == nil || eventType == "" {
		return
	}

	appendEventType(eventType)
	appendAggregateType(aggregateType)

	if eventInterceptors == nil {
		eventInterceptors = make(map[EventType]eventTypeInterceptors)
	}

	interceptor := eventInterceptors[eventType]
	interceptor.eventMapper = eventMapper
	eventInterceptors[eventType] = interceptor
	eventTypeMapping[eventType] = aggregateType
}

func appendEventType(eventType EventType) {
	i := sort.SearchStrings(eventTypes, string(eventType))
	if i < len(eventTypes) && eventTypes[i] == string(eventType) {
		return
	}
	eventTypes = append(eventTypes[:i], append([]string{string(eventType)}, eventTypes[i:]...)...)
}

func appendAggregateType(aggregateType AggregateType) {
	i := sort.SearchStrings(aggregateTypes, string(aggregateType))
	if len(aggregateTypes) > i && aggregateTypes[i] == string(aggregateType) {
		return
	}
	aggregateTypes = append(aggregateTypes[:i], append([]string{string(aggregateType)}, aggregateTypes[i:]...)...)
}

type EventBaseSetter[T any] interface {
	Event
	SetBaseEvent(*EventBase)
	*T
}

func GenericEventMapper[T any, PT EventBaseSetter[T]](event Event) (Event, error) {
	e := PT(new(T))
	e.SetBaseEvent(EventBaseFromEvent(event))

	err := event.UnmarshalPayload(e)
	if err != nil {
		return nil, sverror.NewInternalError("error.unable.to.unmarshal.event", err)
	}

	return e, nil
}
