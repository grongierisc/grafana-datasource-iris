# InterSystems IRIS

Read-only SQL datasource for querying InterSystems IRIS from Grafana panels, Explore, and Grafana-managed alerts.

This plugin uses the Grafana backend plugin runtime and connects to IRIS through the Go `database/sql` driver. It is intended for SQL queries that return table or time series data.

## Configuration

Create a datasource and set the IRIS connection fields:

- Host: IRIS server host, for example `localhost` or `iris`
- Port: IRIS SuperServer port, usually `1972`
- Namespace: IRIS namespace, for example `USER`
- Username: IRIS SQL user
- Password: stored in Grafana secure JSON data
- Query timeout: backend query timeout in seconds
- Row limit: maximum rows returned by default
- Max open connections, max idle connections, connection lifetime: backend connection pool settings

The datasource user should be read-only. The plugin also blocks non-read SQL as a safety layer, but database permissions remain the primary security boundary.

## Query Formats

The query editor has two formats:

- Table: use for regular SQL result sets.
- Time series: use when the result contains a time column and one or more numeric value columns.

The backend accepts SQL beginning with `SELECT` or `WITH`. Multiple statements and write/DDL keywords such as `INSERT`, `UPDATE`, `DELETE`, `DROP`, `ALTER`, and `CREATE` are blocked.

## Table Query Example

Create an IRIS table with mixed column types:

```sql
CREATE TABLE SQLUser.grafana_iris_table_example (
  id INTEGER,
  created_at TIMESTAMP,
  service VARCHAR(50),
  status VARCHAR(20),
  value DOUBLE
);
```

Insert sample rows:

```sql
INSERT INTO SQLUser.grafana_iris_table_example
  (id, created_at, service, status, value)
VALUES
  (1, {ts '2026-07-07 10:00:00'}, 'orders', 'ok', 42.5);

INSERT INTO SQLUser.grafana_iris_table_example
  (id, created_at, service, status, value)
VALUES
  (2, {ts '2026-07-07 10:05:00'}, 'billing', 'warning', 18.75);

INSERT INTO SQLUser.grafana_iris_table_example
  (id, created_at, service, status, value)
VALUES
  (3, {ts '2026-07-07 10:10:00'}, 'shipping', 'ok', 31.2);
```

Set the query format to **Table** and query:

```sql
SELECT
  id,
  created_at,
  service,
  status,
  value
FROM SQLUser.grafana_iris_table_example
WHERE $__timeFilter(created_at)
ORDER BY created_at
```

Grafana renders one table row per SQL row. `created_at` is returned as a time field, `value` as a numeric field, and `service`/`status` as string fields.

## Time Series Example

Create an IRIS table with a real `TIMESTAMP` column and a numeric value column:

```sql
CREATE TABLE SQLUser.grafana_iris_timeseries (
  id INTEGER,
  created_at TIMESTAMP,
  metric VARCHAR(50),
  value DOUBLE
);
```

Insert sample rows:

```sql
INSERT INTO SQLUser.grafana_iris_timeseries
  (id, created_at, metric, value)
VALUES
  (1, {ts '2026-07-07 10:00:00'}, 'temperature', 21.5);

INSERT INTO SQLUser.grafana_iris_timeseries
  (id, created_at, metric, value)
VALUES
  (2, {ts '2026-07-07 10:01:00'}, 'temperature', 22.1);

INSERT INTO SQLUser.grafana_iris_timeseries
  (id, created_at, metric, value)
VALUES
  (3, {ts '2026-07-07 10:02:00'}, 'temperature', 22.8);
```

Set the query format to **Time series** and query:

```sql
SELECT
  created_at,
  value
FROM SQLUser.grafana_iris_timeseries
WHERE
  metric = 'temperature'
  AND $__timeFilter(created_at)
ORDER BY created_at
```

Grafana receives `created_at` as a time field and `value` as a numeric field.

For multiple named series, include a text label column:

```sql
SELECT
  created_at,
  metric,
  value
FROM SQLUser.grafana_iris_timeseries
WHERE $__timeFilter(created_at)
ORDER BY metric, created_at
```

## Grouped Time Series Example

Use `$__timeGroup(column, interval)` to bucket timestamps. Supported interval suffixes are `s`, `m`, `h`, and `d`.

```sql
SELECT
  $__timeGroup(created_at, $__interval) AS bucket,
  AVG(value) AS value
FROM SQLUser.grafana_iris_timeseries
WHERE
  metric = 'temperature'
  AND $__timeFilter(created_at)
GROUP BY $__timeGroup(created_at, $__interval)
ORDER BY bucket
```

You can also use a fixed interval:

```sql
SELECT
  $__timeGroup(created_at, 5m) AS bucket,
  AVG(value) AS value
FROM SQLUser.grafana_iris_timeseries
WHERE $__timeFilter(created_at)
GROUP BY $__timeGroup(created_at, 5m)
ORDER BY bucket
```

## Supported Macros

The plugin expands these macros before sending SQL to IRIS:

- `$__timeFilter(column)`: expands to an inclusive `column >= from AND column <= to` filter using IRIS timestamp literals.
- `$__timeFrom(column)`: expands to `column >= from`.
- `$__timeTo(column)`: expands to `column <= to`.
- `$__interval`: Grafana query interval.
- `$__interval_ms`: Grafana query interval in milliseconds.
- `$__timeGroup(column, interval)`: groups timestamps with IRIS `DATEADD` and `DATEDIFF`.

Unsupported `$__timeGroup` interval units return a query error.

## Local Development Notes

For local Docker development, the included Compose stack exposes:

- Grafana: `http://localhost:3000`
- IRIS SuperServer: `localhost:1972`
- IRIS management portal: `http://localhost:52773`

The local datasource is provisioned for namespace `USER` with username `_SYSTEM` and the dev-only password from the repository's Docker setup.
