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

// MemCache is a simple in-memory cache with optional TTL expiration.
//
// It performs eviction:
//   - lazily on Get (expired items are removed on access)
//   - periodically via a background cleaner (best-effort, capped per run)
//
// Close must be called to stop the background cleaner.
type MemCache struct {
	items     map[string]entry
	stopCh    chan struct{}
	metrics   Metrics
	log       zerolog.Logger
	mx        sync.RWMutex
	cleanerWG sync.WaitGroup
	closed    atomic.Bool
}

// New returns a MemCache using DefaultCleanupInterval.
func New(log zerolog.Logger) *MemCache {
	return WithDeleteInterval(DefaultCleanupInterval, log)
}

// WithDeleteInterval returns a MemCache that runs the background cleaner at the
// provided interval.
func WithDeleteInterval(cleanupInterval time.Duration, log zerolog.Logger) *MemCache {
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
func (mc *MemCache) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	select {
	case <-ctx.Done():
		mc.log.Error().
			Str("key", key).
			Err(ctx.Err()).
			Msg("set key aborted")

		return ErrAborted
	default:
		mc.mx.Lock()
		mc.items[key] = newEntry(value, ttl)
		mc.mx.Unlock()
		mc.metrics.AddSet()
	}

	return nil
}

// Get returns the cached value for key.
//
// If the entry is missing or expired, it returns ErrNotFound. Expired entries are
// removed lazily on access.
func (mc *MemCache) Get(_ context.Context, key string) (any, error) {
	mc.mx.RLock()
	val, ok := mc.items[key]
	mc.mx.RUnlock()

	if !ok {
		mc.metrics.AddMiss()
		return nil, ErrNotFound
	}

	// lazy invalidation
	if val.IsExpired() {
		mc.mx.Lock()
		delete(mc.items, key)
		mc.mx.Unlock()

		mc.metrics.AddLazyEviction()
		mc.metrics.AddMiss()

		return nil, ErrNotFound
	}

	mc.metrics.AddHit()

	return val.value, nil
}

// deleteKeys removes keys from cache and returns the count of deleted items.
// Caller must handle metrics tracking separately.
func (mc *MemCache) deleteKeys(ctx context.Context, keys []string) (uint32, error) {
	mc.mx.Lock()
	defer mc.mx.Unlock()

	deleted := uint32(0)

	for _, key := range keys {
		select {
		case <-ctx.Done():
			mc.log.Error().
				Err(ctx.Err()).
				Str("key", key).
				Msg("delete key aborted")

			return deleted, ErrAborted
		default:
			if _, exists := mc.items[key]; exists {
				delete(mc.items, key)

				deleted++
			}
		}
	}

	return deleted, nil
}

// Delete removes a set of keys from cache.
func (mc *MemCache) Delete(ctx context.Context, keys ...string) error {
	deleted, err := mc.deleteKeys(ctx, keys)
	if err != nil {
		return err
	}

	mc.metrics.AddDelete(deleted)

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
			ctx, cancel := context.WithTimeout(context.Background(), interval)

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

			var (
				totalCleaned uint32
				delErr       error
			)

			if len(keysToClean) > 0 {
				totalCleaned, delErr = mc.deleteKeys(ctx, keysToClean)
			}

			duration := time.Since(start)

			mc.metrics.AddScheduledEviction(totalCleaned)
			mc.metrics.AddCleanupRun(duration, totalCleaned, delErr != nil)

			if delErr != nil {
				mc.log.Error().Err(delErr).
					Dur("duration", duration).
					Dur("interval", interval).
					Uint32("items_cleaned", totalCleaned).
					Msg("cleanup was aborted")
			}

			cancel()
		case <-mc.stopCh:
			mc.log.Info().Msg("gracefully stopped cache cleaner")
			mc.cleanerWG.Done()

			return
		}
	}
}

// Metrics returns a snapshot of internal metrics.
func (mc *MemCache) Metrics() Metrics {
	return mc.metrics.Snapshot()
}

// MetricsJSON returns a JSON snapshot of internal metrics.
func (mc *MemCache) MetricsJSON() string {
	return mc.metrics.JSONStr()
}

// Size returns the current number of entries in the cache.
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
// If key is missing or expired, Digest returns 0.
func (mc *MemCache) Digest(_ context.Context, key string) Digest {
	mc.mx.RLock()
	defer mc.mx.RUnlock()

	val, ok := mc.items[key]
	if !ok {
		return 0
	}

	return val.Digest()
}
