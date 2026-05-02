# Import Inventory

Infra Inventory Stream can start from an existing YAML or JSON inventory file. The importer posts each asset into the Go API, which validates dependencies, publishes Kafka events, and writes the searchable read model into OpenSearch.

YAML is useful when it comes from a real source of record: Azure Resource Graph, Terraform state, Kubernetes manifests, Helm values, Docker Compose, network exports, or a CMDB export. The importer normalizes those facts into a small asset graph so the UI can show ownership, dependencies, CVE risk, and configuration guardrails.

## Start the Stack

```bash
make demo
```

Open http://localhost:5173.

## Import the Example File

```bash
make import FILE=examples/local-office.yaml
```

Then open the UI and search for `San Jose`, `CMDB`, or `Firewall`.

## File Shape

```yaml
assets:
  - id: svc-cmdb-search-api
    type: service
    name: CMDB Search API
    owner: core-infra
    environment: prod
    region: eastus
    service: inventory
    version: 42
    dependencies:
      - kafka-inventory-events
      - os-inventory-index
    risk: medium
```

Required fields are `id`, `type`, `name`, `owner`, `environment`, `region`, `service`, `version`, and `risk`. `dependencies` can be empty.

Supported asset types include `service`, `frontend`, `worker`, `database`, `queue`, `cache`, `endpoint`, `network`, `kubernetes`, `container`, `function`, and `keyvault`.

## What Happens Internally

1. `cmd/importer` reads YAML or JSON.
2. The importer calls `POST /api/assets`.
3. The Go API validates the asset and rejects dependency cycles.
4. The API publishes an `asset.upserted` event to Kafka or Azure Event Hubs.
5. The indexer writes the latest state to OpenSearch.
6. CVE scoring and configuration recommendations read from that indexed state.

This gives the app a repeatable local path for real infrastructure data without requiring cloud credentials.
