# infra-inventory
Infra Inventory Stream can start from an existing YAML or JSON inventory file. The importer posts each asset into the Go API, which validates dependencies, publishes Kafka events, and writes the searchable read model into OpenSearch.
