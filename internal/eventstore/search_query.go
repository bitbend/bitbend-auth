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

type SearchColumns string

const (
	SearchColumnsEvent SearchColumns = "event"
)

type SearchQueryBuilder struct {
	columns               SearchColumns
	awaitOpenTransactions bool
	tenantIds             []string
	resourceOwners        []string
	creators              []string
	queries               []*SearchQuery
	createdAfter          time.Time
	createdBefore         time.Time
	allowTimeTravel       bool
	positionGreaterEqual  decimal.Decimal
	sequenceGreaterEqual  uint64
	desc                  bool
	limit                 uint64
	offset                uint64
}

func (sqb *SearchQueryBuilder) Columns(columns SearchColumns) *SearchQueryBuilder {
	sqb.columns = columns
	return sqb
}

func (sqb *SearchQueryBuilder) TenantIds(tenantIds ...string) *SearchQueryBuilder {
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
	queryBuilder := psql.Select(
		sm.From("events"),
	)

	switch sqb.columns {
	case SearchColumnsEvent:
		queryBuilder.Apply(
			sm.Columns(
				"tenant_id",
				"aggregate_type",
				"aggregate_version",
				"aggregate_id",
				"aggregate_sequence",
				"event_type",
				"payload",
				"creators",
				"resource_owner",
				"correlation_id",
				"causation_id",
				"global_position",
				"created_at",
			),
		)
	}

	if sqb.tenantIds != nil {
		if len(sqb.tenantIds) == 1 {
			queryBuilder.Apply(
				sm.Where(
					psql.Quote("tenant_id").EQ(psql.Arg(sqb.tenantIds[0])),
				),
			)
		} else {
			queryBuilder.Apply(
				sm.Where(
					psql.Quote("tenant_id").EQ(psql.Raw("ANY(?)", sqb.tenantIds)),
				),
			)
		}
	}

	conditions := make([]bob.Expression, 0)
	for _, query := range sqb.queries {
		condition := make([]bob.Expression, 0)
		if query.aggregateTypes != nil {
			if len(query.aggregateTypes) == 1 {
				condition = append(condition, psql.Quote("aggregate_type").EQ(psql.Arg(query.aggregateTypes[0])))
			} else {
				condition = append(condition, psql.Quote("aggregate_type").EQ(psql.Raw("ANY(?)", query.aggregateTypes)))
			}
		}

		if query.aggregateIds != nil {
			if len(query.aggregateIds) == 1 {
				condition = append(condition, psql.Quote("aggregate_id").EQ(psql.Arg(query.aggregateIds[0])))
			} else {
				condition = append(condition, psql.Quote("aggregate_id").EQ(psql.Raw("ANY(?)", query.aggregateIds)))
			}
		}

		if query.eventTypes != nil {
			if len(query.eventTypes) == 1 {
				condition = append(condition, psql.Quote("event_type").EQ(psql.Arg(query.eventTypes[0])))
			} else {
				condition = append(condition, psql.Quote("event_type").EQ(psql.Raw("ANY(?)", query.eventTypes)))
			}
		}

		if query.payload != nil {
			condition = append(condition, psql.Quote("payload").EQ(psql.Arg(query.payload)))
		}

		conditions = append(conditions, psql.And(condition...))
	}

	if len(conditions) > 0 {
		queryBuilder.Apply(
			sm.Where(
				psql.Or(conditions...),
			),
		)
	}

	if sqb.resourceOwners != nil {
		if len(sqb.resourceOwners) == 1 {
			queryBuilder.Apply(
				sm.Where(
					psql.Quote("resource_owner").EQ(psql.Arg(sqb.resourceOwners[0])),
				),
			)
		} else {
			queryBuilder.Apply(
				sm.Where(
					psql.Quote("resource_owner").EQ(psql.Raw("ANY(?)", sqb.resourceOwners)),
				),
			)
		}
	}

	if sqb.creators != nil {
		if len(sqb.creators) == 1 {
			queryBuilder.Apply(
				sm.Where(
					psql.Quote("creators").EQ(psql.Arg(sqb.creators)),
				),
			)
		} else {
			queryBuilder.Apply(
				sm.Where(
					psql.Quote("creators").EQ(psql.Raw("ANY(?)", sqb.creators)),
				),
			)
		}
	}

	if !sqb.createdAfter.IsZero() {
		queryBuilder.Apply(
			sm.Where(
				psql.Quote("created_at").GTE(psql.Arg(sqb.createdAfter)),
			),
		)
	}

	if !sqb.createdBefore.IsZero() {
		queryBuilder.Apply(
			sm.Where(
				psql.Quote("created_at").LTE(psql.Arg(sqb.createdBefore)),
			),
		)
	}

	if !sqb.positionGreaterEqual.IsZero() {
		queryBuilder.Apply(
			sm.Where(
				psql.Quote("global_position").GTE(psql.Arg(sqb.positionGreaterEqual)),
			),
		)
	}

	if sqb.sequenceGreaterEqual > 0 {
		queryBuilder.Apply(
			sm.Where(
				psql.Quote("aggregate_sequence").GTE(psql.Arg(sqb.sequenceGreaterEqual)),
			),
		)
	}

	if sqb.desc {
		queryBuilder.Apply(
			sm.OrderBy("aggregate_sequence").Desc(),
		)
	} else {
		queryBuilder.Apply(
			sm.OrderBy("aggregate_sequence").Asc(),
		)
	}

	if sqb.limit > 0 {
		queryBuilder.Apply(
			sm.Limit(sqb.limit),
		)
	}

	if sqb.offset > 0 {
		queryBuilder.Apply(
			sm.Offset(sqb.offset),
		)
	}

	return queryBuilder.Build()
}

func NewSearchQueryBuilder(columns SearchColumns) *SearchQueryBuilder {
	return &SearchQueryBuilder{
		columns: columns,
	}
}

func (sqb *SearchQueryBuilder) AddQuery() *SearchQuery {
	query := &SearchQuery{
		builder: sqb,
	}

	sqb.queries = append(sqb.queries, query)

	return query
}
