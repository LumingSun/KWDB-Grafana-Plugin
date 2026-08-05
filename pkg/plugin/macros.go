package plugin

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

const macroTimeFormat = "2006-01-02 15:04:05"

var (
	timeFilterRe = regexp.MustCompile(`\$__timeFilter\(\s*([^)]+?)\s*\)`)
	timeGroupRe  = regexp.MustCompile(`\$__timeGroup\(\s*([^,]+?)\s*,\s*('[^']*'|[^)]+?)\s*\)`)
)

// ExpandMacros replaces Grafana time macros with KWDB-compatible SQL fragments.
func ExpandMacros(rawSql string, timeRange backend.TimeRange) string {
	from := fmt.Sprintf("'%s'::TIMESTAMP", timeRange.From.UTC().Format(macroTimeFormat))
	to := fmt.Sprintf("'%s'::TIMESTAMP", timeRange.To.UTC().Format(macroTimeFormat))

	sql := timeFilterRe.ReplaceAllString(rawSql, "${1} >= "+from+" AND ${1} <= "+to)
	sql = strings.ReplaceAll(sql, "$__timeFrom", from)
	sql = strings.ReplaceAll(sql, "$__timeTo", to)
	sql = timeGroupRe.ReplaceAllString(sql, "time_bucket(${1}, ${2})")
	return sql
}
