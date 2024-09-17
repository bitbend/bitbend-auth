package eventstore

import (
	"context"
	_ "embed"
	"github.com/driftbase/auth/internal/sverror"
	"github.com/jackc/pgx/v5"
	"github.com/stephenafamo/bob/dialect/psql"
	"github.com/stephenafamo/bob/dialect/psql/sm"
	"log"
	"strings"
)

type latestSequence struct {
	aggregate *Aggregate
}

func latestSequences(ctx context.Context, tx pgx.Tx, commands []Command) ([]*latestSequence, error) {
	sequences := commandsToSequences(ctx, commands)

	stmt, args := sequencesToSql(sequences)
	rows, err := tx.Query(ctx, stmt, args...)
	if err != nil {
		return nil, sverror.NewInternalError("error.failed.to.query.latest.sequences", err)
	}
	defer rows.Close()

	for rows.Next() {
		if err := scanToSequence(rows, sequences); err != nil {
			return nil, sverror.NewInternalError("error.failed.to.scan.latest.sequences", err)
		}
	}

	if rows.Err() != nil {
		return nil, sverror.NewInternalError("error.failed.to.scan.latest.sequences", rows.Err())
	}

	return sequences, nil
}

func searchSequenceByCommand(sequences []*latestSequence, command Command) *latestSequence {
	for _, sequence := range sequences {
		if sequence.aggregate.Type == command.GetAggregate().Type &&
			sequence.aggregate.Id == command.GetAggregate().Id &&
			sequence.aggregate.TenantId == command.GetAggregate().TenantId {
			return sequence
		}
	}
	return nil
}

func searchSequence(sequences []*latestSequence, aggregateType AggregateType, aggregateId AggregateId, tenantId TenantId) *latestSequence {
	for _, sequence := range sequences {
		if sequence.aggregate.Type == aggregateType &&
			sequence.aggregate.Id == aggregateId &&
			sequence.aggregate.TenantId == tenantId {
			return sequence
		}
	}
	return nil
}

func commandsToSequences(ctx context.Context, commands []Command) []*latestSequence {
	sequences := make([]*latestSequence, 0, len(commands))

	for _, command := range commands {
		if searchSequenceByCommand(sequences, command) != nil {
			continue
		}

		if command.GetAggregate().TenantId == "" {
			command.GetAggregate().TenantId = ""
		}
		sequences = append(sequences, &latestSequence{
			aggregate: command.GetAggregate(),
		})
	}

	return sequences
}

func sequencesToSql(sequences []*latestSequence) (string, []any) {
	stmts := make([]string, 0)
	args := make([]any, 0)

	for _, sequence := range sequences {
		stmts = append(stmts, `(SELECT tenant_id, aggregate_type, aggregate_id, aggregate_sequence FROM events WHERE tenant_id = ? AND aggregate_type = ? AND aggregate_id = ? ORDER BY aggregate_sequence DESC LIMIT 1)`)
		args = append(args, sequence.aggregate.TenantId.String(), sequence.aggregate.Type.String(), sequence.aggregate.Id.String())
	}

	subQuery := psql.RawQuery(strings.Join(stmts, " UNION ALL "), args...)

	query := psql.Select(
		sm.With("existing").As(subQuery),
		sm.From("events"),
		sm.Columns(
			psql.Quote("events", "tenant_id"),
			psql.Quote("events", "resource_owner"),
			psql.Quote("events", "aggregate_type"),
			psql.Quote("events", "aggregate_id"),
			psql.Quote("events", "aggregate_sequence"),
		),
		sm.InnerJoin("existing").On(
			psql.And(
				psql.Quote("events", "tenant_id").EQ(psql.Quote("existing", "tenant_id")),
				psql.Quote("events", "aggregate_type").EQ(psql.Quote("existing", "aggregate_type")),
				psql.Quote("events", "aggregate_id").EQ(psql.Quote("existing", "aggregate_id")),
				psql.Quote("events", "aggregate_sequence").EQ(psql.Quote("existing", "aggregate_sequence")),
			),
		),
		sm.ForUpdate(),
	)

	stmt, args, err := query.Build()
	if err != nil {
		log.Fatal(err)
	}

	return stmt, args
}

func scanToSequence(rows pgx.Rows, sequences []*latestSequence) error {
	var tenantId TenantId
	var aggregateType AggregateType
	var aggregateId AggregateId
	var aggregateSequence uint64
	var resourceOwner string

	if err := rows.Scan(&tenantId, &resourceOwner, &aggregateType, &aggregateId, &aggregateSequence); err != nil {
		return sverror.NewInternalError("error.failed.to.scan.rows", err)
	}

	sequence := searchSequence(sequences, aggregateType, aggregateId, tenantId)
	if sequence == nil {
		return nil
	}

	sequence.aggregate.Sequence = aggregateSequence

	if sequence.aggregate.ResourceOwner == "" {
		sequence.aggregate.ResourceOwner = resourceOwner
	}

	return nil
}
