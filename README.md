# InterSystems IRIS datasource for Grafana

Read-only SQL datasource plugin for querying InterSystems IRIS from Grafana dashboards, Explore, and Grafana-managed alerts.

## Status

This project is preparing for public Grafana plugin catalog submission. It builds and runs as an unsigned plugin in the included Docker Compose environment, and it is structured for a future signed release.

Before a public catalog or Marketplace submission, finish the release metadata and packaging work:

- Review the catalog screenshots and project links in `src/plugin.json`.
- Move `CHANGELOG.md` from `1.0.0 (Unreleased)` to a dated release entry.
- Build a release archive with backend binaries for the supported operating systems and architectures.
- Run the Grafana plugin validator against the ZIP archive.
- Submit the archive URL, source URL, SHA1 hash, and testing guidance to Grafana Labs for review.
- Sign the public plugin only after Grafana Labs approves the submission and grants a signature level.

The backend uses `github.com/caretdev/go-irisnative`, a community `database/sql` driver for IRIS. Driver usage is isolated in the backend connection layer so the connection strategy can be replaced later if needed.

## Requirements

- Grafana 10.0.0 or later.
- InterSystems IRIS with SuperServer/xDBC access enabled.
- A read-only IRIS SQL account for production use.
- Node.js compatible with the generated Grafana toolchain. The current scaffold declares `node >=22`.
- npm 10 or later.
- Go 1.25.5 or later.
- Mage.
- Docker for the local Grafana and IRIS development stack.

This machine previously had Node 18/npm 9 and no local Go or Mage, so use a newer toolchain before running the build commands directly on the host.

## Features

- Backend datasource plugin using Grafana's Go plugin SDK.
- SQL queries against InterSystems IRIS through Go `database/sql`.
- Table and time series query formats.
- Grafana-managed alerting support.
- Secure password storage through Grafana `secureJsonData`.
- Conservative connection pool settings.
- Query timeout and row limit controls.
- Read-only SQL guard for `SELECT` and `WITH` queries.
- Grafana time macro expansion for common dashboard queries.

## Local development

Install frontend dependencies:

```bash
npm install
```

Build the frontend:

```bash
npm run build
```

Build the backend for Linux:

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

## Configuration

Datasource settings:

- Host: IRIS server host, for example `localhost` or `iris`.
- Port: IRIS SuperServer port, usually `1972`.
- Namespace: IRIS namespace, for example `USER`.
- Username: IRIS SQL user.
- Password: stored only in Grafana secure JSON data.
- Query timeout: backend query timeout in seconds.
- Row limit: maximum rows returned by default.
- Max open connections, max idle connections, and connection lifetime: backend connection pool settings.

Use database permissions as the primary security boundary. The plugin blocks non-read SQL as a safety layer, but the configured IRIS account should still only have the minimum required `SELECT` privileges.

## Querying

The query editor sends this model to the backend:

- `rawSql`: SQL text.
- `format`: `table` or `time_series`.
- `rowLimit`: optional per-query row limit.

The backend only allows read-only queries whose first SQL keyword is `SELECT` or `WITH`. It blocks multiple statements and write/DDL keywords such as `INSERT`, `UPDATE`, `DELETE`, `DROP`, `ALTER`, and `CREATE`.

Supported macros:

- `$__timeFilter(column)`
- `$__timeFrom(column)`
- `$__timeTo(column)`
- `$__interval`
- `$__interval_ms`
- `$__timeGroup(column, interval)` for `s`, `m`, `h`, and `d` intervals.

## Tests

Frontend and static checks:

```bash
npm run typecheck
npm run lint
npm run test:ci
npm run build
```

Backend unit checks:

```bash
go test ./pkg/...
golangci-lint run ./...
```

Backend release build:

```bash
mage -v
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

## Packaging and publication

For local release packaging, build the frontend and backend, then package the built plugin under the plugin ID:

```bash
npm ci
npm run build
mage -v
rm -rf package
mkdir -p package
cp -R dist package/grongierisc-iris-datasource
(cd package && zip -r ../grongierisc-iris-datasource-1.0.0.zip grongierisc-iris-datasource)
shasum -a 1 grongierisc-iris-datasource-1.0.0.zip
```

For public catalog submission, Grafana Labs currently reviews the plugin before public signing is available. Submit the packaged ZIP URL, source URL, SHA1 hash, and test guidance through the Grafana Cloud plugin submission flow.

Grafana Marketplace is the optional commercial distribution path inside the Grafana plugin catalog. For a free community release, use the standard public catalog submission flow.

After approval, export a Grafana access policy token with `plugins:write` scope and sign the plugin:

```bash
export GRAFANA_ACCESS_POLICY_TOKEN=<token>
npm run sign
```

For private distribution, pass the expected Grafana root URLs to the signing command:

```bash
npm run sign -- --rootUrls https://example.com/grafana
```

Publishing references:

- [Package a plugin](https://grafana.com/developers/plugin-tools/publish-a-plugin/package-a-plugin)
- [Sign a plugin](https://grafana.com/developers/plugin-tools/publish-a-plugin/sign-a-plugin)
- [Publish or update a plugin](https://grafana.com/developers/plugin-tools/publish-a-plugin/publish-a-plugin)
