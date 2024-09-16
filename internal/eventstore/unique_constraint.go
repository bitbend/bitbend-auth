package eventstore

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"github.com/driftbase/auth/internal/sverror"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"strings"
)

type UniqueType string

func (ut UniqueType) String() string {
	return string(ut)
}

type UniqueConstraintAction string

const (
	UniqueConstraintActionAdd          UniqueConstraintAction = "add"
	UniqueConstraintActionRemove       UniqueConstraintAction = "remove"
	UniqueConstraintActionTenantRemove UniqueConstraintAction = "tenant_remove"
)

type UniqueConstraint struct {
	UniqueType   UniqueType
	UniqueValue  string
	Action       UniqueConstraintAction
	ErrorMessage string
	IsGlobal     bool
}

func NewUniqueConstraint(
	uniqueType UniqueType,
	uniqueValue string,
	uniqueConstraintAction UniqueConstraintAction,
	errorMessage string,
	isGlobal bool,
) *UniqueConstraint {
	return &UniqueConstraint{
		UniqueType:   uniqueType,
		UniqueValue:  uniqueValue,
		Action:       uniqueConstraintAction,
		ErrorMessage: errorMessage,
		IsGlobal:     isGlobal,
	}
}

var (
	//go:embed unique_constraint_add.sql
	uniqueConstraintsAddStmt string
	//go:embed unique_constraint_delete.sql
	uniqueConstraintsDeleteStmt string
)

const uniqueConstraintErrorDetailPlaceholder = "(%s, %s, %s)"

func handleUniqueConstraints(ctx context.Context, tx pgx.Tx, commands []Command) error {
	deletePlaceholders := make([]string, 0)
	deleteArgs := make([]any, 0)
	deleteConstraints := map[string]*UniqueConstraint{}

	addPlaceholders := make([]string, 0)
	addArgs := make([]any, 0)
	addConstraints := map[string]*UniqueConstraint{}

	for _, command := range commands {
		for _, constraint := range command.GetUniqueConstraints() {
			tenantId := command.GetAggregate().TenantId
			if constraint.IsGlobal {
				tenantId = defaultTenantId
			}
			switch constraint.Action {
			case UniqueConstraintActionAdd:
				addPlaceholders = append(addPlaceholders, fmt.Sprintf("($%d, $%d, $%d)", len(addArgs)+1, len(addArgs)+2, len(addArgs)+3))
				addArgs = append(addArgs, tenantId.String(), constraint.UniqueType.String(), constraint.UniqueValue)
				addConstraints[fmt.Sprintf(uniqueConstraintErrorDetailPlaceholder, tenantId.String(), constraint.UniqueType.String(), constraint.UniqueValue)] = constraint
			case UniqueConstraintActionRemove:
				deletePlaceholders = append(deletePlaceholders, fmt.Sprintf("(tenant_id = $%d and unique_type = $%d and unique_value = $%d)", len(deleteArgs)+1, len(deleteArgs)+2, len(deleteArgs)+3))
				deleteArgs = append(deleteArgs, tenantId.String(), constraint.UniqueType.String(), constraint.UniqueValue)
				deleteConstraints[fmt.Sprintf(uniqueConstraintErrorDetailPlaceholder, tenantId.String(), constraint.UniqueType.String(), constraint.UniqueValue)] = constraint
			case UniqueConstraintActionTenantRemove:
				deletePlaceholders = append(deletePlaceholders, fmt.Sprintf("(tenant_id = $%d)", len(deleteArgs)+1))
				deleteArgs = append(deleteArgs, tenantId.String())
				deleteConstraints[fmt.Sprintf(uniqueConstraintErrorDetailPlaceholder, tenantId.String(), constraint.UniqueType.String(), constraint.UniqueValue)] = constraint
			}
		}
	}

	if len(deletePlaceholders) > 0 {
		_, err := tx.Exec(ctx, fmt.Sprintf(uniqueConstraintsDeleteStmt, strings.Join(deletePlaceholders, " or ")), deleteArgs...)
		if err != nil {
			if constraint := uniqueConstraintFromError(err, deleteConstraints); constraint != nil {
				return sverror.NewAlreadyExistsError(constraint.ErrorMessage, nil)
			} else {
				return sverror.NewInternalError("error.failed.to.delete.unique.constraints", err)
			}
		}
	}

	if len(addPlaceholders) > 0 {
		_, err := tx.Exec(ctx, fmt.Sprintf(uniqueConstraintsAddStmt, strings.Join(addPlaceholders, ", ")), addArgs...)
		if err != nil {
			if constraint := uniqueConstraintFromError(err, addConstraints); constraint != nil {
				return sverror.NewAlreadyExistsError(constraint.ErrorMessage, nil)
			} else {
				return sverror.NewInternalError("error.failed.to.add.unique.constraints", err)
			}
		}
	}

	return nil
}

func uniqueConstraintFromError(err error, constraints map[string]*UniqueConstraint) *UniqueConstraint {
	var pgError *pgconn.PgError
	if !errors.As(err, &pgError) {
		return nil
	}
	for key, constraint := range constraints {
		if strings.Contains(pgError.Detail, key) {
			return constraint
		}
	}
	return nil
}
