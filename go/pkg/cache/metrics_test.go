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

package cache_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/patraden/toolkit/pkg/cache"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMetricsCounters(t *testing.T) {
	t.Parallel()

	var mtrcs cache.Metrics

	mtrcs.AddHit()
	mtrcs.AddHit()
	mtrcs.AddMiss()
	mtrcs.AddSet()
	mtrcs.AddSet()
	mtrcs.AddSet()
	mtrcs.AddDelete(0) // should be a no-op
	mtrcs.AddDelete(2) // increments by 2
	mtrcs.AddLazyEviction()
	mtrcs.AddScheduledEviction(0) // no-op
	mtrcs.AddScheduledEviction(3)

	snps := mtrcs.Snapshot()

	assert.Equal(t, uint32(2), snps.Hits)
	assert.Equal(t, uint32(1), snps.Misses)
	assert.Equal(t, uint32(3), snps.Sets)
	assert.Equal(t, uint32(2), snps.Deletes)
	assert.Equal(t, uint32(1), snps.LazyEvictions)
	assert.Equal(t, uint32(3), snps.ScheduledEvictions)
}

func TestMetricsCleanupRun(t *testing.T) {
	t.Parallel()

	var mtrcs cache.Metrics

	mtrcs.AddCleanupRun(0, 0)

	snps := mtrcs.Snapshot()
	assert.Equal(t, uint32(1), snps.CleanupRuns)
	assert.Equal(t, uint32(0), snps.LastCleanupItems)
	assert.Equal(t, uint64(0), snps.LastCleanupDurationMs)

	mtrcs.AddCleanupRun(10*time.Millisecond, 5)

	snps = mtrcs.Snapshot()
	assert.Equal(t, uint32(2), snps.CleanupRuns)
	assert.Equal(t, uint32(5), snps.LastCleanupItems)
	assert.Equal(t, uint64(10), snps.LastCleanupDurationMs)
}

func TestMetricsJSONStr(t *testing.T) {
	t.Parallel()

	var (
		mtrcs   cache.Metrics
		decoded cache.Metrics
	)

	mtrcs.AddHit()
	mtrcs.AddMiss()

	js := mtrcs.JSONStr()

	err := json.Unmarshal([]byte(js), &decoded)
	require.NoError(t, err, "JSONStr should return valid JSON")

	snps := mtrcs.Snapshot()
	assert.Equal(t, snps.Hits, decoded.Hits)
	assert.Equal(t, snps.Misses, decoded.Misses)
	assert.Equal(t, snps.Sets, decoded.Sets)
	assert.Equal(t, snps.Deletes, decoded.Deletes)
	assert.Equal(t, snps.LazyEvictions, decoded.LazyEvictions)
	assert.Equal(t, snps.ScheduledEvictions, decoded.ScheduledEvictions)
	assert.Equal(t, snps.CleanupRuns, decoded.CleanupRuns)
	assert.Equal(t, snps.LastCleanupDurationMs, decoded.LastCleanupDurationMs)
	assert.Equal(t, snps.LastCleanupItems, decoded.LastCleanupItems)
}
