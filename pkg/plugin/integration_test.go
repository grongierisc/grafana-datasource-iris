package plugin

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	_ "github.com/caretdev/go-irisnative"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/data"
	"github.com/grongierisc/grafana-datasource-iris/pkg/models"
)

func TestIRISIntegration(t *testing.T) {
	dsn := os.Getenv("IRIS_DSN")
	if dsn == "" {
		t.Skip("set IRIS_DSN to run IRIS integration tests")
	}

	ctx := context.Background()
	db, err := sql.Open("iris", dsn)
	if err != nil {
		t.Fatalf("open IRIS: %v", err)
	}
	defer func() {
		_ = db.Close()
	}()

	ds := &Datasource{
		db: db,
		settings: &models.PluginSettings{
			QueryTimeoutSeconds: 30,
			RowLimit:            1000,
		},
	}

	queryAndExpectFrame(t, ds, ctx, queryModel{
		RawSQL: "SELECT 1 AS value",
		Format: "table",
	}, 1)

	timeFrame := queryAndExpectFrame(t, ds, ctx, queryModel{
		RawSQL: "SELECT {ts '2026-07-07 10:00:00'} AS created_at, 42.5 AS value WHERE $__timeFilter({ts '2026-07-07 10:00:00'})",
		Format: "time_series",
	}, 2)
	if got := timeFrame.Fields[0].Type(); got != data.FieldTypeTime && got != data.FieldTypeNullableTime {
		t.Fatalf("expected time field, got %s", got)
	}
	if got := timeFrame.Fields[1].Type(); got != data.FieldTypeFloat64 && got != data.FieldTypeNullableFloat64 {
		t.Fatalf("expected float64 value field, got %s", got)
	}

	blockedPayload, err := json.Marshal(queryModel{RawSQL: "DROP TABLE grafana_iris_plugin_test"})
	if err != nil {
		t.Fatalf("marshal blocked query: %v", err)
	}
	blockedResp := ds.query(ctx, backend.PluginContext{}, backend.DataQuery{RefID: "B", JSON: blockedPayload})
	if blockedResp.Error == nil {
		t.Fatal("expected write query to be blocked")
	}

	tableName := fmt.Sprintf("grafana_iris_plugin_test_%d", time.Now().UnixNano())

	if _, err := db.ExecContext(ctx, fmt.Sprintf("CREATE TABLE %s (id INT, created_at TIMESTAMP, value DOUBLE)", tableName)); err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "access denied") {
			t.Skipf("IRIS_DSN account cannot create integration setup table: %v", err)
		}
		t.Fatalf("create setup table: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(ctx, fmt.Sprintf("DROP TABLE %s", tableName))
	})

	if _, err := db.ExecContext(ctx, fmt.Sprintf("INSERT INTO %s (id, created_at, value) VALUES (?, ?, ?)", tableName), 1, time.Date(2026, 7, 7, 10, 0, 0, 0, time.UTC), 42.5); err != nil {
		t.Fatalf("insert setup row: %v", err)
	}

	frame := queryAndExpectFrame(t, ds, ctx, queryModel{
		RawSQL: fmt.Sprintf("SELECT id, value FROM %s", tableName),
		Format: "table",
	}, 2)
	if got := frame.Fields[1].Type(); got != data.FieldTypeFloat64 && got != data.FieldTypeNullableFloat64 {
		t.Fatalf("expected float64 value field, got %s", got)
	}
	value, ok := frame.ConcreteAt(1, 0)
	if !ok {
		t.Fatal("expected non-null value field")
	}
	if fmt.Sprint(value) != "42.5" {
		t.Fatalf("expected DOUBLE value 42.5, got %#v", value)
	}
}

func queryAndExpectFrame(t *testing.T, ds *Datasource, ctx context.Context, model queryModel, fieldCount int) *data.Frame {
	t.Helper()

	payload, err := json.Marshal(model)
	if err != nil {
		t.Fatalf("marshal query: %v", err)
	}

	resp := ds.query(ctx, backend.PluginContext{}, backend.DataQuery{
		RefID: "A",
		JSON:  payload,
		TimeRange: backend.TimeRange{
			From: time.Date(2026, 7, 7, 9, 0, 0, 0, time.UTC),
			To:   time.Date(2026, 7, 7, 11, 0, 0, 0, time.UTC),
		},
	})
	if resp.Error != nil {
		t.Fatalf("query returned error: %v", resp.Error)
	}
	if len(resp.Frames) != 1 || len(resp.Frames[0].Fields) != fieldCount {
		t.Fatalf("expected one frame with two fields, got %#v", resp.Frames)
	}
	return resp.Frames[0]
}
