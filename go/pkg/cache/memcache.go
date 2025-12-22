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
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/rs/zerolog"
)

const (
	// MaxDeletesPerRun is an upper bound on the number of expired items removed
	// by the background cleaner in a single tick.
	MaxDeletesPerRun = 10000
	// DefaultCleanupInterval is the default interval for the background cleaner.
	DefaultCleanupInterval = time.Minute
)

// MemCache is a thread-safe in-memory cache with optional TTL expiration.
//
// It performs eviction:
//   - lazily on Get (expired items are removed on access)
//   - periodically via a background cleaner (best-effort, capped per run)
//
// MemCache uses read-write locks to allow concurrent reads while ensuring
// thread safety. Write operations (Set, Delete) block readers, but reads
// (Get) use read locks for better concurrency.
//
// Close must be called to stop the background cleaner and release resources.
type MemCache struct {
	items     map[string]entry
	stopCh    chan struct{}
	metrics   Metrics
	log       zerolog.Logger
	mx        sync.RWMutex
	cleanerWG sync.WaitGroup
	closed    atomic.Bool
}

// New returns a MemCache using DefaultCleanupInterval for background cleanup.
// The returned cache starts a background goroutine that periodically removes
// expired entries. Call Close to stop the background cleaner.
func New(log zerolog.Logger) *MemCache {
	return WithDeleteInterval(DefaultCleanupInterval, log)
}

// WithDeleteInterval returns a MemCache that runs the background cleaner at the
// provided interval.
//
// The background cleaner removes expired entries in batches, processing at most
// MaxDeletesPerRun items per interval to bound cleanup work.
//
// If cleanupInterval is <= 0, DefaultCleanupInterval is used instead to avoid
// ticker panics.
//
// The returned cache starts a background goroutine. Call Close to stop it.
func WithDeleteInterval(cleanupInterval time.Duration, log zerolog.Logger) *MemCache {
	if cleanupInterval <= 0 {
		cleanupInterval = DefaultCleanupInterval
	}

	cache := &MemCache{
		items:     make(map[string]entry),
		stopCh:    make(chan struct{}),
		metrics:   Metrics{},
		log:       log,
		mx:        sync.RWMutex{},
		cleanerWG: sync.WaitGroup{},
		closed:    atomic.Bool{},
	}

	cache.cleanerWG.Add(1)
	go cache.cleaner(cleanupInterval)

	return cache
}

// Set stores key/value with the provided TTL.
//
// TTL semantics:
//   - ttl > 0: expires at now+ttl
//   - ttl <= 0: does not expire
//
// Set is safe for concurrent use. If the key already exists, it is overwritten.
func (mc *MemCache) Set(_ context.Context, key string, value any, ttl time.Duration) error {
	mc.mx.Lock()
	mc.items[key] = newEntry(value, ttl)
	mc.mx.Unlock()

	mc.metrics.AddSet()

	return nil
}

// Checks if key can be invalidated.
func (mc *MemCache) invalidated(key string) bool {
	mc.mx.Lock()
	defer mc.mx.Unlock()

	if val, ok := mc.items[key]; ok && val.IsExpired() {
		delete(mc.items, key)
		mc.metrics.AddLazyEviction()

		return true
	}

	return false
}

// Fast get with concurrent read.
func (mc *MemCache) get(key string) (entry, error) {
	mc.mx.RLock()
	defer mc.mx.RUnlock()

	val, ok := mc.items[key]
	if !ok {
		mc.metrics.AddMiss()
		return entry{}, ErrNotFound
	}

	return val, nil
}

// Get returns the cached value for key.
//
// If the entry is missing or expired, Get returns ErrNotFound.
// Expired entries are removed lazily on access (when Get is called).
//
// Get is safe for concurrent use. It uses read locks for fast access and
// only acquires a write lock when deleting expired entries. If a key is
// refreshed between the read and write lock acquisition, Get will return
// the fresh value.
func (mc *MemCache) Get(_ context.Context, key string) (any, error) {
	val, err := mc.get(key)
	if err != nil {
		return nil, err
	}

	// lazy invalidation
	if val.IsExpired() {
		if mc.invalidated(key) {
			mc.metrics.AddMiss()
			return nil, ErrNotFound
		}

		val, err = mc.get(key)
		if err != nil {
			return nil, err
		}
	}

	mc.metrics.AddHit()

	return val.value, nil
}

// Delete removes a set of keys from cache.
//
// Delete is safe for concurrent use. If a key does not exist, it is ignored.
// The operation can be cancelled via the context.
// If cancelled, Delete returns ErrAborted.
// Metrics are still updated for keys deleted before cancellation.
func (mc *MemCache) Delete(ctx context.Context, keys ...string) error {
	mc.mx.Lock()
	defer mc.mx.Unlock()

	deleted := uint32(0)
	defer func() { mc.metrics.AddDelete(deleted) }()

	for _, key := range keys {
		select {
		case <-ctx.Done():
			mc.log.Error().
				Err(ctx.Err()).
				Str("key", key).
				Msg("delete key aborted")

			return ErrAborted
		default:
			if _, exists := mc.items[key]; exists {
				delete(mc.items, key)

				deleted++
			}
		}
	}

	return nil
}

func (mc *MemCache) cleaner(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	mc.log.Info().Msg("started cache cleaner")

	for {
		select {
		case <-ticker.C:
			start := time.Now()

			mc.mx.RLock()

			keysToClean := make([]string, 0, MaxDeletesPerRun)
			for key, val := range mc.items {
				if len(keysToClean) >= MaxDeletesPerRun {
					break
				}

				if val.IsExpired() {
					keysToClean = append(keysToClean, key)
				}
			}

			mc.mx.RUnlock()

			deleted := uint32(0)

			for _, key := range keysToClean {
				if mc.invalidated(key) {
					deleted++
				}
			}

			duration := time.Since(start)

			mc.metrics.AddScheduledEviction(deleted)
			mc.metrics.AddCleanupRun(duration, deleted)
		case <-mc.stopCh:
			mc.log.Info().Msg("gracefully stopped cache cleaner")
			mc.cleanerWG.Done()

			return
		}
	}
}

// Metrics returns a snapshot of internal metrics.
//
// The returned snapshot is a point-in-time copy of all metrics, safe for
// concurrent access. Metrics include hits, misses, sets, deletes, and
// eviction statistics.
func (mc *MemCache) Metrics() Metrics {
	return mc.metrics.Snapshot()
}

// MetricsJSON returns a JSON snapshot of internal metrics as a string.
//
// The returned JSON is a point-in-time snapshot, safe for concurrent access.
// Useful for logging or monitoring endpoints.
func (mc *MemCache) MetricsJSON() string {
	return mc.metrics.JSONStr()
}

// Size returns the current number of entries in the cache.
//
// Size includes both expired and non-expired entries. Expired entries are
// removed lazily on Get or by the background cleaner, so Size may include
// entries that would return ErrNotFound on Get.
func (mc *MemCache) Size() int {
	mc.mx.RLock()
	defer mc.mx.RUnlock()

	return len(mc.items)
}

// Close stops the background cleaner. It is safe to call Close multiple times.
func (mc *MemCache) Close(_ context.Context) error {
	if mc.closed.CompareAndSwap(false, true) {
		close(mc.stopCh)
	}

	mc.cleanerWG.Wait()

	return nil
}

// Digest returns a fingerprint for the current (non-expired) value of key.
//
// If key is missing or expired, Digest returns 0. The digest is computed
// using FNV-1a hash of the value's string representation.
//
// Digest is safe for concurrent use. It does not remove expired entries;
// use Get for lazy eviction.
func (mc *MemCache) Digest(_ context.Context, key string) Digest {
	mc.mx.RLock()
	defer mc.mx.RUnlock()

	val, ok := mc.items[key]
	if !ok {
		return 0
	}

	return val.Digest()
}
