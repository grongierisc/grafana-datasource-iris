package models

import (
	"encoding/json"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

const (
	DefaultHost                   = "localhost"
	DefaultPort                   = 1972
	DefaultNamespace              = "USER"
	DefaultQueryTimeoutSeconds    = 30
	DefaultRowLimit               = 1000
	DefaultMaxOpenConnections     = 10
	DefaultMaxIdleConnections     = 5
	DefaultConnMaxLifetimeSeconds = 1800
)

type PluginSettings struct {
	Host                   string                `json:"host"`
	Port                   int                   `json:"port"`
	Namespace              string                `json:"namespace"`
	Username               string                `json:"username"`
	QueryTimeoutSeconds    int                   `json:"queryTimeoutSeconds"`
	RowLimit               int                   `json:"rowLimit"`
	MaxOpenConnections     int                   `json:"maxOpenConns"`
	MaxIdleConnections     int                   `json:"maxIdleConns"`
	ConnMaxLifetimeSeconds int                   `json:"connMaxLifetimeSeconds"`
	Secrets                *SecretPluginSettings `json:"-"`
}

type SecretPluginSettings struct {
	Password string `json:"password"`
}

func LoadPluginSettings(source backend.DataSourceInstanceSettings) (*PluginSettings, error) {
	settings := PluginSettings{
		Host:                   DefaultHost,
		Port:                   DefaultPort,
		Namespace:              DefaultNamespace,
		QueryTimeoutSeconds:    DefaultQueryTimeoutSeconds,
		RowLimit:               DefaultRowLimit,
		MaxOpenConnections:     DefaultMaxOpenConnections,
		MaxIdleConnections:     DefaultMaxIdleConnections,
		ConnMaxLifetimeSeconds: DefaultConnMaxLifetimeSeconds,
	}
	if len(source.JSONData) > 0 {
		err := json.Unmarshal(source.JSONData, &settings)
		if err != nil {
			return nil, fmt.Errorf("could not unmarshal PluginSettings json: %w", err)
		}
	}

	settings.Secrets = loadSecretPluginSettings(source.DecryptedSecureJSONData)
	settings.applyDefaults()

	return &settings, nil
}

func loadSecretPluginSettings(source map[string]string) *SecretPluginSettings {
	return &SecretPluginSettings{
		Password: source["password"],
	}
}

func (s *PluginSettings) applyDefaults() {
	if s.Host == "" {
		s.Host = DefaultHost
	}
	if s.Port <= 0 {
		s.Port = DefaultPort
	}
	if s.Namespace == "" {
		s.Namespace = DefaultNamespace
	}
	if s.QueryTimeoutSeconds <= 0 {
		s.QueryTimeoutSeconds = DefaultQueryTimeoutSeconds
	}
	if s.RowLimit <= 0 {
		s.RowLimit = DefaultRowLimit
	}
	if s.MaxOpenConnections <= 0 {
		s.MaxOpenConnections = DefaultMaxOpenConnections
	}
	if s.MaxIdleConnections <= 0 {
		s.MaxIdleConnections = DefaultMaxIdleConnections
	}
	if s.ConnMaxLifetimeSeconds <= 0 {
		s.ConnMaxLifetimeSeconds = DefaultConnMaxLifetimeSeconds
	}
	if s.Secrets == nil {
		s.Secrets = &SecretPluginSettings{}
	}
}

func (s *PluginSettings) Validate() error {
	if s.Host == "" {
		return fmt.Errorf("host is missing")
	}
	if s.Port <= 0 {
		return fmt.Errorf("port must be greater than zero")
	}
	if s.Namespace == "" {
		return fmt.Errorf("namespace is missing")
	}
	if s.Username == "" {
		return fmt.Errorf("username is missing")
	}
	if s.Secrets == nil || s.Secrets.Password == "" {
		return fmt.Errorf("password is missing")
	}
	return nil
}

func (s *PluginSettings) ConnectionLifetime() time.Duration {
	return time.Duration(s.ConnMaxLifetimeSeconds) * time.Second
}

func (s *PluginSettings) DSN() string {
	u := url.URL{
		Scheme: "iris",
		User:   url.UserPassword(s.Username, s.Secrets.Password),
		Host:   net.JoinHostPort(s.Host, strconv.Itoa(s.Port)),
		Path:   "/" + s.Namespace,
	}

	query := url.Values{}
	query.Set("max_rows", strconv.Itoa(s.RowLimit))
	query.Set("query_timeout", strconv.Itoa(s.QueryTimeoutSeconds))
	u.RawQuery = query.Encode()

	return u.String()
}
