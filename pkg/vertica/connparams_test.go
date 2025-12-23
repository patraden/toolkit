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
	"testing"

	"github.com/patraden/toolkit/pkg/vertica"
	"github.com/stretchr/testify/require"
)

func TestConnParams(t *testing.T) {
	t.Parallel()

	t.Run("getstring_no_params", func(t *testing.T) {
		t.Parallel()

		cp := vertica.NewConnString("localhost", 5433, "user", "p@ s#", "db")
		s := cp.GetString()
		require.Equal(t, "vertica://user:p%40+s%23@localhost:5433/db", s)
	})

	t.Run("getstring_params_sorted_escaped", func(t *testing.T) {
		t.Parallel()

		cp := vertica.NewConnString("h", 1, "u", "p", "d")
		cp.Params["z"] = "9 9"
		cp.Params["a"] = "@!"

		s := cp.GetString()
		require.Equal(t, "vertica://u:p@h:1/d?a=%40%21&z=9+9", s)
	})

	t.Run("connstring_masked_password", func(t *testing.T) {
		t.Parallel()

		cp := vertica.NewConnString("host", 1234, "me", "secret", "db")
		s := cp.ConnString()
		require.Equal(t, "vertica://me:***@host:1234/db", s)
	})

	t.Run("getpassword", func(t *testing.T) {
		t.Parallel()

		cp := vertica.NewConnString("host", 1234, "me", "secret", "db")
		require.Equal(t, "secret", cp.GetPassword())
	})
}
