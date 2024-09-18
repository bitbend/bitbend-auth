package eventstore

import "context"

type Reducer interface {
	Query() *SearchQueryBuilder
	AppendEvents(...Event)
	Reduce() error
}

func (es *EventStore) Reducer(ctx context.Context, searchQueryBuilder *SearchQueryBuilder, reducer Reducer) error {
	return nil
}
