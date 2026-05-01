# Open Source Notes

The core Go implementation is intentionally kept small and readable:

- `cmd/api`: HTTP API, auth wiring, metrics, and service startup.
- `cmd/indexer`: Kafka consumer that writes the OpenSearch read model.
- `cmd/importer`: YAML/JSON inventory importer.
- `internal/core`: domain services, CVE scoring, topology, dependency validation.
- `internal/adapters`: HTTP, Kafka, OpenSearch, auth, and config adapters.
- `internal/platform/ds`: radix lookup, token bucket, ring buffer, retry queue.

Contributions are welcome for focused improvements: new import adapters, better scoring features, extra OpenSearch mappings, Azure deployment hardening, and UI workflows that make infrastructure evidence easier to understand.
