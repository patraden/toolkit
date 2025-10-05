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

	return vertica.NewConnString("qa-etlgavdb.rtty.in", 5433, "dbadmin", "", "bi")
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
