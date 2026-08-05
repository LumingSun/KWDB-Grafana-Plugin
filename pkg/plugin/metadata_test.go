package plugin

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
)

func TestParseDDL(t *testing.T) {
	ddl := `CREATE TABLE ts_db.sensors (
    ts TIMESTAMP NOT NULL,
    temperature DOUBLE,
    humidity DOUBLE
) TAGS (
    device_id INT NOT NULL,
    location VARCHAR(100)
) PRIMARY TAGS (device_id);`

	infos := parseDDL(ddl)
	if len(infos) != 5 {
		t.Fatalf("got %d columns, want 5: %#v", len(infos), infos)
	}
	if infos[0].Name != "ts" || !infos[0].IsTimeColumn || infos[0].IsTag {
		t.Errorf("unexpected ts column: %#v", infos[0])
	}
	if infos[1].Name != "temperature" || infos[1].Type != "DOUBLE" {
		t.Errorf("unexpected temperature column: %#v", infos[1])
	}
	deviceID := infos[3]
	if deviceID.Name != "device_id" || !deviceID.IsTag || !deviceID.IsPrimaryTag || deviceID.Type != "INT" {
		t.Errorf("unexpected device_id column: %#v", deviceID)
	}
	if !infos[4].IsTag || infos[4].Name != "location" || infos[4].Type != "VARCHAR(100)" {
		t.Errorf("unexpected location column: %#v", infos[4])
	}
}

func TestParseDDLRelationalTable(t *testing.T) {
	ddl := `CREATE TABLE devices (id INT PRIMARY KEY, name VARCHAR(100));`
	infos := parseDDL(ddl)
	if len(infos) != 2 {
		t.Fatalf("got %d columns, want 2", len(infos))
	}
	for _, info := range infos {
		if info.IsTag || info.IsPrimaryTag {
			t.Errorf("relational column should not be a tag: %#v", info)
		}
	}
}

func TestParseDDLComplexKWDB(t *testing.T) {
	ddl := `CREATE TABLE device_info (
    create_time TIMESTAMPTZ NOT NULL,
    device_id INT COMMENT 'device ID' NOT NULL,
    install_date TIMESTAMPTZ,
    warranty_period INT2
) TAGS (
    plant_code INT2 NOT NULL COMMENT = 'plant code',
    workshop VARCHAR(128) NOT NULL,
    device_type CHAR(1023) NOT NULL,
    manufacturer NCHAR(254) NOT NULL
) PRIMARY TAGS(plant_code, workshop, device_type, manufacturer) COMMENT = 'table for device information';`

	infos := parseDDL(ddl)
	if len(infos) != 8 {
		t.Fatalf("got %d columns, want 8: %#v", len(infos), infos)
	}
	if !infos[0].IsTimeColumn || infos[0].Type != "TIMESTAMPTZ" {
		t.Errorf("unexpected create_time column: %#v", infos[0])
	}
	if infos[1].Type != "INT" {
		t.Errorf("unexpected device_id type: %#v", infos[1])
	}
	primary := map[string]bool{}
	for _, info := range infos {
		if info.IsPrimaryTag {
			primary[info.Name] = true
		}
	}
	for _, name := range []string{"plant_code", "workshop", "device_type", "manufacturer"} {
		if !primary[name] {
			t.Errorf("expected primary tag %s, got %#v", name, infos)
		}
	}
}

func TestParseDDLFailureReturnsNil(t *testing.T) {
	if infos := parseDDL("not a create table statement"); infos != nil {
		t.Fatalf("expected nil, got %#v", infos)
	}
}

func TestTablesResource(t *testing.T) {
	handler := newMetadataHandler(&fakeQuerier{
		fn: func(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
			return &mockRows{rows: [][]any{{"sensors", "TIME SERIES TABLE"}, {"devices", "TABLE"}}}, nil
		},
	}, "defaultdb")

	resp := callResource(t, handler, "/tables")
	if resp.Status != 200 {
		t.Fatalf("status = %d, body = %s", resp.Status, resp.Body)
	}
	var tables []TableInfo
	if err := json.Unmarshal(resp.Body, &tables); err != nil {
		t.Fatal(err)
	}
	if len(tables) != 2 || tables[0].Name != "sensors" || tables[0].Type != "TIME SERIES TABLE" ||
		tables[1].Name != "devices" || tables[1].Type != "TABLE" {
		t.Errorf("unexpected tables: %#v", tables)
	}
}

func TestTablesResourceQueryError(t *testing.T) {
	handler := newMetadataHandler(&fakeQuerier{
		fn: func(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
			return nil, errors.New("connection refused")
		},
	}, "defaultdb")

	resp := callResource(t, handler, "/tables")
	if resp.Status != 500 {
		t.Fatalf("status = %d, want 500", resp.Status)
	}
}

func TestColumnsResource(t *testing.T) {
	ddl := `CREATE TABLE iot_db.sensor_data (
    ts TIMESTAMPTZ NOT NULL,
    temperature DOUBLE,
    device_id INT4
) TAGS (
    device_id INT4 NOT NULL,
    location VARCHAR(100)
) PRIMARY TAGS (device_id)`

	handler := newMetadataHandler(&fakeQuerier{
		fn: func(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
			return &mockRows{rows: [][]any{{"sensor_data", ddl}}}, nil
		},
	}, "iot_db")

	resp := callResource(t, handler, "/columns?table=sensor_data")
	if resp.Status != 200 {
		t.Fatalf("status = %d, body = %s", resp.Status, resp.Body)
	}
	var cols []ColumnInfo
	if err := json.Unmarshal(resp.Body, &cols); err != nil {
		t.Fatal(err)
	}
	if len(cols) != 4 {
		t.Fatalf("got %d columns, want 4: %#v", len(cols), cols)
	}
	if !cols[0].IsTimeColumn {
		t.Errorf("expected ts time column: %#v", cols[0])
	}
	if !cols[2].IsTag || !cols[2].IsPrimaryTag || cols[2].Name != "device_id" {
		t.Errorf("unexpected device_id column: %#v", cols[2])
	}
	if !cols[3].IsTag || cols[3].Name != "location" {
		t.Errorf("unexpected location column: %#v", cols[3])
	}
}

func TestColumnsResourceMissingTable(t *testing.T) {
	handler := newMetadataHandler(&fakeQuerier{}, "defaultdb")
	resp := callResource(t, handler, "/columns")
	if resp.Status != 400 {
		t.Fatalf("status = %d, want 400", resp.Status)
	}
}

func TestColumnsResourceQueryError(t *testing.T) {
	handler := newMetadataHandler(&fakeQuerier{
		fn: func(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
			return nil, errors.New("table does not exist")
		},
	}, "defaultdb")

	resp := callResource(t, handler, "/columns?table=missing")
	if resp.Status != 500 {
		t.Fatalf("status = %d, want 500", resp.Status)
	}
}

func TestColumnsResourceUnparseableDDL(t *testing.T) {
	handler := newMetadataHandler(&fakeQuerier{
		fn: func(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
			return &mockRows{rows: [][]any{{"weird", "not a create table"}}}, nil
		},
	}, "defaultdb")

	resp := callResource(t, handler, "/columns?table=weird")
	if resp.Status != 200 {
		t.Fatalf("status = %d, want 200", resp.Status)
	}
	if string(resp.Body) != "[]\n" {
		t.Fatalf("body = %q, want []", resp.Body)
	}
}
