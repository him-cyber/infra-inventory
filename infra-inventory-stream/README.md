# Infra Inventory Stream

Golang + Kafka + OpenSearch inventory platform for turning organization infrastructure into searchable operational knowledge.

Infra Inventory Stream accepts account and office asset changes, streams them through Kafka-compatible Azure Event Hubs or local Redpanda, indexes searchable state in OpenSearch, runs CVE risk scoring, recommends safer configuration, and produces ServiceNow-style incident/change evidence.

## Local Start

```bash
make doctor
make demo
make verify
```

Open http://localhost:5173.

Local requirements:

- Docker Desktop for the full stack.
- Terraform only for Azure validation/deploy.
- No Azure account is needed for the local demo.

Product walkthrough: [docs/product-walkthrough.md](docs/product-walkthrough.md)
Project layout: [docs/project-layout.md](docs/project-layout.md)
Import inventory: [docs/import-inventory.md](docs/import-inventory.md)
Azure SSO: [docs/azure-sso.md](docs/azure-sso.md)
Security pipeline: [docs/security-pipeline.md](docs/security-pipeline.md)
Open source notes: [docs/open-source.md](docs/open-source.md)

## Product Capabilities

| Capability | Implementation |
| --- | --- |
| Inventory ingestion | `POST /api/assets` and `POST /api/org/import` publish asset changes as Kafka events |
| Searchable knowledge | indexer writes OpenSearch documents consumed by search and topology APIs |
| CVE scoring | Go model scores live indexed assets and returns ranked findings |
| Configuration intelligence | backend recommends search fields, replay window, owner routing, and stateful controls from observed data |
| Secure user handling | Azure Entra ID OIDC, PKCE, verified ID tokens, encrypted HttpOnly session cookie |
| Cloud path | Terraform for Azure Container Apps, Event Hubs, Blob Storage, ACR, and Log Analytics |
| Reliability | DFS dependency guard, rate limits, replay buffer, retry heap, Prometheus metrics, OpenTelemetry traces |

## Architecture

```text
React portal
    |
    v
Go API ---- publishes ----> Kafka topic / Azure Event Hubs
    |                            |
    |                            v
    |                       Go indexer
    |                            |
    v                            v
Auth / config / topology     OpenSearch
    ^
    |
Azure Entra ID + Azure Blob config in cloud mode
```

## Local Demo

```bash
make demo
```

Open:

- UI: http://localhost:5173
- API: http://localhost:8080/healthz
- Metrics: http://localhost:8080/metrics

Try:

```bash
make import FILE=examples/local-office.yaml
curl "http://localhost:8080/api/assets/search?q=indexer"
curl "http://localhost:8080/api/topology"
curl "http://localhost:8080/api/analytics"
curl "http://localhost:8080/api/auth/session"
curl -X POST "http://localhost:8080/api/security/analyze"
curl -X POST "http://localhost:8080/api/config/recommendations"
curl -X POST "http://localhost:8080/api/servicenow/tickets"
curl "http://localhost:8080/api/stream/recent"
```

## Azure SSO

Local mode uses `AUTH_MODE=dev` and does not require cloud credentials. For Azure Entra ID:

```bash
AUTH_MODE=azure
AUTH_REQUIRED=true
AUTH_REDIRECT_URL=https://<app-host>/auth/callback
AZURE_TENANT_ID=<tenant-id>
AZURE_CLIENT_ID=<app-registration-client-id>
AZURE_CLIENT_SECRET=<client-secret>
AUTH_SESSION_KEY=<base64-encoded-32-byte-key>
```

The API uses Microsoft identity platform OIDC discovery, authorization-code flow with PKCE, ID-token verification, and an AES-GCM encrypted HttpOnly SameSite cookie. Set `COOKIE_SECURE=true` behind HTTPS.

Generate a session key:

```bash
openssl rand -base64 32
```

## Security Export

Container images are pinned to current demo-safe versions, including Go `1.25.7`, Redpanda `v26.1.5`, OpenSearch `3.5.0`, Nginx `1.29-alpine`, and Curl `8.19.0`.

To export Docker Scout findings:

```bash
make security-login
make security-export
```

Reports are written to `reports/` as Markdown and SARIF files. The directory is ignored by Git so local scans do not leak into commits.

## API

- `GET /api/auth/session`
- `POST /api/org/import`
- `GET /api/org/imports`
- `POST /api/assets`
- `GET /api/assets/search?q=&type=&env=&owner=`
- `GET /api/assets/{id}`
- `GET /api/topology`
- `GET /api/analytics`
- `POST /api/security/analyze`
- `POST /api/config/recommendations`
- `POST /api/servicenow/tickets`
- `GET /api/stream/recent`
- `POST /api/stream/replay`
- `GET /api/config/version`
- `GET /metrics`

## Azure Path

Terraform in `deploy/terraform` provisions:

- Azure Container Apps for `api`, `indexer`, `web`, and demo OpenSearch.
- Azure Event Hubs namespace and `inventory.events` event hub.
- Azure Storage account/container/blob for config.
- Azure Container Registry and Log Analytics.

The Kafka adapter switches to Event Hubs when `EVENTHUB_CONNECTION_STRING` is set and uses SASL/SSL with the Event Hubs Kafka endpoint.

Validate:

```bash
make terraform-check
```

Tear down cloud resources when finished:

```bash
make azure-down
```

## Config

`config/inventory-config.json` controls allowed asset types, search fields, tenant rate limits, index routing, and replay window size. The API reloads it every 10 seconds. In Azure mode, it reads the same JSON from Blob Storage using managed identity.

## Contributing

The core Go API, indexer, importer, data structures, and adapters are open for focused contributions. Small, readable patches that make infrastructure evidence easier to import, search, score, or explain are welcome.
