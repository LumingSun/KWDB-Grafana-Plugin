package plugin

import (
	"context"
	"testing"
)

func TestIsReadOnly(t *testing.T) {
	tests := []struct {
		name string
		sql  string
		want bool
	}{
		{name: "select", sql: "SELECT * FROM sensors", want: true},
		{name: "show", sql: "SHOW TABLES", want: true},
		{name: "explain", sql: "EXPLAIN SELECT 1", want: true},
		{name: "with", sql: "WITH x AS (SELECT 1) SELECT * FROM x", want: true},
		{name: "leading whitespace", sql: "   \n\t SELECT 1", want: true},
		{name: "line comment", sql: "-- comment\nSELECT 1", want: true},
		{name: "block comment", sql: "/* comment */ SHOW TABLES", want: true},
		{name: "multiple comments", sql: "-- first\n/* second */ SELECT 1", want: true},
		{name: "trailing semicolon", sql: "SELECT 1;", want: true},
		{name: "trailing semicolon and comment", sql: "SELECT 1; -- done", want: true},
		{name: "semicolon in string", sql: "SELECT ';'", want: true},
		{name: "semicolon in quoted identifier", sql: `SELECT "a;b" FROM t`, want: true},
		{name: "semicolon in block comment", sql: "SELECT 1 /* ; */", want: true},
		{name: "multi statement", sql: "SELECT 1; SELECT 2", want: false},
		{name: "select into", sql: "SELECT * INTO new_table FROM t", want: false},
		{name: "with data modifying cte", sql: "WITH x AS (DELETE FROM t) SELECT * FROM x", want: false},
		{name: "write keyword in string", sql: "SELECT 'delete'", want: true},
		{name: "quoted write identifier", sql: `SELECT "delete" FROM t`, want: true},
		{name: "insert", sql: "INSERT INTO t VALUES (1)", want: false},
		{name: "update", sql: "UPDATE t SET a = 1", want: false},
		{name: "delete", sql: "DELETE FROM t", want: false},
		{name: "drop", sql: "DROP TABLE t", want: false},
		{name: "create", sql: "CREATE TABLE t (a INT)", want: false},
		{name: "alter", sql: "ALTER TABLE t ADD COLUMN b INT", want: false},
		{name: "truncate", sql: "TRUNCATE t", want: false},
		{name: "empty", sql: "", want: false},
		{name: "whitespace only", sql: "  \n ", want: false},
		{name: "comment only", sql: "-- just a comment", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsReadOnly(tt.sql); got != tt.want {
				t.Errorf("IsReadOnly(%q) = %v, want %v", tt.sql, got, tt.want)
			}
		})
	}
}

func TestExecuteQueryNilPool(t *testing.T) {
	if _, err := ExecuteQuery(context.Background(), nil, "SELECT 1"); err == nil {
		t.Fatal("expected error for nil pool")
	}
}
