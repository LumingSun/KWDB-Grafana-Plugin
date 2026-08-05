package plugin

import (
	"context"
	"errors"
	"regexp"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var leadingCommentRe = regexp.MustCompile(`(?s)^\s*(?:--[^\n]*(?:\n|$)|\/\*.*?\*\/)`)

// IsReadOnly reports whether the SQL statement starts with an allowed read-only keyword.
func IsReadOnly(sql string) bool {
	switch firstKeyword(sql) {
	case "SELECT", "SHOW", "EXPLAIN", "WITH":
		return true
	default:
		return false
	}
}

func firstKeyword(sql string) string {
	for {
		stripped := leadingCommentRe.ReplaceAllString(sql, "")
		if stripped == sql {
			break
		}
		sql = stripped
	}
	sql = strings.TrimSpace(sql)
	if sql == "" {
		return ""
	}

	end := strings.IndexFunc(sql, func(r rune) bool {
		return !(r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r == '_')
	})
	if end < 0 {
		end = len(sql)
	}
	return strings.ToUpper(sql[:end])
}

// ExecuteQuery runs the SQL on the pool and returns the pgx result rows.
func ExecuteQuery(ctx context.Context, pool *pgxpool.Pool, sql string) (pgx.Rows, error) {
	if pool == nil {
		return nil, errors.New("database connection pool is not initialized")
	}
	return pool.Query(ctx, sql)
}
