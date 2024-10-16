package eventstore

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type AggregateType string

var aggregateVersionRegexp = regexp.MustCompile(`^v[0-9]+(\.[0-9]+){0,2}$`)

type AggregateVersion string

func (av AggregateVersion) Int() (error, int) {
	if !aggregateVersionRegexp.MatchString(string(av)) {
		return fmt.Errorf("error invalid aggregate version format"), 0
	}
	version, err := strconv.Atoi(strings.TrimPrefix(string(av), "v"))
	if err != nil {
		return fmt.Errorf("error parsing aggregate version to int"), 0
	}

	return nil, version
}

type Aggregate struct {
	TenantId  string        `json:"-"`
	Id        string        `json:"-"`
	Owner     string        `json:"-"`
	Type      AggregateType `json:"-"`
	Version   int           `json:"-"`
	Sequence  int64         `json:"-"`
	CreatedAt time.Time     `json:"-"`
}
