package eventstore

const defaultTenantId = TenantId("default")

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
	TenantId TenantId         `json:"-"`
	Type     AggregateType    `json:"-"`
	Version  AggregateVersion `json:"-"`
	Id       AggregateId      `json:"-"`
	Sequence uint64           `json:"-"`
	Owner    string           `json:"-"`
}

func NewAggregate(
	tenantId TenantId,
	aggregateType AggregateType,
	aggregateVersion AggregateVersion,
	aggregateId AggregateId,
	sequence uint64,
	owner string,
) *Aggregate {
	return &Aggregate{
		TenantId: tenantId,
		Type:     aggregateType,
		Version:  aggregateVersion,
		Id:       aggregateId,
		Sequence: sequence,
		Owner:    owner,
	}
}
