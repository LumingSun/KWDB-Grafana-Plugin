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

var writeKeywords = map[string]struct{}{
	"ALTER":    {},
	"COPY":     {},
	"CREATE":   {},
	"DELETE":   {},
	"DROP":     {},
	"GRANT":    {},
	"INSERT":   {},
	"INTO":     {},
	"MERGE":    {},
	"REVOKE":   {},
	"TRUNCATE": {},
	"UPDATE":   {},
	"UPSERT":   {},
}

// IsReadOnly reports whether the SQL statement starts with an allowed read-only keyword.
func IsReadOnly(sql string) bool {
	if !isAllowedFirstKeyword(sql) {
		return false
	}
	if _, ok := scanForbiddenSQL(sql); ok {
		return false
	}
	return true
}

func isAllowedFirstKeyword(sql string) bool {
	for {
		stripped := leadingCommentRe.ReplaceAllString(sql, "")
		if stripped == sql {
			break
		}
		sql = stripped
	}
	sql = strings.TrimSpace(sql)
	if sql == "" {
		return false
	}

	end := strings.IndexFunc(sql, func(r rune) bool {
		return !isIdentRune(r)
	})
	if end < 0 {
		end = len(sql)
	}
	switch strings.ToUpper(sql[:end]) {
	case "SELECT", "SHOW", "EXPLAIN", "WITH":
		return true
	default:
		return false
	}
}

// scanForbiddenSQL rejects statements that contain a second statement or write
// keywords outside string literals, quoted identifiers, and comments.
func scanForbiddenSQL(sql string) (string, bool) {
	wordStart := -1
	for i := 0; i < len(sql); {
		c := sql[i]
		switch {
		case c == '\'':
			i = skipSingleQuoted(sql, i)
		case c == '"':
			i = skipDoubleQuoted(sql, i)
		case c == '-' && i+1 < len(sql) && sql[i+1] == '-':
			i = skipLineComment(sql, i)
		case c == '/' && i+1 < len(sql) && sql[i+1] == '*':
			i = skipBlockComment(sql, i)
		case c == ';':
			if !restIsCommentOrWhitespace(sql[i+1:]) {
				return "multi-statement semicolon", true
			}
			i++
		case isWordByte(c):
			if wordStart < 0 {
				wordStart = i
			}
			i++
		default:
			if wordStart >= 0 {
				word := strings.ToUpper(sql[wordStart:i])
				if _, ok := writeKeywords[word]; ok {
					return word, true
				}
				wordStart = -1
			}
			i++
		}
	}
	if wordStart >= 0 {
		word := strings.ToUpper(sql[wordStart:])
		if _, ok := writeKeywords[word]; ok {
			return word, true
		}
	}
	return "", false
}

func skipSingleQuoted(sql string, start int) int {
	for i := start + 1; i < len(sql); i++ {
		if sql[i] == '\'' {
			if i+1 < len(sql) && sql[i+1] == '\'' {
				i++
				continue
			}
			return i + 1
		}
	}
	return len(sql)
}

func skipDoubleQuoted(sql string, start int) int {
	for i := start + 1; i < len(sql); i++ {
		if sql[i] == '"' {
			if i+1 < len(sql) && sql[i+1] == '"' {
				i++
				continue
			}
			return i + 1
		}
	}
	return len(sql)
}

func skipLineComment(sql string, start int) int {
	for i := start + 2; i < len(sql); i++ {
		if sql[i] == '\n' {
			return i + 1
		}
	}
	return len(sql)
}

func skipBlockComment(sql string, start int) int {
	for i := start + 2; i+1 < len(sql); i++ {
		if sql[i] == '*' && sql[i+1] == '/' {
			return i + 2
		}
	}
	return len(sql)
}

func restIsCommentOrWhitespace(s string) bool {
	for i := 0; i < len(s); {
		switch {
		case s[i] == ' ' || s[i] == '\t' || s[i] == '\r' || s[i] == '\n':
			i++
		case s[i] == '-' && i+1 < len(s) && s[i+1] == '-':
			i = skipLineComment(s, i)
		case s[i] == '/' && i+1 < len(s) && s[i+1] == '*':
			i = skipBlockComment(s, i)
		default:
			return false
		}
	}
	return true
}

func isWordByte(c byte) bool {
	return c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || c == '_'
}

func isIdentRune(r rune) bool {
	return r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r == '_'
}

// ExecuteQuery runs the SQL on the pool and returns the pgx result rows.
func ExecuteQuery(ctx context.Context, pool *pgxpool.Pool, sql string) (pgx.Rows, error) {
	if pool == nil {
		return nil, errors.New("database connection pool is not initialized")
	}
	return pool.Query(ctx, sql)
}
