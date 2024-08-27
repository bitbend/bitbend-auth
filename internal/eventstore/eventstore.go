package eventstore

import (
	"sync"
	"time"
)

type EventStore struct {
	PushTimeout     time.Duration
	maxRetries      int
	tenants         []string
	lastTenantQuery time.Time
	tenantMutex     sync.Mutex
}
