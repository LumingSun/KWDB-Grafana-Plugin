package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/data"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestRemoteKWDBIntegration(t *testing.T) {
	dsn := os.Getenv("KWDB_TEST_DSN")
	if dsn == "" {
		t.Skip("set KWDB_TEST_DSN to run the remote KWDB integration test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	if err := pool.Ping(ctx); err != nil {
		t.Fatalf("ping: %v", err)
	}
	t.Log("PING OK")

	table := fmt.Sprintf("grafana_probe_%d", time.Now().Unix())
	dropSQL := fmt.Sprintf("DROP TABLE %s.%s", quoteIdent("ts_db"), quoteIdent(table))
	createSQL := `CREATE TABLE ` + quoteIdent("ts_db") + `.` + quoteIdent(table) + ` (
		ts TIMESTAMP NOT NULL,
		temperature DOUBLE,
		humidity DOUBLE
	) TAGS (
		device_id INT NOT NULL,
		location VARCHAR(32)
	) PRIMARY TAGS (device_id)`
	if _, err := pool.Exec(ctx, createSQL); err != nil {
		t.Fatalf("create temp TS table: %v", err)
	}
	defer func() {
		_, _ = pool.Exec(ctx, dropSQL)
	}()

	base := time.Now().UTC().Truncate(time.Second)
	inserts := [][]string{
		{base.Add(-10 * time.Minute).Format(macroTimeFormat), "21.4", "60.1", "1", "room-a"},
		{base.Add(-10 * time.Minute).Format(macroTimeFormat), "22.0", "59.8", "2", "room-b"},
		{base.Add(-5 * time.Minute).Format(macroTimeFormat), "21.8", "60.0", "1", "room-a"},
		{base.Add(-5 * time.Minute).Format(macroTimeFormat), "22.4", "NULL", "2", "room-b"},
		{base.Add(-2 * time.Minute).Format(macroTimeFormat), "22.1", "59.9", "1", "room-a"},
		{base.Add(-2 * time.Minute).Format(macroTimeFormat), "22.7", "59.6", "2", "room-b"},
		{base.Add(-1 * time.Minute).Format(macroTimeFormat), "22.3", "59.7", "1", "room-a"},
		{base.Add(-1 * time.Minute).Format(macroTimeFormat), "23.1", "59.4", "2", "room-b"},
	}
	for _, row := range inserts {
		values := fmt.Sprintf("'%s', %s, %s, %s, '%s'", row[0], row[1], row[2], row[3], row[4])
		if _, err := pool.Exec(ctx, fmt.Sprintf(
			"INSERT INTO %s.%s (ts, temperature, humidity, device_id, location) VALUES (%s)",
			quoteIdent("ts_db"), quoteIdent(table), values,
		)); err != nil {
			t.Fatalf("insert into temp TS table: %v", err)
		}
	}

	verifyMetadataResources(t, pool, table)
	verifyTableFrame(t, ctx, pool, table)
	verifyTimeSeriesFrame(t, ctx, pool, table)
	verifyTimeBucket(t, ctx, pool, table)
	verifyGapfill(t, ctx, pool, table)
	verifyLatest(t, ctx, pool, table)
}

func verifyMetadataResources(t *testing.T, pool *pgxpool.Pool, table string) {
	t.Helper()
	handler := newMetadataHandler(pool, "ts_db")

	tablesResp := callResource(t, handler, "/tables")
	if tablesResp.Status != 200 {
		t.Fatalf("GET /tables status = %d, body = %s", tablesResp.Status, tablesResp.Body)
	}
	var tables []string
	if err := json.Unmarshal(tablesResp.Body, &tables); err != nil {
		t.Fatal(err)
	}
	if !containsString(tables, table) {
		t.Fatalf("temp table %s not listed: %v", table, tables)
	}
	t.Logf("TABLES contain %s", table)

	colsResp := callResource(t, handler, "/columns?table="+table)
	if colsResp.Status != 200 {
		t.Fatalf("GET /columns status = %d, body = %s", colsResp.Status, colsResp.Body)
	}
	var cols []ColumnInfo
	if err := json.Unmarshal(colsResp.Body, &cols); err != nil {
		t.Fatal(err)
	}
	got := map[string]ColumnInfo{}
	for _, c := range cols {
		got[c.Name] = c
	}
	if !got["ts"].IsTimeColumn {
		t.Errorf("ts not marked as time column: %#v", got["ts"])
	}
	if !got["device_id"].IsTag || !got["device_id"].IsPrimaryTag {
		t.Errorf("device_id tag metadata wrong: %#v", got["device_id"])
	}
	if !got["location"].IsTag || got["location"].IsPrimaryTag {
		t.Errorf("location tag metadata wrong: %#v", got["location"])
	}
	if got["temperature"].IsTag {
		t.Errorf("temperature should not be a tag: %#v", got["temperature"])
	}
	t.Logf("COLUMNS: %v", cols)
}

func verifyTableFrame(t *testing.T, ctx context.Context, pool *pgxpool.Pool, table string) {
	t.Helper()
	rows, err := pool.Query(ctx, fmt.Sprintf(
		"SELECT * FROM %s.%s ORDER BY ts",
		quoteIdent("ts_db"), quoteIdent(table),
	))
	if err != nil {
		t.Fatal(err)
	}
	frame, err := RowsToFrame(rows, "table", "", nil)
	if err != nil {
		t.Fatal(err)
	}
	rowLen, _ := frame.RowLen()
	if rowLen != 8 {
		t.Fatalf("table frame row count = %d, want 8", rowLen)
	}
	byName := map[string]*data.Field{}
	for _, f := range frame.Fields {
		byName[f.Name] = f
	}
	if byName["humidity"].Type() != data.FieldTypeNullableFloat64 {
		t.Errorf("humidity type = %v, want nullable float64", byName["humidity"].Type())
	}
	logFrameSummary(t, "TABLE FORMAT", frame)
}

func verifyTimeSeriesFrame(t *testing.T, ctx context.Context, pool *pgxpool.Pool, table string) {
	t.Helper()
	rows, err := pool.Query(ctx, fmt.Sprintf(
		"SELECT * FROM %s.%s ORDER BY ts",
		quoteIdent("ts_db"), quoteIdent(table),
	))
	if err != nil {
		t.Fatal(err)
	}
	frame, err := RowsToFrame(rows, "time_series", "ts", []string{"device_id", "location"})
	if err != nil {
		t.Fatal(err)
	}
	if frame.Fields[0].Name != "time" {
		t.Fatalf("time series first field = %q, want time", frame.Fields[0].Name)
	}
	labels := map[string]string{}
	for _, f := range frame.Fields[1:] {
		labels[f.Labels["device_id"]+":"+f.Labels["location"]] = f.Name
	}
	if len(labels) != 2 {
		t.Fatalf("expected 2 series, got %#v", labels)
	}
	logFrameSummary(t, "TIME_SERIES FORMAT", frame)
}

func verifyTimeBucket(t *testing.T, ctx context.Context, pool *pgxpool.Pool, table string) {
	t.Helper()
	sql := fmt.Sprintf(
		"SELECT time_bucket(ts, '5m') AS time, avg(temperature) AS value FROM %s.%s GROUP BY time ORDER BY time",
		quoteIdent("ts_db"), quoteIdent(table),
	)
	rows, err := pool.Query(ctx, sql)
	if err != nil {
		t.Fatal(err)
	}
	frame, err := RowsToFrame(rows, "time_series", "time", nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range frame.Fields {
		if f.Name == "value" && f.Type() != data.FieldTypeFloat64 && f.Type() != data.FieldTypeNullableFloat64 {
			t.Errorf("time_bucket value type = %v, want float64", f.Type())
		}
	}
	logFrameSummary(t, "TIME_BUCKET", frame)
}

func verifyGapfill(t *testing.T, ctx context.Context, pool *pgxpool.Pool, table string) {
	t.Helper()
	sql := fmt.Sprintf(
		"SELECT time_bucket_gapfill(ts, '5m') AS time, interpolate(avg(temperature), 'linear') AS value FROM %s.%s GROUP BY time ORDER BY time",
		quoteIdent("ts_db"), quoteIdent(table),
	)
	rows, err := pool.Query(ctx, sql)
	if err != nil {
		t.Fatal(err)
	}
	frame, err := RowsToFrame(rows, "time_series", "time", nil)
	if err != nil {
		t.Fatal(err)
	}
	logFrameSummary(t, "GAPFILL", frame)
}

func verifyLatest(t *testing.T, ctx context.Context, pool *pgxpool.Pool, table string) {
	t.Helper()
	sql := fmt.Sprintf(
		"SELECT device_id, last(temperature) AS value, last(ts) AS time FROM %s.%s GROUP BY device_id ORDER BY device_id",
		quoteIdent("ts_db"), quoteIdent(table),
	)
	rows, err := pool.Query(ctx, sql)
	if err != nil {
		t.Fatal(err)
	}
	frame, err := RowsToFrame(rows, "time_series", "time", []string{"device_id"})
	if err != nil {
		t.Fatal(err)
	}
	rowLen, _ := frame.RowLen()
	if rowLen < 1 {
		t.Fatalf("latest frame row count = %d, want at least 1", rowLen)
	}
	byLabel := map[string]*data.Field{}
	for _, f := range frame.Fields[1:] {
		byLabel[f.Labels["device_id"]] = f
	}
	if len(byLabel) != 2 {
		t.Fatalf("latest frame has %d series, want 2", len(byLabel))
	}
	logFrameSummary(t, "LATEST", frame)
}

func containsString(items []string, target string) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
}

func logFrameSummary(t *testing.T, title string, frame *data.Frame) {
	t.Helper()
	t.Logf("--- %s (rows=%d) ---", title, frame.Rows())
	for _, f := range frame.Fields {
		values := make([]string, 0, f.Len())
		for i := 0; i < f.Len() && i < 3; i++ {
			values = append(values, fmt.Sprint(f.At(i)))
		}
		t.Logf("FIELD name=%q type=%v labels=%v first=[%s]", f.Name, f.Type(), f.Labels, strings.Join(values, ", "))
	}
}
