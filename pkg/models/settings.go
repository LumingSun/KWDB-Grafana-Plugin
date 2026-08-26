package models

import (
	"encoding/json"
	"fmt"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

// DataSourceSettings holds the non-secret KWDB connection options stored in jsonData.
type DataSourceSettings struct {
	Host        string `json:"host"`
	Port        int    `json:"port"`
	Database    string `json:"database"`
	User        string `json:"user"`
	SSLMode     string `json:"sslMode"`
	SSLRootCert string `json:"sslRootCert"`

	// Connection pool options; durations are expressed in seconds.
	MaxConns        int `json:"maxConns"`        // default 8
	MaxConnLifetime int `json:"maxConnLifetime"` // default 3600 (1h)
	MaxConnIdleTime int `json:"maxConnIdleTime"` // default 900 (15min)
	ConnectTimeout  int `json:"connectTimeout"`  // default 5
}

// SecretDataSourceSettings holds the password from DecryptedSecureJSONData.
type SecretDataSourceSettings struct {
	Password string `json:"password"`
}

// LoadSettings reads jsonData and the decrypted password from a data source instance.
func LoadSettings(source backend.DataSourceInstanceSettings) (*DataSourceSettings, *SecretDataSourceSettings, error) {
	settings := DataSourceSettings{}
	if len(source.JSONData) > 0 {
		if err := json.Unmarshal(source.JSONData, &settings); err != nil {
			return nil, nil, fmt.Errorf("could not unmarshal DataSourceSettings json: %w", err)
		}
	}
	return &settings, &SecretDataSourceSettings{Password: source.DecryptedSecureJSONData["password"]}, nil
}
