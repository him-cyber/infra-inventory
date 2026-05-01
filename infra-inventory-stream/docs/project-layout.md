# Project Layout

The code is split so product boundaries are visible quickly.

```text
cmd/
  api/             HTTP API process
  indexer/         Kafka consumer -> OpenSearch writer

internal/
  core/
    domain/        asset, event, config models
    security/      DFS dependency graph validation
    service/       inventory, org import, CVE, config intelligence, and ports
  adapters/
    auth/          Azure OIDC and encrypted cookie sessions
    httpapi/       REST handlers
    kafka/         franz-go producer/consumer
    opensearch/    OpenSearch document store
    config/        file/Azure Blob config source
  platform/
    ds/            radix tree, token bucket, retry heap, replay ring
    observability/ Prometheus and OpenTelemetry wiring

deploy/terraform/  Azure Container Apps, Event Hubs, Blob Storage, ACR
web/               thin React dashboard
```

Why this matters:

- `core` does not know about HTTP, Kafka, OpenSearch, or Azure.
- `adapters` contain replaceable infrastructure edges.
- `platform` holds reusable infrastructure utilities.
- `cmd` keeps the two backend services independently runnable and containerized.
