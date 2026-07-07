# InterSystems IRIS datasource for Grafana

Read-only SQL datasource plugin for querying InterSystems IRIS from Grafana dashboards, Explore, and Grafana-managed alerts.

## Status

This is a local-development plugin. It is configured to run unsigned in the included Docker Compose environment and is not packaged for the public Grafana plugin catalog.

The backend uses `github.com/caretdev/go-irisnative`, an alpha community `database/sql` driver for IRIS. Until the upstream driver PR is accepted, the module is replaced with `github.com/grongierisc/go-irisnative` to correctly decode IRIS numeric list values and expose SQL column metadata. The driver use is isolated in the backend connection layer so the connection strategy can be replaced later if needed.

## Prerequisites

- Node.js compatible with the generated Grafana toolchain. The current scaffold declares `node >=22`.
- npm 10 or later.
- Go 1.25.5 or later.
- Mage.
- Docker.

This machine previously had Node 18/npm 9 and no local Go or Mage, so use a newer toolchain before running the build commands directly on the host.

## Local development

Install frontend dependencies:

```bash
npm install
```

Build the frontend:

```bash
npm run build
```

Build the backend:

```bash
mage -v build:linux
```

Start Grafana and IRIS:

```bash
docker compose up
```

Local services:

- Grafana: http://localhost:3000
- IRIS SuperServer: `localhost:1972`
- IRIS portal: http://localhost:52773

The Compose stack provisions a datasource named `InterSystems IRIS`. Grafana connects to the IRIS container through the Compose service name `iris`; host tools can connect through `localhost:1972`.

Development credentials are local only:

- Username: `_SYSTEM`
- Password: `grafana-iris-dev-password`
- Namespace: `USER`

The IRIS container uses `iris-main --password-file` with `docker/iris-password.txt` so xDBC connections do not depend on the first-login password-change flow.

## Querying

The query editor sends this model to the backend:

- `rawSql`: SQL text.
- `format`: `table` or `time_series`.
- `rowLimit`: optional per-query row limit.

The backend only allows read-only queries whose first SQL keyword is `SELECT` or `WITH`. It blocks multiple statements and write/DDL keywords such as `INSERT`, `UPDATE`, `DELETE`, `DROP`, `ALTER`, and `CREATE`. Use a read-only IRIS account as the primary security boundary; the plugin guard is an additional safety layer.

Supported macros:

- `$__timeFilter(column)`
- `$__timeFrom(column)`
- `$__timeTo(column)`
- `$__interval`
- `$__interval_ms`
- `$__timeGroup(column, interval)` for `s`, `m`, `h`, and `d` intervals.

## Tests

Frontend unit checks:

```bash
npm run test:ci
```

Backend unit checks:

```bash
go test ./pkg/...
```

Optional IRIS integration checks:

```bash
IRIS_DSN='iris://_SYSTEM:grafana-iris-dev-password@localhost:1972/USER' go test ./pkg/plugin -run TestIRISIntegration
```

If you run Go from Docker, attach the test container to the Compose network and use the `iris` service name:

```bash
docker run --rm --network grafana-datasource-iris_default -e IRIS_DSN='iris://_SYSTEM:grafana-iris-dev-password@iris:1972/USER' -v "$PWD:/workspace" -w /workspace golang:1.25-bookworm go test ./pkg/plugin -run TestIRISIntegration -v
```

E2E checks require a running Compose stack and built plugin binaries:

```bash
npm run e2e
```

Any change to `src/plugin.json` requires restarting Grafana.
