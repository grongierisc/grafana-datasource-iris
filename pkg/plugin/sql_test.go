package plugin

import (
	"strings"
	"testing"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

func TestValidateReadOnlySQLAllowsSelectAndWith(t *testing.T) {
	allowed := []string{
		"SELECT * FROM Sample.Person",
		"  -- comment\nSELECT Name FROM Sample.Person",
		"WITH people AS (SELECT Name FROM Sample.Person) SELECT * FROM people",
		"SELECT ';' AS semicolon",
	}

	for _, sql := range allowed {
		if err := validateReadOnlySQL(sql); err != nil {
			t.Fatalf("expected query to be allowed %q: %v", sql, err)
		}
	}
}

func TestValidateReadOnlySQLBlocksWritesAndMultipleStatements(t *testing.T) {
	blocked := []string{
		"INSERT INTO Sample.Person(Name) VALUES ('Ada')",
		"UPDATE Sample.Person SET Name = 'Ada'",
		"DELETE FROM Sample.Person",
		"DROP TABLE Sample.Person",
		"SELECT * FROM Sample.Person; DROP TABLE Sample.Person",
		"WITH deleted AS (DELETE FROM Sample.Person RETURNING *) SELECT * FROM deleted",
	}

	for _, sql := range blocked {
		if err := validateReadOnlySQL(sql); err == nil {
			t.Fatalf("expected query to be blocked %q", sql)
		}
	}
}

func TestInterpolateSQLTimeMacros(t *testing.T) {
	from := time.Date(2026, 7, 7, 10, 11, 12, 0, time.UTC)
	to := time.Date(2026, 7, 7, 11, 12, 13, 0, time.UTC)

	got, err := interpolateSQL("SELECT * FROM events WHERE $__timeFilter(created_at)", backend.DataQuery{
		TimeRange: backend.TimeRange{From: from, To: to},
	})
	if err != nil {
		t.Fatalf("interpolateSQL returned error: %v", err)
	}

	expected := "created_at >= {ts '2026-07-07 10:11:12'} AND created_at <= {ts '2026-07-07 11:12:13'}"
	if !strings.Contains(got, expected) {
		t.Fatalf("expected interpolated SQL to contain %q, got %q", expected, got)
	}
}

func TestInterpolateSQLTimeGroup(t *testing.T) {
	got, err := interpolateSQL("SELECT $__timeGroup(created_at, 5m) AS time FROM events", backend.DataQuery{})
	if err != nil {
		t.Fatalf("interpolateSQL returned error: %v", err)
	}

	expected := "DATEADD(second, FLOOR(DATEDIFF(second, {ts '1970-01-01 00:00:00'}, created_at) / 300) * 300, {ts '1970-01-01 00:00:00'})"
	if !strings.Contains(got, expected) {
		t.Fatalf("expected interpolated SQL to contain %q, got %q", expected, got)
	}
}

func TestInterpolateSQLRejectsUnsupportedTimeGroupInterval(t *testing.T) {
	_, err := interpolateSQL("SELECT $__timeGroup(created_at, 1w) AS time FROM events", backend.DataQuery{})
	if err == nil {
		t.Fatal("expected unsupported interval to fail")
	}
}
