package models

import (
	"testing"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

func TestLoadPluginSettingsAppliesDefaults(t *testing.T) {
	settings, err := LoadPluginSettings(backend.DataSourceInstanceSettings{})
	if err != nil {
		t.Fatalf("LoadPluginSettings returned error: %v", err)
	}

	if settings.Host != DefaultHost {
		t.Fatalf("expected default host %q, got %q", DefaultHost, settings.Host)
	}
	if settings.Port != DefaultPort {
		t.Fatalf("expected default port %d, got %d", DefaultPort, settings.Port)
	}
	if settings.Namespace != DefaultNamespace {
		t.Fatalf("expected default namespace %q, got %q", DefaultNamespace, settings.Namespace)
	}
	if settings.RowLimit != DefaultRowLimit {
		t.Fatalf("expected default row limit %d, got %d", DefaultRowLimit, settings.RowLimit)
	}
}

func TestLoadPluginSettingsLoadsSecrets(t *testing.T) {
	settings, err := LoadPluginSettings(backend.DataSourceInstanceSettings{
		JSONData: []byte(`{"host":"iris.example.test","port":51773,"namespace":"APP","username":"grafana"}`),
		DecryptedSecureJSONData: map[string]string{
			"password": "secret",
		},
	})
	if err != nil {
		t.Fatalf("LoadPluginSettings returned error: %v", err)
	}

	if settings.Secrets.Password != "secret" {
		t.Fatalf("expected password to be loaded from secure json data")
	}
	if err := settings.Validate(); err != nil {
		t.Fatalf("expected settings to validate: %v", err)
	}
}

func TestPluginSettingsDSNEscapesCredentials(t *testing.T) {
	settings := &PluginSettings{
		Host:                "iris.example.test",
		Port:                1972,
		Namespace:           "USER",
		Username:            "user@example.test",
		QueryTimeoutSeconds: 7,
		RowLimit:            25,
		Secrets: &SecretPluginSettings{
			Password: "p@ss word/with:specials",
		},
	}

	dsn := settings.DSN()
	expected := "iris://user%40example.test:p%40ss%20word%2Fwith%3Aspecials@iris.example.test:1972/USER?max_rows=25&query_timeout=7"
	if dsn != expected {
		t.Fatalf("unexpected DSN:\nwant: %s\n got: %s", expected, dsn)
	}
}
