package plugin

import (
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
)

// latestValueRows returns the exact result shape of a Latest Values query
// grouped by device_id: one row per device sharing an identical timestamp.
func latestValueRows() (*mockRows, time.Time) {
	t0 := time.Date(2026, 8, 24, 9, 57, 13, 474_000_000, time.UTC)
	fields := []pgconn.FieldDescription{
		{Name: "device_id", DataTypeOID: 25},
		{Name: "电压", DataTypeOID: 701},
		{Name: "转速", DataTypeOID: 20},
		{Name: "温度", DataTypeOID: 701},
		{Name: "time", DataTypeOID: 1114},
	}
	rows := [][]any{
		{"CNC-002", 222.86, int64(1536), 44.69, t0},
		{"INJ-001", 222.49, int64(1532), 44.7, t0},
		{"CNC-001", 222.88, int64(1536), 44.77, t0},
		{"INJ-002", 222.24, int64(1538), 44.7, t0},
	}
	return &mockRows{fields: fields, rows: rows}, t0
}

func TestRowsToFramesLatestTableSplitsPerDevice(t *testing.T) {
	rows, _ := latestValueRows()

	frames, err := RowsToFrames(rows, FrameOptions{Format: "table", Mode: "latest", Tags: []string{"device_id"}, MaxRows: DefaultMaxRows})
	if err != nil {
		t.Fatal(err)
	}

	wantNames := []string{"CNC-001", "CNC-002", "INJ-001", "INJ-002"}
	if len(frames) != len(wantNames) {
		t.Fatalf("frame count = %d, want %d", len(frames), len(wantNames))
	}

	voltages := map[string]float64{"CNC-001": 222.88, "CNC-002": 222.86, "INJ-001": 222.49, "INJ-002": 222.24}
	for i, frame := range frames {
		if got := frame.Name; got != wantNames[i] {
			t.Errorf("frame %d name = %q, want %q", i, got, wantNames[i])
		}
		if len(frame.Fields) != 5 {
			t.Fatalf("frame %d has %d fields, want 5 (tag columns are kept)", i, len(frame.Fields))
		}

		deviceField := frame.Fields[0]
		if got := deviceField.At(0); got != wantNames[i] {
			t.Errorf("frame %d device_id = %v, want %q", i, got, wantNames[i])
		}
		if deviceField.Labels != nil {
			t.Errorf("frame %d device_id should not carry labels (frame name carries identity), got %v", i, deviceField.Labels)
		}

		voltage := frame.Fields[1]
		if got := voltage.At(0).(float64); got != voltages[wantNames[i]] {
			t.Errorf("frame %d 电压 = %v, want %v", i, got, voltages[wantNames[i]])
		}
		if voltage.Labels != nil {
			t.Errorf("frame %d 电压 should not carry labels (frame name carries identity), got %v", i, voltage.Labels)
		}

		timeField := frame.Fields[4]
		if timeField.Labels != nil {
			t.Errorf("frame %d time field should not carry labels, got %v", i, timeField.Labels)
		}
	}
}

func TestRowsToFramesLatestTableWithoutTagsStaysSingleFrame(t *testing.T) {
	rows, _ := latestValueRows()
	// Drop the tag column to emulate a latest query without GROUP BY.
	rows = &mockRows{fields: rows.fields[1:], rows: [][]any{
		{222.86, int64(1536), 44.69, time.Date(2026, 8, 24, 9, 57, 13, 474_000_000, time.UTC)},
		{222.49, int64(1532), 44.7, time.Date(2026, 8, 24, 9, 57, 13, 474_000_000, time.UTC)},
	}}

	frames, err := RowsToFrames(rows, FrameOptions{Format: "table", Mode: "latest", MaxRows: DefaultMaxRows})
	if err != nil {
		t.Fatal(err)
	}
	if len(frames) != 1 {
		t.Fatalf("frame count = %d, want 1", len(frames))
	}
	if rowLen, err := frames[0].RowLen(); err != nil || rowLen != 2 {
		t.Fatalf("RowLen = %d, %v; want 2", rowLen, err)
	}
}

func TestRowsToFramesLatestTableSplitDisabledStaysSingleFrame(t *testing.T) {
	rows, _ := latestValueRows()
	split := false

	frames, err := RowsToFrames(rows, FrameOptions{Format: "table", Mode: "latest", Tags: []string{"device_id"}, SplitByTag: &split, MaxRows: DefaultMaxRows})
	if err != nil {
		t.Fatal(err)
	}
	if len(frames) != 1 {
		t.Fatalf("frame count = %d, want 1 when splitting is disabled", len(frames))
	}
	if rowLen, err := frames[0].RowLen(); err != nil || rowLen != 4 {
		t.Fatalf("RowLen = %d, %v; want 4", rowLen, err)
	}
	if frames[0].Fields[1].Labels != nil {
		t.Fatalf("disabled split should not label fields, got %v", frames[0].Fields[1].Labels)
	}
}

func TestRowsToFramesNonLatestModesStaySingleFrame(t *testing.T) {
	rows, _ := latestValueRows()

	frames, err := RowsToFrames(rows, FrameOptions{Format: "table", Mode: "downsampling", Tags: []string{"device_id"}, MaxRows: DefaultMaxRows})
	if err != nil {
		t.Fatal(err)
	}
	if len(frames) != 1 {
		t.Fatalf("frame count = %d, want 1 for non-latest mode", len(frames))
	}
	if rowLen, err := frames[0].RowLen(); err != nil || rowLen != 4 {
		t.Fatalf("RowLen = %d, %v; want 4", rowLen, err)
	}
}

func TestRowsToFramesLatestTableMissingTagColumnFallsBack(t *testing.T) {
	rows, _ := latestValueRows()

	frames, err := RowsToFrames(rows, FrameOptions{Format: "table", Mode: "latest", Tags: []string{"location"}, MaxRows: DefaultMaxRows})
	if err != nil {
		t.Fatal(err)
	}
	if len(frames) != 1 {
		t.Fatalf("frame count = %d, want 1 when the tag column is absent", len(frames))
	}
}

func TestRowsToFrameLatestTimeSeriesLabelsEveryDevice(t *testing.T) {
	rows, _ := latestValueRows()

	frame, err := RowsToFrame(rows, "time_series", "", []string{"device_id"})
	if err != nil {
		t.Fatal(err)
	}

	// time + 3 metrics x 4 devices
	if len(frame.Fields) != 13 {
		t.Fatalf("field count = %d, want 13", len(frame.Fields))
	}
	labeled := map[string]int{}
	for _, field := range frame.Fields[1:] {
		labeled[field.Name+"|"+field.Labels["device_id"]]++
		if field.Len() != 1 {
			t.Errorf("field %v has %d values, want 1", field.Name, field.Len())
		}
	}
	if len(labeled) != 12 {
		t.Fatalf("got %d labeled series, want 12", len(labeled))
	}
	if _, ok := labeled["电压|CNC-001"]; !ok {
		t.Fatalf("missing labeled series 电压/CNC-001: %v", labeled)
	}
}
