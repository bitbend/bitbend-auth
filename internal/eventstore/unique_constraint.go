package eventstore

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"github.com/driftbase/auth/internal/sverror"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stephenafamo/bob"
	"github.com/stephenafamo/bob/dialect/psql"
	"github.com/stephenafamo/bob/dialect/psql/dm"
	"github.com/stephenafamo/bob/dialect/psql/im"
	"log"
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

const uniqueConstraintErrorDetailPlaceholder = "(%s, %s, %s)"

func handleUniqueConstraints(ctx context.Context, tx pgx.Tx, commands []Command) error {
	deleteConditions := make([]bob.Expression, 0)
	deleteQuery := psql.Delete(
		dm.From("unique_constraints"),
	)
	deleteConstraints := map[string]*UniqueConstraint{}

	addQuery := psql.Insert(
		im.Into(
			"unique_constraints",
			"tenant_id",
			"unique_type",
			"unique_value",
		),
	)
	addConstraints := map[string]*UniqueConstraint{}

	for _, command := range commands {
		for _, constraint := range command.GetUniqueConstraints() {
			tenantId := command.GetAggregate().TenantId
			if constraint.IsGlobal {
				tenantId = defaultTenantId
			}
			switch constraint.Action {
			case UniqueConstraintActionAdd:
				addQuery.Apply(im.Values(psql.Arg(tenantId.String(), constraint.UniqueType.String(), constraint.UniqueValue)))
				addConstraints[fmt.Sprintf(uniqueConstraintErrorDetailPlaceholder, tenantId.String(), constraint.UniqueType.String(), constraint.UniqueValue)] = constraint
			case UniqueConstraintActionRemove:
				deleteConditions = append(deleteConditions,
					psql.And(
						psql.Quote("tenant_id").EQ(psql.Arg(tenantId.String())),
						psql.Quote("unique_type").EQ(psql.Arg(constraint.UniqueType.String())),
						psql.Quote("unique_value").EQ(psql.Arg(constraint.UniqueValue)),
					),
				)
				deleteConstraints[fmt.Sprintf(uniqueConstraintErrorDetailPlaceholder, tenantId.String(), constraint.UniqueType.String(), constraint.UniqueValue)] = constraint
			case UniqueConstraintActionTenantRemove:
				deleteConditions = append(deleteConditions,
					psql.And(
						psql.Quote("tenant_id").EQ(psql.Arg(tenantId.String())),
					),
				)
				deleteConstraints[fmt.Sprintf(uniqueConstraintErrorDetailPlaceholder, tenantId.String(), constraint.UniqueType.String(), constraint.UniqueValue)] = constraint
			}
		}
	}

	deleteQuery.Apply(
		dm.Where(
			psql.Or(deleteConditions...),
		),
	)

	deleteStmt, deleteArgs, err := deleteQuery.Build()
	if err != nil {
		log.Fatal(err)
	}

	if len(deleteArgs) > 0 {
		_, err := tx.Exec(ctx, deleteStmt, deleteArgs...)
		if err != nil {
			if constraint := uniqueConstraintFromError(err, deleteConstraints); constraint != nil {
				return sverror.NewAlreadyExistsError(constraint.ErrorMessage, nil)
			} else {
				return sverror.NewInternalError("error.failed.to.delete.unique.constraints", err)
			}
		}
	}

	addStmt, addArgs, err := addQuery.Build()
	if err != nil {
		log.Fatal(err)
	}

	if len(addArgs) > 0 {
		_, err := tx.Exec(ctx, addStmt, addArgs...)
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
