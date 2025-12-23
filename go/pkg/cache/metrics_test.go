package cache

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMetricsCounters(t *testing.T) {
	t.Parallel()

	var m Metrics

	m.AddHit()
	m.AddHit()
	m.AddMiss()
	m.AddSet()
	m.AddSet()
	m.AddSet()
	m.AddDelete(0) // should be a no-op
	m.AddDelete(2) // increments by 2
	m.AddLazyEviction()
	m.AddScheduledEviction(0) // no-op
	m.AddScheduledEviction(3)

	s := m.Snapshot()

	assert.Equal(t, uint32(2), s.Hits)
	assert.Equal(t, uint32(1), s.Misses)
	assert.Equal(t, uint32(3), s.Sets)
	assert.Equal(t, uint32(2), s.Deletes)
	assert.Equal(t, uint32(1), s.LazyEvictions)
	assert.Equal(t, uint32(3), s.ScheduledEvictions)
}

func TestMetricsCleanupRun(t *testing.T) {
	t.Parallel()

	var m Metrics

	m.AddCleanupRun(0, 0)

	s := m.Snapshot()
	assert.Equal(t, uint32(1), s.CleanupRuns)
	assert.Equal(t, uint32(0), s.LastCleanupItems)
	assert.Equal(t, uint64(0), s.LastCleanupDurationMs)

	m.AddCleanupRun(10*time.Millisecond, 5)

	s = m.Snapshot()
	assert.Equal(t, uint32(2), s.CleanupRuns)
	assert.Equal(t, uint32(5), s.LastCleanupItems)
	assert.Equal(t, uint64(10), s.LastCleanupDurationMs)
}

func TestMetricsJSONStr(t *testing.T) {
	t.Parallel()

	var m Metrics
	m.AddHit()
	m.AddMiss()

	js := m.JSONStr()

	var decoded Metrics
	err := json.Unmarshal([]byte(js), &decoded)
	assert.NoError(t, err, "JSONStr should return valid JSON")

	s := m.Snapshot()
	assert.Equal(t, s.Hits, decoded.Hits)
	assert.Equal(t, s.Misses, decoded.Misses)
	assert.Equal(t, s.Sets, decoded.Sets)
	assert.Equal(t, s.Deletes, decoded.Deletes)
	assert.Equal(t, s.LazyEvictions, decoded.LazyEvictions)
	assert.Equal(t, s.ScheduledEvictions, decoded.ScheduledEvictions)
	assert.Equal(t, s.CleanupRuns, decoded.CleanupRuns)
	assert.Equal(t, s.LastCleanupDurationMs, decoded.LastCleanupDurationMs)
	assert.Equal(t, s.LastCleanupItems, decoded.LastCleanupItems)
}
