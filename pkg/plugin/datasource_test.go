package plugin

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

func TestQueryDataRejectsNonReadOnly(t *testing.T) {
	ds := Datasource{}
	resp, err := ds.QueryData(
		context.Background(),
		&backend.QueryDataRequest{
			Queries: []backend.DataQuery{
				{RefID: "A", JSON: json.RawMessage(`{"rawSql":"DELETE FROM sensors"}`)},
			},
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	dr := resp.Responses["A"]
	if dr.Error == nil {
		t.Fatal("expected read-only rejection error")
	}
	if dr.Status != backend.StatusBadRequest {
		t.Fatalf("status = %v, want bad request", dr.Status)
	}
}

func TestQueryDataInvalidJSON(t *testing.T) {
	ds := Datasource{}
	resp, err := ds.QueryData(
		context.Background(),
		&backend.QueryDataRequest{
			Queries: []backend.DataQuery{
				{RefID: "A", JSON: json.RawMessage(`not json`)},
			},
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	if resp.Responses["A"].Error == nil {
		t.Fatal("expected json unmarshal error")
	}
}

func TestQueryDataNilPool(t *testing.T) {
	ds := Datasource{}
	resp, err := ds.QueryData(
		context.Background(),
		&backend.QueryDataRequest{
			Queries: []backend.DataQuery{
				{RefID: "A", JSON: json.RawMessage(`{"rawSql":"SELECT 1"}`)},
			},
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	if resp.Responses["A"].Error == nil {
		t.Fatal("expected pool error")
	}
}

func TestCheckHealthNilPool(t *testing.T) {
	ds := Datasource{}
	result, err := ds.CheckHealth(context.Background(), &backend.CheckHealthRequest{})
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != backend.HealthStatusError {
		t.Fatalf("status = %v, want error", result.Status)
	}
}

func TestResolveFormatDefaultsToTimeSeriesForBuilderModes(t *testing.T) {
	cases := []struct {
		mode   string
		format string
		want   string
	}{
		{"downsampling", "", "time_series"},
		{"gapfill", "", "time_series"},
		{"latest", "", "time_series"},
		{"window", "", "time_series"},
		{"downsampling", "table", "table"},
		{"latest", "time_series", "time_series"},
		{"raw", "", ""},
		{"raw", "table", "table"},
		{"raw", "time_series", "time_series"},
	}
	for _, tc := range cases {
		if got := resolveFormat(tc.mode, tc.format); got != tc.want {
			t.Errorf("resolveFormat(%q, %q) = %q, want %q", tc.mode, tc.format, got, tc.want)
		}
	}
}

func TestCallResourceNilHandler(t *testing.T) {
	ds := Datasource{}
	err := ds.CallResource(context.Background(), &backend.CallResourceRequest{}, backend.CallResourceResponseSenderFunc(func(*backend.CallResourceResponse) error {
		return nil
	}))
	if err == nil {
		t.Fatal("expected error for nil resource handler")
	}
}
