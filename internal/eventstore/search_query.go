package eventstore

import (
	"github.com/shopspring/decimal"
	"github.com/stephenafamo/bob"
	"github.com/stephenafamo/bob/dialect/psql"
	"github.com/stephenafamo/bob/dialect/psql/sm"
	"time"
)

type SearchQuery struct {
	builder        *SearchQueryBuilder
	aggregateTypes []AggregateType
	aggregateIds   []AggregateId
	eventTypes     []EventType
	payload        map[string]any
}

func (sq *SearchQuery) Or() *SearchQuery {
	return sq.builder.AddQuery()
}

func (sq *SearchQuery) AggregateTypes(aggregateTypes ...AggregateType) *SearchQuery {
	sq.aggregateTypes = aggregateTypes
	return sq
}

func (sq *SearchQuery) AggregateIds(aggregateIds ...AggregateId) *SearchQuery {
	sq.aggregateIds = aggregateIds
	return sq
}

func (sq *SearchQuery) EventTypes(eventTypes ...EventType) *SearchQuery {
	sq.eventTypes = eventTypes
	return sq
}

func (sq *SearchQuery) Payload(payload map[string]any) *SearchQuery {
	sq.payload = payload
	return sq
}

func (sq *SearchQuery) Build() *SearchQueryBuilder {
	return sq.builder
}

type SearchQueryBuilder struct {
	tenantIds            []TenantId
	resourceOwners       []string
	creators             []string
	queries              []*SearchQuery
	createdAfter         time.Time
	createdBefore        time.Time
	allowTimeTravel      bool
	positionGreaterEqual decimal.Decimal
	sequenceGreaterEqual uint64
	desc                 bool
	limit                uint64
	offset               uint64
}

func (sqb *SearchQueryBuilder) TenantIds(tenantIds ...TenantId) *SearchQueryBuilder {
	sqb.tenantIds = tenantIds
	return sqb
}

func (sqb *SearchQueryBuilder) ResourceOwners(resourceOwners ...string) *SearchQueryBuilder {
	sqb.resourceOwners = resourceOwners
	return sqb
}

func (sqb *SearchQueryBuilder) Creators(creators ...string) *SearchQueryBuilder {
	sqb.creators = creators
	return sqb
}

func (sqb *SearchQueryBuilder) CreatedAfter(createdAfter time.Time) *SearchQueryBuilder {
	sqb.createdAfter = createdAfter
	return sqb
}

func (sqb *SearchQueryBuilder) CreatedBefore(createdBefore time.Time) *SearchQueryBuilder {
	sqb.createdBefore = createdBefore
	return sqb
}

func (sqb *SearchQueryBuilder) AllowTimeTravel() *SearchQueryBuilder {
	sqb.allowTimeTravel = true
	return sqb
}

func (sqb *SearchQueryBuilder) PositionGreaterEqual(position decimal.Decimal) *SearchQueryBuilder {
	sqb.positionGreaterEqual = position
	return sqb
}

func (sqb *SearchQueryBuilder) SequenceGreaterEqual(sequence uint64) *SearchQueryBuilder {
	sqb.sequenceGreaterEqual = sequence
	return sqb
}

func (sqb *SearchQueryBuilder) Asc() *SearchQueryBuilder {
	sqb.desc = false
	return sqb
}

func (sqb *SearchQueryBuilder) Desc() *SearchQueryBuilder {
	sqb.desc = true
	return sqb
}

func (sqb *SearchQueryBuilder) Limit(limit uint64) *SearchQueryBuilder {
	sqb.limit = limit
	return sqb
}

func (sqb *SearchQueryBuilder) Offset(offset uint64) *SearchQueryBuilder {
	sqb.offset = offset
	return sqb
}

func (sqb *SearchQueryBuilder) ToSql() (string, []any, error) {
	return eventSearchQuery(sqb)
}

func NewSearchQueryBuilder() *SearchQueryBuilder {
	return &SearchQueryBuilder{}
}

func (sqb *SearchQueryBuilder) AddQuery() *SearchQuery {
	query := &SearchQuery{
		builder: sqb,
	}

	sqb.queries = append(sqb.queries, query)

	return query
}

func eventSearchQuery(builder *SearchQueryBuilder) (string, []any, error) {
	query := psql.Select(
		sm.From("events"),
		sm.Columns(
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
			"global_position",
			"created_at",
		),
	)

	if builder.tenantIds != nil {
		if len(builder.tenantIds) == 1 {
			query.Apply(
				sm.Where(
					psql.Quote("tenant_id").EQ(psql.Arg(builder.tenantIds[0])),
				),
			)
		} else {
			query.Apply(
				sm.Where(
					psql.Quote("tenant_id").EQ(psql.Raw("ANY(?)", builder.tenantIds)),
				),
			)
		}
	}

	if len(builder.queries) > 0 {
		expressions := buildSearchQueryExpressions(builder.queries)
		query.Apply(
			sm.Where(
				psql.Or(
					expressions...,
				),
			),
		)
	}

	if builder.resourceOwners != nil {
		if len(builder.resourceOwners) == 1 {
			query.Apply(
				sm.Where(
					psql.Quote("stream_owner").EQ(psql.Arg(builder.resourceOwners[0])),
				),
			)
		} else {
			query.Apply(
				sm.Where(
					psql.Quote("stream_owner").EQ(psql.Raw("ANY(?)", builder.resourceOwners)),
				),
			)
		}
	}

	if builder.creators != nil {
		if len(builder.creators) == 1 {
			query.Apply(
				sm.Where(
					psql.Quote("creator").EQ(psql.Arg(builder.creators[0])),
				),
			)
		} else {
			query.Apply(
				sm.Where(
					psql.Quote("creator").EQ(psql.Raw("ANY(?)", builder.creators)),
				),
			)
		}
	}

	if !builder.createdAfter.IsZero() {
		query.Apply(
			sm.Where(
				psql.Quote("created_at").GTE(psql.Arg(builder.createdAfter)),
			),
		)
	}

	if !builder.createdBefore.IsZero() {
		query.Apply(
			sm.Where(
				psql.Quote("created_at").LTE(psql.Arg(builder.createdBefore)),
			),
		)
	}

	if !builder.positionGreaterEqual.IsZero() {
		query.Apply(
			sm.Where(
				psql.Quote("global_sequence").GTE(psql.Arg(builder.positionGreaterEqual)),
			),
		)
	}

	if builder.sequenceGreaterEqual > 0 {
		query.Apply(
			sm.Where(
				psql.Quote("stream_sequence").GTE(psql.Arg(builder.sequenceGreaterEqual)),
			),
		)
	}

	if builder.desc {
		query.Apply(
			sm.OrderBy("stream_sequence").Desc(),
		)
	} else {
		query.Apply(
			sm.OrderBy("stream_sequence").Asc(),
		)
	}

	if builder.limit > 0 {
		query.Apply(
			sm.Limit(builder.limit),
		)
	}

	if builder.offset > 0 {
		query.Apply(
			sm.Offset(builder.offset),
		)
	}

	return query.Build()
}

func buildSearchQueryExpressions(searchQueries []*SearchQuery) []bob.Expression {
	expressions := make([]bob.Expression, 0)
	for _, searchQuery := range searchQueries {
		expression := make([]bob.Expression, 0)
		if searchQuery.aggregateTypes != nil {
			if len(searchQuery.aggregateTypes) == 1 {
				expression = append(expression, psql.Quote("stream_type").EQ(psql.Arg(searchQuery.aggregateTypes[0])))
			} else {
				expression = append(expression, psql.Quote("stream_type").EQ(psql.Raw("ANY(?)", searchQuery.aggregateTypes)))
			}
		}

		if searchQuery.aggregateIds != nil {
			if len(searchQuery.aggregateIds) == 1 {
				expression = append(expression, psql.Quote("stream_id").EQ(psql.Arg(searchQuery.aggregateIds[0])))
			} else {
				expression = append(expression, psql.Quote("stream_id").EQ(psql.Raw("ANY(?)", searchQuery.aggregateIds)))
			}
		}

		if searchQuery.eventTypes != nil {
			if len(searchQuery.eventTypes) == 1 {
				expression = append(expression, psql.Quote("event_type").EQ(psql.Arg(searchQuery.eventTypes[0])))
			} else {
				expression = append(expression, psql.Quote("event_type").EQ(psql.Raw("ANY(?)", searchQuery.eventTypes)))
			}
		}

		if searchQuery.payload != nil {
			expression = append(expression, psql.Quote("payload").EQ(psql.Arg(searchQuery.payload)))
		}

		expressions = append(expressions, psql.And(expression...))
	}

	return expressions
}
