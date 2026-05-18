# Product Walkthrough

Infra Inventory Stream turns organization infrastructure data into searchable operational knowledge. The useful correlation is: source facts become a CMDB-like graph, graph relationships expose blast radius, CVE scoring ranks weak points, and configuration automation turns those signals into a reviewable workflow action.

## Run Locally

```bash
make demo
```

Open http://localhost:5173.

The local product path is:

```text
React UI -> Go API -> Kafka topic -> Go indexer -> OpenSearch -> Go search/API read models
```

## Core Workflows

1. Search the indexed inventory by asset, owner, service, region, or risk.
2. Run CVE scoring against live OpenSearch assets.
3. Generate a model-backed incident brief from top findings, graph context, Kafka evidence, and the ServiceNow target when `OPENAI_API_KEY` is configured.
4. Generate configuration recommendations from observed owners, services, risks, and replay behavior.
5. Import a cloud account inventory snapshot and publish each discovered asset through Kafka.
6. Review and submit a ServiceNow-style incident/change payload with evidence from Kafka, OpenSearch, and the incident brief.

## Security And Identity

- Azure Entra ID support uses OIDC discovery, authorization-code flow with PKCE, and verified ID tokens.
- Browser identity state is stored in an AES-GCM encrypted, HttpOnly, SameSite cookie.
- Production transport expects TLS at ingress.
- Azure mode uses Event Hubs Kafka over SASL_SSL.
- Local mode keeps auth optional so the stack starts without cloud credentials.

## What To Review

- `cmd/api`: HTTP API, auth setup, inventory publish path.
- `cmd/indexer`: Kafka consumer and OpenSearch writer.
- `internal/adapters/auth`: Azure OIDC and encrypted cookie session handling.
- `internal/core/service`: inventory, CVE scoring, config intelligence, org import.
- `internal/platform`: rate limits, replay ring, radix routing, retry queue, observability.
- `deploy/terraform`: Azure Container Apps, Event Hubs, Blob config, ACR, Log Analytics.
