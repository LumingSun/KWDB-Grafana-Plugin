package plugin

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/instancemgmt"
	"github.com/grafana/grafana-plugin-sdk-go/data"
	"github.com/grafana/grafana-plugin-sdk-go/experimental/concurrent"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kaiwudb/kwdb-tsdb-datasource/pkg/models"
)

var (
	_ backend.QueryDataHandler      = (*Datasource)(nil)
	_ backend.CheckHealthHandler    = (*Datasource)(nil)
	_ backend.CallResourceHandler   = (*Datasource)(nil)
	_ instancemgmt.InstanceDisposer = (*Datasource)(nil)
)

// Datasource implements the Grafana backend data source handlers.
type Datasource struct {
	pool    *pgxpool.Pool
	handler backend.CallResourceHandler
}

// QueryData handles multiple queries and returns one response per query.
func (d *Datasource) QueryData(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	return concurrent.QueryData(ctx, req, d.handleQuery, 10)
}

func (d *Datasource) handleQuery(ctx context.Context, query concurrent.Query) (res backend.DataResponse) {
	return d.query(ctx, query.DataQuery)
}

func (d *Datasource) query(ctx context.Context, query backend.DataQuery) backend.DataResponse {
	var qm models.QueryModel
	err := json.Unmarshal(query.JSON, &qm)
	if err != nil {
		return backend.ErrDataResponseWithSource(backend.StatusBadRequest, backend.ErrorSourcePlugin, fmt.Sprintf("json unmarshal: %v", err.Error()))
	}

	sql := ExpandMacros(qm.RawSql, query.TimeRange)
	if !IsReadOnly(sql) {
		return backend.ErrDataResponseWithSource(backend.StatusBadRequest, backend.ErrorSourcePlugin, "only read-only queries (SELECT/SHOW/EXPLAIN/WITH) are allowed")
	}

	rows, err := ExecuteQuery(ctx, d.pool, sql)
	if err != nil {
		backend.Logger.FromContext(ctx).Error("KWDB query failed", "refId", query.RefID, "error", err)
		return backend.DataResponse{Error: err, ErrorSource: backend.ErrorSourceDownstream}
	}
	frames, err := RowsToFrames(rows, FrameOptions{
		Format:     qm.Format,
		Mode:       qm.Mode,
		TimeColumn: qm.TimeColumn,
		Tags:       qm.Tags,
		SplitByTag: qm.SplitByTag,
		MaxRows:    DefaultMaxRows,
	})
	if err != nil {
		backend.Logger.FromContext(ctx).Error("KWDB frame conversion failed", "refId", query.RefID, "error", err)
		return backend.DataResponse{Error: err, ErrorSource: backend.ErrorSourcePlugin}
	}
	response := make(data.Frames, 0, len(frames))
	for _, frame := range frames {
		frame.RefID = query.RefID
		response = append(response, frame)
	}
	return backend.DataResponse{Frames: response}
}

// CheckHealth verifies connectivity by executing SELECT 1.
func (d *Datasource) CheckHealth(ctx context.Context, _ *backend.CheckHealthRequest) (*backend.CheckHealthResult, error) {
	if d.pool == nil {
		return &backend.CheckHealthResult{Status: backend.HealthStatusError, Message: "database connection pool is not initialized"}, nil
	}
	rows, err := d.pool.Query(ctx, "SELECT 1")
	if err != nil {
		return &backend.CheckHealthResult{Status: backend.HealthStatusError, Message: err.Error()}, nil
	}
	defer rows.Close()
	for rows.Next() {
		var one int32
		if err := rows.Scan(&one); err != nil {
			return &backend.CheckHealthResult{Status: backend.HealthStatusError, Message: err.Error()}, nil
		}
	}
	if err := rows.Err(); err != nil {
		return &backend.CheckHealthResult{Status: backend.HealthStatusError, Message: err.Error()}, nil
	}
	return &backend.CheckHealthResult{
		Status:  backend.HealthStatusOk,
		Message: "Data source is working",
	}, nil
}

// CallResource delegates to the metadata resource handler.
func (d *Datasource) CallResource(ctx context.Context, req *backend.CallResourceRequest, sender backend.CallResourceResponseSender) error {
	if d.handler == nil {
		return fmt.Errorf("resource handler is not initialized")
	}
	return d.handler.CallResource(ctx, req, sender)
}

// Dispose closes the connection pool when the instance is replaced.
func (d *Datasource) Dispose() {
	if d.pool != nil {
		d.pool.Close()
	}
}
