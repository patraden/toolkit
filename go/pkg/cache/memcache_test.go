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
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/patraden/toolkit/pkg/cache"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Logger(t *testing.T) zerolog.Logger {
	t.Helper()

	return zerolog.New(os.Stdout).With().Timestamp().Caller().Logger().Level(zerolog.DebugLevel)
}

func TestMemCacheStress(t *testing.T) {
	t.Parallel()

	log := Logger(t)

	t.Run("parallel set/get/delete has consistent metrics", func(t *testing.T) {
		t.Parallel()

		mcache := cache.WithDeleteInterval(time.Hour, log)

		t.Cleanup(func() {
			require.NoError(t, mcache.Close(t.Context()))
		})

		const keys = 500_000

		var wg sync.WaitGroup

		wg.Add(keys)

		for key := range keys {
			go func(k int) {
				defer wg.Done()

				err := mcache.Set(t.Context(), fmt.Sprintf("k-%d", k), k, 0)
				require.NoError(t, err)
			}(key)
		}

		wg.Wait()
		wg.Add(keys)

		for key := range keys {
			go func(k int) {
				defer wg.Done()

				v, err := mcache.Get(t.Context(), fmt.Sprintf("k-%d", k))
				require.NoError(t, err)
				assert.Equal(t, k, v)
			}(key)
		}

		wg.Wait()
		wg.Add(keys)

		for key := range keys {
			go func(k int) {
				defer wg.Done()

				err := mcache.Delete(t.Context(), fmt.Sprintf("k-%d", k))
				require.NoError(t, err)
			}(key)
		}

		wg.Wait()
		wg.Add(keys)

		for key := range keys {
			go func(k int) {
				defer wg.Done()

				_, err := mcache.Get(t.Context(), fmt.Sprintf("k-%d", k))

				require.Error(t, err)
				require.ErrorIs(t, err, cache.ErrNotFound)
			}(key)
		}

		wg.Wait()

		mtcs := mcache.Metrics()

		assert.Equal(t, uint32(keys), mtcs.Sets)
		assert.Equal(t, uint32(keys), mtcs.Hits)
		assert.Equal(t, uint32(keys), mtcs.Deletes)
		assert.Equal(t, uint32(keys), mtcs.Misses)
		assert.Equal(t, uint32(0), mtcs.LazyEvictions, "No TTL used here, so there shouldn't be any evictions.")
	})

	t.Run("lazy eviction under concurrent gets", func(t *testing.T) {
		t.Parallel()

		mcache := cache.WithDeleteInterval(time.Hour, log)

		t.Cleanup(func() {
			require.NoError(t, mcache.Close(t.Context()))
		})

		const keys = 500_000

		var wg sync.WaitGroup

		wg.Add(keys)

		for key := range keys {
			go func(k int) {
				defer wg.Done()
				require.NoError(t, mcache.Set(t.Context(), fmt.Sprintf("e-%d", k), k, 10*time.Millisecond))
			}(key)
		}

		wg.Wait()

		time.Sleep(20 * time.Millisecond)

		wg.Add(keys)

		for key := range keys {
			go func(k int) {
				defer wg.Done()

				_, err := mcache.Get(t.Context(), fmt.Sprintf("e-%d", k))

				require.Error(t, err)
				require.ErrorIs(t, err, cache.ErrNotFound)
			}(key)
		}

		wg.Wait()

		m := mcache.Metrics()

		assert.Equal(t, uint32(keys), m.LazyEvictions)
		assert.Equal(t, uint32(keys), m.Misses)
		assert.Equal(t, uint32(0), m.Hits)
	})
}
