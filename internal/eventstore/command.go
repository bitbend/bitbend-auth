package eventstore

type Command interface {
	GetData() any
	GetUniqueConstraints() []*UniqueConstraint
}
