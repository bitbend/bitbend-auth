package eventstore

var (
	eventInterceptors map[AggregateType]map[EventType]eventInterceptor
)

type eventInterceptor struct {
	eventMapper func(Event) (Event, error)
}

func RegisterEventMapper(aggregateType AggregateType, eventType EventType, eventMapper func(Event) (Event, error)) {
	if aggregateType == "" || eventType == "" || eventMapper == nil {
		return
	}

	if eventInterceptors == nil {
		eventInterceptors = make(map[AggregateType]map[EventType]eventInterceptor)
	}

	interceptors := eventInterceptors[aggregateType]
	interceptor := interceptors[eventType]
	interceptor.eventMapper = eventMapper
}

func (es *EventStore) mapEvent(event Event) (Event, error) {
	if interceptors, ok := eventInterceptors[event.GetAggregate().Type]; ok {
		if interceptor, ok := interceptors[event.GetType()]; ok && interceptor.eventMapper != nil {
			return interceptor.eventMapper(event)
		}
	}

	return EventBaseFromEvent(event), nil
}

func (es *EventStore) mapEvents(events []Event) ([]Event, error) {
	mappedEvents := make([]Event, 0, len(events))

	for _, event := range events {
		mappedEvent, err := es.mapEvent(event)
		if err != nil {
			return nil, err
		}
		mappedEvents = append(mappedEvents, mappedEvent)
	}

	return mappedEvents, nil
}
