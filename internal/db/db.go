package db

import (
	"context"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/pkg/errors"
)

func Tx(ctx context.Context, pool *pgxpool.Pool, op func(tx pgx.Tx) error) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return errors.Wrap(err, "could not start transaction")
	}
	// noinspection GoUnhandledErrorResult
	defer tx.Rollback(ctx) // nolint

	err = op(tx)
	if err != nil {
		return err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return errors.Wrap(err, "could not commit transaction")
	}
	return nil
}

func IterRows(rows pgx.Rows, op func(rows pgx.Rows) error) error {
	// noinspection GoUnhandledErrorResult
	defer rows.Close()
	for rows.Next() {
		err := op(rows)
		if err != nil {
			return errors.WithStack(err)
		}
	}
	if err := rows.Err(); err != nil {
		return errors.Wrap(err, "could not read rows")
	}
	return nil
}
