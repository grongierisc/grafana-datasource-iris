package plugin

import (
	"encoding/json"
	"testing"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

func TestParseQueryDefaultsFormat(t *testing.T) {
	payload, err := json.Marshal(map[string]any{
		"rawSql": "SELECT 1",
	})
	if err != nil {
		t.Fatalf("json marshal: %v", err)
	}

	query, err := parseQuery(backend.DataQuery{JSON: payload})
	if err != nil {
		t.Fatalf("parseQuery returned error: %v", err)
	}
	if query.Format != "table" {
		t.Fatalf("expected table format, got %q", query.Format)
	}
}

func TestParseQueryRequiresSQL(t *testing.T) {
	_, err := parseQuery(backend.DataQuery{JSON: []byte(`{"format":"table"}`)})
	if err == nil {
		t.Fatal("expected empty SQL query to fail")
	}
}
