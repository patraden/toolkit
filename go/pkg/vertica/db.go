package vertica

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/vertica/vertica-sql-go" // register Vertica driver
)

type DB struct {
	connParams ConnParams
	conn       *sql.DB
}

func NewDB(connParams ConnParams) (*DB, error) {
	connDB, err := sql.Open("vertica", connParams.GetString())
	if err != nil {
		return nil, err
	}

	return &DB{
		connParams: connParams,
		conn:       connDB,
	}, nil
}

func (db *DB) Ping(ctx context.Context) error {
	return db.conn.PingContext(ctx)
}

func (db *DB) ConnString() string {
	return db.connParams.ConnString()
}

func (db *DB) ConnParams() ConnParams {
	return db.connParams
}

func (db *DB) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return db.conn.ExecContext(ctx, query, args...)
}

func (db *DB) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	return db.conn.QueryRowContext(ctx, query, args...)
}

func (db *DB) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return db.conn.QueryContext(ctx, query, args...)
}

func (db *DB) WithTxOps(ctx context.Context, opts *sql.TxOptions, qFunc func(*sql.Tx) error) error {
	trx, err := db.conn.BeginTx(ctx, opts)
	if err != nil {
		return err
	}

	if err := qFunc(trx); err != nil {
		if rbErr := trx.Rollback(); rbErr != nil {
			return fmt.Errorf("trx failed: %v, trx rollback failed: %w", err.Error(), rbErr)
		}

		return err
	}

	return trx.Commit()
}

func (db *DB) WithTx(ctx context.Context, fn func(*sql.Tx) error) error {
	return db.WithTxOps(ctx, nil, fn)
}

func (db *DB) Close() error {
	if db.conn == nil {
		return nil
	}

	return db.conn.Close()
}
