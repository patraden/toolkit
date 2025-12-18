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
	"time"
)

// Digest is a stable, cheap fingerprint of a cached value.
// Implementations may return 0 when a key is missing or the value is expired.
type Digest uint64

// Cache is a minimal cache interface with TTL support.
//
// TTL semantics:
//   - ttl > 0: the entry expires at now+ttl
//   - ttl <= 0: the entry does not expire
//
// Implementations may evict expired items lazily (on reads) and/or via a background
// cleanup process.
type Cache interface {
	Set(ctx context.Context, key string, value any, ttl time.Duration) error
	Get(ctx context.Context, key string) (any, error)
	Delete(ctx context.Context, keys ...string) error
	Digest(ctx context.Context, key string) Digest
	Close(ctx context.Context) error
}
