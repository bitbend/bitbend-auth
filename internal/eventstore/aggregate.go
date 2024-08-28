package eventstore

type TenantId string

func (ti TenantId) String() string {
	return string(ti)
}

type AggregateType string

func (at AggregateType) String() string {
	return string(at)
}

type AggregateVersion uint

func (av AggregateVersion) Uint() uint {
	return uint(av)
}

type AggregateId string

func (ai AggregateId) String() string {
	return string(ai)
}

type Aggregate struct {
	TenantId TenantId
	Type     AggregateType
	Version  AggregateVersion
	Id       AggregateId
	Sequence uint64
	Owner    string
}
