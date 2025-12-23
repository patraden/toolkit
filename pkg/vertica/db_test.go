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

//nolint:testpackage // white-box tests require access to internal fields
package vertica

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

func newTestDBWithMock(t *testing.T) (*DB, sqlmock.Sqlmock, func()) {
	t.Helper()

	mockDB, mock, err := sqlmock.New(
		sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp),
		sqlmock.MonitorPingsOption(true),
	)
	require.NoError(t, err)

	params := NewConnString("host", 5433, "user", "pass", "db")
	db := &DB{connParams: params, conn: mockDB}

	cleanup := func() {
		_ = db.Close()
	}

	return db, mock, cleanup
}

func TestDBPing(t *testing.T) {
	t.Parallel()

	db, mock, cleanup := newTestDBWithMock(t)

	defer cleanup()

	mock.ExpectPing().WillReturnError(nil)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)

	defer cancel()

	err := db.Ping(ctx)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestDBExec(t *testing.T) {
	t.Parallel()

	db, mock, cleanup := newTestDBWithMock(t)

	defer cleanup()

	mock.ExpectExec("INSERT INTO foo\\(id, name\\) VALUES\\(\\?, \\?\\)").
		WithArgs(1, "alice").
		WillReturnResult(sqlmock.NewResult(1, 1))

	res, err := db.ExecContext(
		context.Background(),
		"INSERT INTO foo(id, name) VALUES(?, ?)",
		1,
		"alice",
	)
	require.NoError(t, err)
	affected, err := res.RowsAffected()
	require.NoError(t, err)
	require.Equal(t, int64(1), affected)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestDBQueryRow(t *testing.T) {
	t.Parallel()

	db, mock, cleanup := newTestDBWithMock(t)

	defer cleanup()

	rows := sqlmock.NewRows([]string{"id", "name"}).AddRow(42, "bob")
	mock.ExpectQuery("SELECT id, name FROM foo WHERE id = \\?").
		WithArgs(42).
		WillReturnRows(rows)

	var gotID int

	var name string
	err := db.QueryRowContext(
		context.Background(),
		"SELECT id, name FROM foo WHERE id = ?",
		42,
	).Scan(&gotID, &name)
	require.NoError(t, err)
	require.Equal(t, 42, gotID)
	require.Equal(t, "bob", name)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestDBQuery(t *testing.T) {
	t.Parallel()

	db, mock, cleanup := newTestDBWithMock(t)

	defer cleanup()

	rows := sqlmock.NewRows([]string{"n"}).AddRow(1).AddRow(2).AddRow(3)
	mock.ExpectQuery("SELECT n FROM nums").WillReturnRows(rows)

	queryRows, err := db.QueryContext(context.Background(), "SELECT n FROM nums")
	require.NoError(t, err)

	defer queryRows.Close()

	var nums []int

	for queryRows.Next() {
		var n int

		require.NoError(t, queryRows.Scan(&n))

		nums = append(nums, n)
	}

	require.NoError(t, queryRows.Err())
	require.Equal(t, []int{1, 2, 3}, nums)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestDBTxCommit(t *testing.T) {
	t.Parallel()

	db, mock, cleanup := newTestDBWithMock(t)

	defer cleanup()

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE accounts SET balance = balance - \\? WHERE id = \\?").
		WithArgs(100, 1).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err := db.WithTxOps(
		context.Background(),
		&sql.TxOptions{Isolation: sql.LevelSerializable, ReadOnly: false},
		func(tx *sql.Tx) error {
			_, err := tx.Exec(
				"UPDATE accounts SET balance = balance - ? WHERE id = ?",
				100,
				1,
			)

			return err
		},
	)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestDBTxRollback(t *testing.T) {
	t.Parallel()

	db, mock, cleanup := newTestDBWithMock(t)

	defer cleanup()

	mock.ExpectBegin()
	mock.ExpectRollback()

	wantErr := sql.ErrNoRows
	err := db.WithTx(context.Background(), func(_ *sql.Tx) error {
		return wantErr
	})
	require.True(t, errors.Is(err, wantErr)) //nolint:testifylint // testify v1.3.0 lacks require.ErrorIs
	require.NoError(t, mock.ExpectationsWereMet())
}
