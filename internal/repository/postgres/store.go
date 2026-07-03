package postgres

import (
	"errors"
	"log/slog"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nivas/server/pkg/apperror"
)

// Store implements repository.Store using PostgreSQL.
type Store struct {
	pool *pgxpool.Pool
}

func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

func mapPgError(err error, notFoundMsg string) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return apperror.NotFound(notFoundMsg)
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		if pgErr.Code == "23505" {
			mapped := mapUniqueViolation(pgErr)
			slog.Default().Warn("database unique violation",
				"constraint", pgErr.ConstraintName,
				"table", pgErr.TableName,
				"error_code", mapped.Code,
			)
			return mapped
		}
		slog.Default().Error("postgres error",
			"sql_state", pgErr.Code,
			"message", pgErr.Message,
			"detail", pgErr.Detail,
			"table", pgErr.TableName,
		)
		return apperror.Internal("database error", err)
	}

	slog.Default().Error("database error", "error", err)
	return apperror.Internal("database error", err)
}

func mapUniqueViolation(pgErr *pgconn.PgError) *apperror.AppError {
	constraint := strings.ToLower(pgErr.ConstraintName)
	switch {
	case strings.Contains(constraint, "staff") && strings.Contains(constraint, "email"):
		return apperror.DuplicateEmail("Staff member")
	case strings.Contains(constraint, "tenant") && strings.Contains(constraint, "email"):
		return apperror.DuplicateEmail("Tenant")
	case strings.Contains(constraint, "tenant") && strings.Contains(constraint, "phone"):
		return apperror.DuplicatePhone("Tenant")
	case strings.Contains(constraint, "tenant") && strings.Contains(constraint, "name"):
		return apperror.DuplicateName("Tenant")
	case strings.Contains(constraint, "room"):
		return apperror.DuplicateRoomNumber()
	case strings.Contains(constraint, "kitchen") && strings.Contains(constraint, "name"):
		return apperror.DuplicateName("Kitchen item")
	default:
		return apperror.Conflict("A record with the same value already exists")
	}
}
