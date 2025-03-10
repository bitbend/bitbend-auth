package eventstore

import (
	"log"
	"sync"
)

var (
	subscriptions = map[AggregateType][]*Subscription{}
	subsMutext    sync.Mutex
)

type Subscription struct {
	Events chan Event
	types  map[AggregateType][]EventType
}

func SubscribeAggregates(eventQueue chan Event, aggregates ...AggregateType) *Subscription {
	types := make(map[AggregateType][]EventType, len(aggregates))
	for _, aggregate := range aggregates {
		types[aggregate] = nil
	}
	sub := &Subscription{
		Events: eventQueue,
		types:  types,
	}

	subsMutext.Lock()
	defer subsMutext.Unlock()

	for _, aggregate := range aggregates {
		subscriptions[aggregate] = append(subscriptions[aggregate], sub)
	}

	return sub
}

func SubscribeEventTypes(eventQueue chan Event, types map[AggregateType][]EventType) *Subscription {
	sub := &Subscription{
		Events: eventQueue,
		types:  types,
	}

	subsMutext.Lock()
	defer subsMutext.Unlock()

	for aggregate := range types {
		subscriptions[aggregate] = append(subscriptions[aggregate], sub)
	}

	return sub
}

func (es *EventStore) notify(events []Event) {
	subsMutext.Lock()
	defer subsMutext.Unlock()
	for _, event := range events {
		subs, ok := subscriptions[event.GetAggregate().Type]
		if !ok {
			continue
		}
		for _, sub := range subs {
			eventTypes := sub.types[event.GetAggregate().Type]
			if len(eventTypes) == 0 {
				sub.Events <- event
				continue
			}
			for _, eventType := range eventTypes {
				if event.GetType() == eventType {
					select {
					case sub.Events <- event:
					default:
						log.Println("unable to push event")
					}
					break
				}
			}
		}
	}
}

func (s *Subscription) Unsubscribe() {
	subsMutext.Lock()
	defer subsMutext.Unlock()
	for aggregate := range s.types {
		subs, ok := subscriptions[aggregate]
		if !ok {
			continue
		}
		for i := len(subs) - 1; i >= 0; i-- {
			if subs[i] == s {
				subs[i] = subs[len(subs)-1]
				subs[len(subs)-1] = nil
				subs = subs[:len(subs)-1]
			}
		}
	}
	_, ok := <-s.Events
	if ok {
		close(s.Events)
	}
}
