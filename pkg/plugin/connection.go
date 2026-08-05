package plugin

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"strconv"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/instancemgmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kaiwudb/kwdb-tsdb-datasource/pkg/models"
)

const (
	defaultHost     = "localhost"
	defaultPort     = 26257
	defaultDatabase = "defaultdb"
	defaultUser     = "root"
	defaultSSLMode  = "disable"
)

// NewDatasource creates a datasource instance with a KWDB pgx connection pool.
func NewDatasource(ctx context.Context, settings backend.DataSourceInstanceSettings) (instancemgmt.Instance, error) {
	cfg, secrets, err := models.LoadSettings(settings)
	if err != nil {
		return nil, err
	}
	applyDefaults(cfg)

	pool, err := pgxpool.New(ctx, buildConnString(*cfg, secrets.Password))
	if err != nil {
		return nil, fmt.Errorf("could not create pgx connection pool: %w", err)
	}

	return &Datasource{
		pool:    pool,
		handler: newMetadataHandler(pool, cfg.Database),
	}, nil
}

func applyDefaults(cfg *models.DataSourceSettings) {
	if cfg.Host == "" {
		cfg.Host = defaultHost
	}
	if cfg.Port == 0 {
		cfg.Port = defaultPort
	}
	if cfg.Database == "" {
		cfg.Database = defaultDatabase
	}
	if cfg.User == "" {
		cfg.User = defaultUser
	}
	if cfg.SSLMode == "" {
		cfg.SSLMode = defaultSSLMode
	}
}

func buildConnString(cfg models.DataSourceSettings, password string) string {
	u := url.URL{
		Scheme: "postgresql",
		User:   url.UserPassword(cfg.User, password),
		Host:   net.JoinHostPort(cfg.Host, strconv.Itoa(cfg.Port)),
		Path:   "/" + cfg.Database,
	}
	query := u.Query()
	query.Set("sslmode", cfg.SSLMode)
	u.RawQuery = query.Encode()
	return u.String()
}
