package eventstore

import (
	"context"
	_ "embed"
	"fmt"
	"github.com/jackc/pgx/v5"
	"strings"
)

var (
	//go:embed sequence.sql
	sequenceStmt string
)

type latestSequence struct {
	aggregate *Aggregate
}

func latestSequences(ctx context.Context, tx pgx.Tx, commands []Command) ([]*latestSequence, error) {
	sequences := commandsToSequences(ctx, commands)

	placeholders, args := sequencesToSql(sequences)
	rows, err := tx.Query(ctx, fmt.Sprintf(sequenceStmt, strings.Join(placeholders, " union all ")), args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query latest sequences: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		if err := scanToSequence(rows, sequences); err != nil {
			return nil, fmt.Errorf("failed to scan latest sequences: %w", err)
		}
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("failed to scan latest sequences: %w", rows.Err())
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

func sequencesToSql(sequences []*latestSequence) (placeholders []string, args []any) {
	placeholders = make([]string, 0)
	args = make([]any, 0)

	for _, sequence := range sequences {
		placeholders = append(placeholders, fmt.Sprintf(`(select tenant_id, aggregate_type, aggregate_id, aggregate_sequence from events where tenant_id = $%d and aggregate_type = $%d and aggregate_id = $%d order by aggregate_sequence desc limit 1)`,
			len(args)+1,
			len(args)+2,
			len(args)+3,
		))
		args = append(args, sequence.aggregate.TenantId.String(), sequence.aggregate.Type.String(), sequence.aggregate.Id.String())
	}

	return placeholders, args
}

func scanToSequence(rows pgx.Rows, sequences []*latestSequence) error {
	var tenantId TenantId
	var aggregateType AggregateType
	var aggregateId AggregateId
	var currentSequence uint64
	var owner string

	if err := rows.Scan(&tenantId, &owner, &aggregateType, &aggregateId, &currentSequence); err != nil {
		return fmt.Errorf("failed to scan row: %w", err)
	}

	sequence := searchSequence(sequences, aggregateType, aggregateId, tenantId)
	if sequence == nil {
		return nil
	}

	sequence.aggregate.Sequence = currentSequence

	if sequence.aggregate.ResourceOwner == "" {
		sequence.aggregate.ResourceOwner = owner
	}

	return nil
}
