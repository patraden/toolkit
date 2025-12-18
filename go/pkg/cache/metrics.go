// Copyright 2025 The Toolkit Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cache

import (
	"encoding/json"
	"fmt"
	"sync/atomic"
	"time"
)

// Metrics tracks cache operations and evictions.
//
// All fields are updated atomically and are safe to read concurrently.
type Metrics struct {
	Hits                  uint32 `json:"hits"`
	Misses                uint32 `json:"misses"`
	Sets                  uint32 `json:"sets"`
	Deletes               uint32 `json:"deletes"`
	LazyEvictions         uint32 `json:"lazy_evictions"`           // Expired items found during Get
	ScheduledEvictions    uint32 `json:"scheduled_evictions"`      // Expired items removed by cleaner
	CleanupRuns           uint32 `json:"cleanup_runs"`             // Number of scheduled cleanup runs
	CleanupTimeouts       uint32 `json:"cleanup_timeouts"`         // Number of times cleanup exceeded interval
	LastCleanupDurationMs uint64 `json:"last_cleanup_duration_ms"` // Duration of last cleanup in milliseconds
	LastCleanupItems      uint32 `json:"last_cleanup_items"`       // Items cleaned in last run
}

// Snapshot returns an atomic-load copy of all metrics fields.
func (m *Metrics) Snapshot() Metrics {
	return Metrics{
		Hits:                  atomic.LoadUint32(&m.Hits),
		Misses:                atomic.LoadUint32(&m.Misses),
		Sets:                  atomic.LoadUint32(&m.Sets),
		Deletes:               atomic.LoadUint32(&m.Deletes),
		LazyEvictions:         atomic.LoadUint32(&m.LazyEvictions),
		ScheduledEvictions:    atomic.LoadUint32(&m.ScheduledEvictions),
		CleanupRuns:           atomic.LoadUint32(&m.CleanupRuns),
		CleanupTimeouts:       atomic.LoadUint32(&m.CleanupTimeouts),
		LastCleanupDurationMs: atomic.LoadUint64(&m.LastCleanupDurationMs),
		LastCleanupItems:      atomic.LoadUint32(&m.LastCleanupItems),
	}
}

func (m *Metrics) AddHit() {
	atomic.AddUint32(&m.Hits, 1)
}

func (m *Metrics) AddMiss() {
	atomic.AddUint32(&m.Misses, 1)
}

func (m *Metrics) AddSet() {
	atomic.AddUint32(&m.Sets, 1)
}

func (m *Metrics) AddDelete(count uint32) {
	if count > 0 {
		atomic.AddUint32(&m.Deletes, count)
	}
}

func (m *Metrics) AddLazyEviction() {
	atomic.AddUint32(&m.LazyEvictions, 1)
}

func (m *Metrics) AddScheduledEviction(count uint32) {
	if count > 0 {
		atomic.AddUint32(&m.ScheduledEvictions, count)
	}
}

func (m *Metrics) AddCleanupRun(duration time.Duration, itemsCleaned uint32, timedOut bool) {
	atomic.AddUint32(&m.CleanupRuns, 1)

	if timedOut {
		atomic.AddUint32(&m.CleanupTimeouts, 1)
	}

	atomic.StoreUint32(&m.LastCleanupItems, itemsCleaned)

	ms := duration.Milliseconds()

	var ums uint64
	if ms > 0 {
		ums = uint64(ms)
	} else {
		ums = 0
	}

	atomic.StoreUint64(&m.LastCleanupDurationMs, ums)
}

// JSONStr returns a JSON snapshot of the current metrics.
func (m *Metrics) JSONStr() string {
	b, err := json.Marshal(m.Snapshot())
	if err != nil {
		return fmt.Sprintf(`"error":"%v"`, err)
	}

	return string(b)
}
