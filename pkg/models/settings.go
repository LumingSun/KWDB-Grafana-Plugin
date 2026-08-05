package models

import (
	"encoding/json"
	"fmt"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

// DataSourceSettings holds the non-secret KWDB connection options stored in jsonData.
type DataSourceSettings struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Database string `json:"database"`
	User     string `json:"user"`
	SSLMode  string `json:"sslMode"`
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
