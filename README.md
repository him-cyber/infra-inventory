# Infra Inventory Stream

A local-first infrastructure knowledge console. Attach office, cloud, or service inventory; the Go backend streams each change through Kafka/Redpanda, indexes it in OpenSearch, scores CVE risk, recommends config changes, and creates ServiceNow-style ticket evidence.

## Run Locally

```bash
make doctor
make demo
```

Open http://localhost:5173.

Use the UI in this order:

1. Click `Attach JSON` to stream the sample inventory from the browser.
2. Click `Generate knowledge` to refresh analytics, CVE scoring, and config suggestions.
3. Drag a screenshot or topology image into `Evidence Drop`.
4. Click `Put on cooldown` for the highest-risk CI, then `Create ticket`.
5. Search inventory by asset, owner, service, risk, or region.

Verify the stack:

```bash
make verify
make test
```

Stop it:

```bash
make down
```

## Import Your Own Inventory

Use YAML or JSON shaped like `examples/local-office.yaml`:

```bash
make import FILE=examples/local-office.yaml
```

What happens:

1. `cmd/importer` reads the file.
2. `api` validates each asset and publishes `asset.upserted` events.
3. Redpanda carries the Kafka-compatible event stream.
4. `indexer` writes OpenSearch documents.
5. The web console reads the search index, accepts screenshot evidence, and generates remediation knowledge from the current state.

## Local URLs

- Web console: http://localhost:5173
- API health: http://localhost:8080/healthz
- Metrics: http://localhost:8080/metrics
- Search API: http://localhost:8080/api/assets/search?q=vpn

## Core Workflow

```text
Inventory JSON/YAML
  -> Go API
  -> Kafka-compatible stream
  -> Go indexer
  -> OpenSearch
  -> CVE scoring + config recommendations + image evidence + remediation ticket
```

## Remediation Workflow

Use this path when a weak service or endpoint needs action:

1. Attach inventory or import `examples/local-office.yaml`.
2. Run `Generate knowledge`.
3. Drop screenshots, diagrams, or scan evidence into the browser.
4. Put the highest-risk CI on cooldown.
5. Create the ServiceNow-style ticket payload.

The browser evidence is kept local for the demo. The ticket payload carries the CVE finding, impacted configuration items, and hardening notes.

## Useful API Calls

```bash
curl "http://localhost:8080/api/assets/search?q=network"
curl "http://localhost:8080/api/topology"
curl "http://localhost:8080/api/analytics"
curl -X POST "http://localhost:8080/api/security/analyze"
curl -X POST "http://localhost:8080/api/config/recommendations"
curl -X POST "http://localhost:8080/api/servicenow/tickets"
curl "http://localhost:8080/api/stream/recent"
```

## Security Checks

Go vulnerability checks run in GitHub Actions with `govulncheck ./...`.

Export local Docker Scout reports after logging in to Docker:

```bash
make security-login
make security-export
```

Reports go to `reports/`, which is ignored by Git.

## Azure SSO and Cloud Mode

Local mode works without Azure. For Azure Entra ID sign-in, set:

```bash
AUTH_MODE=azure
AUTH_REQUIRED=true
AUTH_REDIRECT_URL=https://<app-host>/auth/callback
AZURE_TENANT_ID=<tenant-id>
AZURE_CLIENT_ID=<app-registration-client-id>
AZURE_CLIENT_SECRET=<client-secret>
AUTH_SESSION_KEY=<base64-encoded-32-byte-key>
COOKIE_SECURE=true
```

Generate the session key:

```bash
openssl rand -base64 32
```

Terraform lives in `deploy/terraform` for Azure Container Apps, Event Hubs, Blob config, ACR, and Log Analytics:

```bash
make terraform-check
```

## Project Map

- `cmd/api`: Go HTTP API and product workflows.
- `cmd/indexer`: Kafka consumer that writes OpenSearch documents.
- `cmd/importer`: YAML/JSON inventory importer.
- `internal/core`: domain, service logic, security checks, data structures.
- `internal/adapters`: HTTP, Kafka, OpenSearch, auth, config, storage adapters.
- `web`: React knowledge console.
- `examples`: ready-to-import inventory files.
- `.github/workflows`: backend, web, Terraform, and security checks.

More detail:

- [Import inventory](docs/import-inventory.md)
- [Azure SSO](docs/azure-sso.md)
- [Security pipeline](docs/security-pipeline.md)
- [Project layout](docs/project-layout.md)

## Contributing

The core Go API, indexer, importer, data structures, and adapters are open for focused contributions. Small patches that make infrastructure evidence easier to import, search, score, or explain are welcome.
