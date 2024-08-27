package eventstore

type UniqueType string

func (ut UniqueType) String() string {
	return string(ut)
}

type UniqueConstraintAction string

const (
	UniqueConstraintActionAdd           UniqueConstraintAction = "add"
	UniqueConstraintActionRemove        UniqueConstraintAction = "remove"
	UniqueConstraintActionUpdate        UniqueConstraintAction = "update"
	UniqueConstraintActionTenantRemoved UniqueConstraintAction = "tenant_removed"
)

type UniqueConstraint struct {
	UniqueType   UniqueType
	UniqueValue  string
	Action       UniqueConstraintAction
	ErrorMessage string
	IsGlobal     bool
}
