package plugin

import (
	"strings"
	"testing"

	"github.com/kaiwudb/kwdb-tsdb-datasource/pkg/models"
)

func TestApplyDefaults(t *testing.T) {
	cfg := models.DataSourceSettings{}
	applyDefaults(&cfg)

	if cfg.Host != defaultHost {
		t.Errorf("Host = %q, want %q", cfg.Host, defaultHost)
	}
	if cfg.Port != defaultPort {
		t.Errorf("Port = %d, want %d", cfg.Port, defaultPort)
	}
	if cfg.Database != defaultDatabase {
		t.Errorf("Database = %q, want %q", cfg.Database, defaultDatabase)
	}
	if cfg.User != defaultUser {
		t.Errorf("User = %q, want %q", cfg.User, defaultUser)
	}
	if cfg.SSLMode != defaultSSLMode {
		t.Errorf("SSLMode = %q, want %q", cfg.SSLMode, defaultSSLMode)
	}
}

func TestApplyDefaultsKeepsProvidedValues(t *testing.T) {
	cfg := models.DataSourceSettings{
		Host:     "10.0.0.1",
		Port:     26258,
		Database: "ts_db",
		User:     "reader",
		SSLMode:  "require",
	}
	applyDefaults(&cfg)

	if cfg.Host != "10.0.0.1" || cfg.Port != 26258 || cfg.Database != "ts_db" ||
		cfg.User != "reader" || cfg.SSLMode != "require" {
		t.Errorf("applyDefaults overwrote provided values: %#v", cfg)
	}
}

func TestBuildConnString(t *testing.T) {
	cfg := models.DataSourceSettings{
		Host:     "10.110.105.80",
		Port:     26258,
		Database: "ts_db",
		User:     "root",
		SSLMode:  "disable",
	}
	dsn := buildConnString(cfg, "secret")

	for _, want := range []string{"postgresql://root:secret@10.110.105.80:26258/ts_db", "sslmode=disable"} {
		if !strings.Contains(dsn, want) {
			t.Errorf("DSN %q does not contain %q", dsn, want)
		}
	}
}
