package eventstore

import (
	"github.com/shopspring/decimal"
	"github.com/stephenafamo/bob"
	"github.com/stephenafamo/bob/dialect/psql"
	"github.com/stephenafamo/bob/dialect/psql/sm"
	"time"
)

type Query struct {
	aggregateTypes []AggregateType
	aggregateIds   []string
	eventTypes     []EventType
	eventPayload   map[string]any
	builder        *QueryBuilder
}

type QueryBuilder struct {
	tenantIds            []string
	queries              []*Query
	globalPositionAfter  decimal.Decimal
	globalPositionBefore decimal.Decimal
	createdAtAfter       time.Time
	createdAtBefore      time.Time
	desc                 bool
}

func (qb *QueryBuilder) ToSQL() (string, []any, error) {
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

	if len(qb.tenantIds) > 0 {
		query.Apply(
			sm.Where(
				psql.Quote("tenant_id").EQ(psql.Raw("ANY(?)", qb.tenantIds)),
			),
		)
	}

	if len(qb.queries) > 0 {
		var exprs []bob.Expression
		for _, q := range qb.queries {
			var expr []bob.Expression
			if len(q.aggregateIds) > 0 {
				expr = append(expr,
					psql.Quote("stream_id").EQ(psql.Raw("ANY(?)", q.aggregateIds)),
				)
			}
			if len(q.aggregateTypes) > 0 {
				expr = append(expr,
					psql.Quote("stream_type").EQ(psql.Raw("ANY(?)", q.aggregateTypes)),
				)
			}
			if len(q.eventTypes) > 0 {
				expr = append(expr,
					psql.Quote("event_type").EQ(psql.Raw("ANY(?)", q.eventTypes)),
				)
			}
			if q.eventPayload != nil {
				expr = append(expr,
					psql.Quote("payload").EQ(psql.Arg(q.eventPayload)),
				)
			}
			exprs = append(exprs, psql.And(expr...))
		}
		if len(exprs) > 0 {
			query.Apply(
				sm.Where(
					psql.Or(exprs...),
				),
			)
		}
	}

	if !qb.globalPositionAfter.IsZero() {
		query.Apply(
			sm.Where(
				psql.Quote("global_position").GTE(psql.Arg(qb.globalPositionAfter)),
			),
		)
	}

	if !qb.globalPositionBefore.IsZero() {
		query.Apply(
			sm.Where(
				psql.Quote("global_position").LTE(psql.Arg(qb.globalPositionAfter)),
			),
		)
	}

	if !qb.createdAtAfter.IsZero() {
		query.Apply(
			sm.Where(
				psql.Quote("created_at").GTE(psql.Arg(qb.createdAtAfter)),
			),
		)
	}

	if !qb.createdAtBefore.IsZero() {
		query.Apply(
			sm.Where(
				psql.Quote("created_at").LTE(psql.Arg(qb.createdAtBefore)),
			),
		)
	}

	if !qb.desc {
		query.Apply(
			sm.OrderBy("global_position").Asc(),
		)
	} else {
		query.Apply(
			sm.OrderBy("global_position").Desc(),
		)
	}

	return query.Build()
}

func NewQuery() *QueryBuilder {
	return &QueryBuilder{}
}
