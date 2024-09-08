package eventstore

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

const defaultTenantId = TenantId("default")

type TenantId string

func (ti TenantId) String() string {
	return string(ti)
}

type AggregateType string

func (at AggregateType) String() string {
	return string(at)
}

type AggregateVersion string

var aggregateVersionRegexp = regexp.MustCompile(`^v[0-9]+(\.[0-9]+){0,2}$`)

func (av AggregateVersion) String() string {
	return string(av)
}

func (av AggregateVersion) Int() (int, error) {
	aggregateVersion, err := strconv.Atoi(strings.TrimPrefix(string(av), "v"))
	if err != nil {
		return 0, fmt.Errorf("aggregate version invalid: %w", err)
	}
	return aggregateVersion, nil
}

func (av AggregateVersion) Validate() error {
	if !aggregateVersionRegexp.MatchString(string(av)) {
		return fmt.Errorf("error aggregate version invalid %s", av)
	}
	return nil
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
