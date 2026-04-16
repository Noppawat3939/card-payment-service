package repository

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

const (
	PgUniqueViolation     = "23505"
	PgForeignKeyViolation = "23503"
)

func IsPostgresCode(err error, code string) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == code
}
