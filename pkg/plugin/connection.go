package plugin

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"time"

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

	// Connection pool defaults; durations are expressed in seconds.
	defaultPoolMaxConns        = 8
	poolMinConns               = 0
	defaultPoolMaxConnLifetime = 3600 // 1h
	defaultPoolMaxConnIdleTime = 900  // 15min
	defaultPoolConnectTimeout  = 5
)

// NewDatasource creates a datasource instance with a KWDB pgx connection pool.
func NewDatasource(ctx context.Context, settings backend.DataSourceInstanceSettings) (instancemgmt.Instance, error) {
	cfg, secrets, err := models.LoadSettings(settings)
	if err != nil {
		return nil, err
	}
	applyDefaults(cfg)

	poolCfg, err := pgxpool.ParseConfig(buildConnString(*cfg, secrets.Password))
	if err != nil {
		return nil, fmt.Errorf("could not parse pgx connection string: %w", err)
	}
	poolCfg.MaxConns = int32(cfg.MaxConns)
	poolCfg.MinConns = poolMinConns
	poolCfg.MaxConnLifetime = time.Duration(cfg.MaxConnLifetime) * time.Second
	poolCfg.MaxConnIdleTime = time.Duration(cfg.MaxConnIdleTime) * time.Second
	poolCfg.ConnConfig.ConnectTimeout = time.Duration(cfg.ConnectTimeout) * time.Second

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
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
	if cfg.MaxConns <= 0 {
		cfg.MaxConns = defaultPoolMaxConns
	}
	if cfg.MaxConnLifetime <= 0 {
		cfg.MaxConnLifetime = defaultPoolMaxConnLifetime
	}
	if cfg.MaxConnIdleTime <= 0 {
		cfg.MaxConnIdleTime = defaultPoolMaxConnIdleTime
	}
	if cfg.ConnectTimeout <= 0 {
		cfg.ConnectTimeout = defaultPoolConnectTimeout
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
	if cfg.SSLRootCert != "" {
		query.Set("sslrootcert", cfg.SSLRootCert)
	}
	u.RawQuery = query.Encode()
	return u.String()
}
