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
	"fmt"
	"hash/fnv"
	"time"
)

type entry struct {
	value     any
	expiresAt time.Time
}

func newEntry(value any, ttl time.Duration) entry {
	e := entry{value: value}
	if ttl > 0 {
		e.expiresAt = time.Now().UTC().Add(ttl)
	}

	return e
}

func (e entry) IsExpired() bool {
	// Zero expiration time means "never expires".
	if e.expiresAt.IsZero() {
		return false
	}

	return time.Now().UTC().After(e.expiresAt)
}

func (e *entry) AsBytes() ([]byte, error) {
	b, ok := e.value.([]byte)
	if !ok {
		return nil, ErrType
	}

	return b, nil
}

func (e entry) Digest() Digest {
	if e.IsExpired() {
		return 0
	}

	hash := fnv.New64a()

	switch val := e.value.(type) {
	case nil:
		return 0
	case []byte:
		_, _ = hash.Write(val)
	case string:
		_, _ = hash.Write([]byte(val))
	case int, int8, int16, int32, int64:
		fmt.Fprintf(hash, "%d", val)
	case uint, uint8, uint16, uint32, uint64, uintptr:
		fmt.Fprintf(hash, "%d", val)
	case float32, float64:
		fmt.Fprintf(hash, "%g", val)
	case bool:
		_, _ = hash.Write([]byte{boolToUint8(val)})
	default:
		// Unsupported type: no stable, cheap digest defined.
		return 0
	}

	return Digest(hash.Sum64())
}

func boolToUint8(b bool) uint8 {
	if b {
		return 1
	}

	return 0
}
