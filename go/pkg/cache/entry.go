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

	h := fnv.New64a()
	fmt.Fprint(h, e.value)

	return Digest(h.Sum64())
}
