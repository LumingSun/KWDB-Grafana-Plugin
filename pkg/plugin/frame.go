package plugin

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/data"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type columnKind int

const (
	kindString columnKind = iota
	kindInt64
	kindInt32
	kindFloat64
	kindBool
	kindTime
)

type columnPlan struct {
	name   string
	oid    uint32
	kind   columnKind
	values []any
	labels data.Labels
}

// DefaultMaxRows caps the number of rows buffered into a single frame.
const DefaultMaxRows = 100000

// RowsToFrame converts pgx result rows into a Grafana Data Frame.
func RowsToFrame(rows pgx.Rows, format string, timeColumn string, tags []string) (*data.Frame, error) {
	return rowsToFrame(rows, format, timeColumn, tags, DefaultMaxRows)
}

func rowsToFrame(rows pgx.Rows, format string, timeColumn string, tags []string, maxRows int) (*data.Frame, error) {
	defer rows.Close()

	descriptions := rows.FieldDescriptions()
	plans := make([]columnPlan, len(descriptions))
	for i, fd := range descriptions {
		plans[i] = columnPlan{
			name: fd.Name,
			oid:  fd.DataTypeOID,
			kind: kindForOID(fd.DataTypeOID),
		}
	}

	truncated := false
	rowCount := 0
	for rows.Next() {
		if maxRows > 0 && rowCount >= maxRows {
			truncated = true
			break
		}
		values, err := rows.Values()
		if err != nil {
			return nil, err
		}
		if len(values) != len(plans) {
			return nil, fmt.Errorf("row has %d values but %d fields were described", len(values), len(plans))
		}
		for i := range plans {
			value, err := convertValue(plans[i].kind, values[i])
			if err != nil {
				return nil, fmt.Errorf("column %q: %w", plans[i].name, err)
			}
			plans[i].values = append(plans[i].values, value)
		}
		rowCount++
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	if format == "time_series" {
		if err := prepareTimeSeries(plans, timeColumn, tags); err != nil {
			return nil, err
		}
	}

	frame := data.NewFrame("response")
	for _, plan := range plans {
		frame.Fields = append(frame.Fields, buildField(plan))
	}

	if format == "time_series" {
		rowLen, err := frame.RowLen()
		if err != nil {
			return nil, err
		}
		if rowLen == 0 {
			return frame, nil
		}
		switch frame.TimeSeriesSchema().Type {
		case data.TimeSeriesTypeLong:
			wide, err := data.LongToWide(frame, nil)
			if err != nil {
				return nil, err
			}
			if truncated {
				applyMaxRowsNotice(wide, maxRows)
			}
			return wide, nil
		case data.TimeSeriesTypeWide:
			if truncated {
				applyMaxRowsNotice(frame, maxRows)
			}
			return frame, nil
		default:
			return nil, fmt.Errorf("could not convert frame to time series")
		}
	}
	if truncated {
		applyMaxRowsNotice(frame, maxRows)
	}
	return frame, nil
}

func applyMaxRowsNotice(frame *data.Frame, maxRows int) {
	if frame.Meta == nil {
		frame.Meta = &data.FrameMeta{}
	}
	frame.Meta.Notices = append(frame.Meta.Notices, data.Notice{
		Severity: data.NoticeSeverityWarning,
		Text:     fmt.Sprintf("Query result was truncated at %d rows; add LIMIT or narrow the time range", maxRows),
	})
}

func kindForOID(oid uint32) columnKind {
	switch oid {
	case 20, 21:
		return kindInt64
	case 23, 26:
		return kindInt32
	case 700, 701:
		return kindFloat64
	case 1700:
		return kindFloat64
	case 16:
		return kindBool
	case 1083, 1114, 1184:
		return kindTime
	case 25, 1043:
		return kindString
	default:
		return kindString
	}
}

func convertValue(kind columnKind, v any) (any, error) {
	if v == nil {
		return nil, nil
	}
	switch kind {
	case kindInt64:
		switch n := v.(type) {
		case int64:
			return n, nil
		case int32:
			return int64(n), nil
		case int16:
			return int64(n), nil
		case int:
			return int64(n), nil
		case float64:
			return int64(n), nil
		}
		return nil, fmt.Errorf("cannot convert %T to int64", v)
	case kindInt32:
		switch n := v.(type) {
		case int32:
			return n, nil
		case int16:
			return int32(n), nil
		case int64:
			return int32(n), nil
		case uint32:
			return int32(n), nil
		case int:
			return int32(n), nil
		}
		return nil, fmt.Errorf("cannot convert %T to int32", v)
	case kindFloat64:
		switch n := v.(type) {
		case float64:
			return n, nil
		case float32:
			return float64(n), nil
		case pgtype.Numeric:
			f, err := n.Float64Value()
			if err != nil {
				return nil, fmt.Errorf("cannot convert numeric: %w", err)
			}
			if !f.Valid {
				return nil, nil
			}
			return f.Float64, nil
		case int64:
			return float64(n), nil
		case int32:
			return float64(n), nil
		case int16:
			return float64(n), nil
		}
		return nil, fmt.Errorf("cannot convert %T to float64", v)
	case kindBool:
		if b, ok := v.(bool); ok {
			return b, nil
		}
		return nil, fmt.Errorf("cannot convert %T to bool", v)
	case kindString:
		switch s := v.(type) {
		case string:
			return s, nil
		case []byte:
			return string(s), nil
		default:
			return fmt.Sprint(v), nil
		}
	case kindTime:
		if t, ok := v.(time.Time); ok {
			return t, nil
		}
		if t, ok := v.(pgtype.Time); ok && t.Valid {
			return time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC).Add(time.Duration(t.Microseconds) * time.Microsecond), nil
		}
		return nil, fmt.Errorf("cannot convert %T to time.Time", v)
	}
	return nil, fmt.Errorf("unsupported column kind %d", kind)
}

func prepareTimeSeries(plans []columnPlan, timeColumn string, tags []string) error {
	timeIdx := findTimeColumn(plans, timeColumn)
	if timeIdx < 0 {
		return fmt.Errorf("time_series format requires a time column")
	}
	sortPlansByTime(plans, timeIdx)
	plans[timeIdx].name = "time"

	if len(tags) > 0 {
		for i := range plans {
			if containsFold(tags, plans[i].name) {
				plans[i].labels = data.Labels{}
			}
		}
	}
	return nil
}

func findTimeColumn(plans []columnPlan, timeColumn string) int {
	if timeColumn != "" {
		for i := range plans {
			if plans[i].name == timeColumn {
				return i
			}
		}
	}
	for i := range plans {
		if plans[i].kind == kindTime {
			return i
		}
	}
	for _, name := range []string{"time", "ts", "k_timestamp", "timestamp"} {
		for i := range plans {
			if plans[i].kind == kindTime && strings.EqualFold(plans[i].name, name) {
				return i
			}
		}
	}
	return -1
}

func sortPlansByTime(plans []columnPlan, timeIdx int) {
	times := plans[timeIdx].values
	order := make([]int, len(times))
	for i := range order {
		order[i] = i
	}
	sort.SliceStable(order, func(i, j int) bool {
		return asTime(times[order[i]]).Before(asTime(times[order[j]]))
	})
	for i := range plans {
		values := plans[i].values
		reordered := make([]any, len(values))
		for j, src := range order {
			reordered[j] = values[src]
		}
		plans[i].values = reordered
	}
}

func asTime(v any) time.Time {
	if t, ok := v.(time.Time); ok {
		return t
	}
	return time.Time{}
}

func containsFold(items []string, target string) bool {
	for _, item := range items {
		if strings.EqualFold(item, target) {
			return true
		}
	}
	return false
}

func buildField(plan columnPlan) *data.Field {
	if plan.labels != nil && plan.kind != kindTime && plan.kind != kindString {
		return buildTagField(plan)
	}
	if hasNil(plan.values) {
		return buildNullableField(plan)
	}
	switch plan.kind {
	case kindInt64:
		return data.NewField(plan.name, plan.labels, asValues[int64](plan.values))
	case kindInt32:
		return data.NewField(plan.name, plan.labels, asValues[int32](plan.values))
	case kindFloat64:
		return data.NewField(plan.name, plan.labels, asValues[float64](plan.values))
	case kindBool:
		return data.NewField(plan.name, plan.labels, asValues[bool](plan.values))
	case kindTime:
		return data.NewField(plan.name, plan.labels, asValues[time.Time](plan.values))
	default:
		return data.NewField(plan.name, plan.labels, asValues[string](plan.values))
	}
}

func buildTagField(plan columnPlan) *data.Field {
	if hasNil(plan.values) {
		values := make([]*string, len(plan.values))
		for i, v := range plan.values {
			if v != nil {
				s := fmt.Sprint(v)
				values[i] = &s
			}
		}
		return data.NewField(plan.name, plan.labels, values)
	}
	values := make([]string, len(plan.values))
	for i, v := range plan.values {
		values[i] = fmt.Sprint(v)
	}
	return data.NewField(plan.name, plan.labels, values)
}

func hasNil(values []any) bool {
	for _, v := range values {
		if v == nil {
			return true
		}
	}
	return false
}

func asValues[T any](values []any) []T {
	out := make([]T, len(values))
	for i, v := range values {
		out[i] = v.(T)
	}
	return out
}

func buildNullableField(plan columnPlan) *data.Field {
	switch plan.kind {
	case kindInt64:
		values := make([]*int64, len(plan.values))
		for i, v := range plan.values {
			if n, ok := v.(int64); ok {
				n := n
				values[i] = &n
			}
		}
		return data.NewField(plan.name, plan.labels, values)
	case kindInt32:
		values := make([]*int32, len(plan.values))
		for i, v := range plan.values {
			if n, ok := v.(int32); ok {
				n := n
				values[i] = &n
			}
		}
		return data.NewField(plan.name, plan.labels, values)
	case kindFloat64:
		values := make([]*float64, len(plan.values))
		for i, v := range plan.values {
			if n, ok := v.(float64); ok {
				n := n
				values[i] = &n
			}
		}
		return data.NewField(plan.name, plan.labels, values)
	case kindBool:
		values := make([]*bool, len(plan.values))
		for i, v := range plan.values {
			if b, ok := v.(bool); ok {
				b := b
				values[i] = &b
			}
		}
		return data.NewField(plan.name, plan.labels, values)
	case kindTime:
		values := make([]*time.Time, len(plan.values))
		for i, v := range plan.values {
			if t, ok := v.(time.Time); ok {
				t := t
				values[i] = &t
			}
		}
		return data.NewField(plan.name, plan.labels, values)
	default:
		values := make([]*string, len(plan.values))
		for i, v := range plan.values {
			if s, ok := v.(string); ok {
				s := s
				values[i] = &s
			}
		}
		return data.NewField(plan.name, plan.labels, values)
	}
}
