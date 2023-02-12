package store

import (
	"context"
	"database/sql"
	"github.com/jacktantram/user-service/pkg/driver/v1/postgres"
	"github.com/jmoiron/sqlx"
)

type Store struct {
	db postgres.Client
}

type conn interface {
	QueryRowxContext(ctx context.Context, query string, args ...interface{}) *sqlx.Row
	Queryx(query string, args ...interface{}) (*sqlx.Rows, error)
	NamedQueryContext(ctx context.Context, query string, arg interface{}) (*sqlx.Rows, error)
	NamedExecContext(ctx context.Context, query string, arg interface{}) (sql.Result, error)
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
}

func NewStore(db postgres.Client) Store {
	return Store{db: db}
}

// ExecInTransaction allows db calls to be made in transactions across multiple db calls
func (r Store) ExecInTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	// already in a transaction
	if conn, ok := ctx.Value(connKey{}).(conn); conn != nil && ok {
		return fn(ctx)
	}

	tx, err := r.db.DB.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}
	if err = fn(context.WithValue(ctx, connKey{}, tx)); err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return rollbackErr
		}
		return err
	}
	if err = tx.Commit(); err != nil {
		return err
	}
	return nil
}

type connKey struct{}

func (r Store) connFromContext(ctx context.Context) conn {
	c := ctx.Value(connKey{})
	if conn, ok := c.(conn); ok {
		return conn
	}
	return r.db.DB
}
