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

package vertica_test

import (
	"context"
	"testing"
	"time"

	"github.com/patraden/toolkit/pkg/vertica"
	"github.com/stretchr/testify/require"
)

func testConnString(t *testing.T) vertica.ConnParams {
	t.Helper()

	return vertica.NewConnString("localhost", 5433, "dbadmin", "", "VMart")
}

func TestQAVerticaDBPing(t *testing.T) {
	t.Parallel()

	t.Skip("integration test")

	connStr := testConnString(t)

	db, err := vertica.NewDB(connStr)
	require.NoError(t, err)
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	err = db.Ping(ctx)
	require.NoError(t, err)
}
