APP=infra-inventory-stream
FILE?=examples/local-office.yaml
TENANT?=demo

.PHONY: demo up down seed import logs test smoke verify doctor security-export security-login terraform-check azure-down

demo: up seed
	@echo "UI: http://localhost:5173"
	@echo "API: http://localhost:8080/healthz"
	@echo "Search: curl 'http://localhost:8080/api/assets/search?q=Search'"

doctor:
	@command -v docker >/dev/null || (echo "Docker is required for the local demo" && exit 1)
	@docker compose version >/dev/null
	@command -v terraform >/dev/null && terraform version | head -n 1 || echo "Terraform optional: install it for Azure validation"
	@echo "Local toolchain ready"

up:
	docker compose up --build -d --remove-orphans

down:
	docker compose down

seed:
	docker compose run --rm seed

import:
	docker run --rm -v "$$(pwd)":/src -w /src golang:1.26.2-alpine go run ./cmd/importer -file /src/$(FILE) -tenant $(TENANT) -api http://host.docker.internal:8080

logs:
	docker compose logs -f api indexer

test:
	docker run --rm -v "$$(pwd)":/src -w /src golang:1.26.2-alpine go test ./...

smoke:
	COUNT=5000 sh scripts/load-smoke.sh

verify:
	docker compose run --rm --entrypoint curl seed -fsS http://api:8080/healthz
	docker compose run --rm --entrypoint curl seed -fsS http://api:8080/api/config/version
	docker compose run --rm --entrypoint curl seed -fsS http://api:8080/api/auth/session
	docker compose run --rm --entrypoint curl seed -fsS http://api:8080/api/org/imports
	docker compose run --rm seed sh /scripts/import-sample.sh
	docker compose run --rm --entrypoint curl seed -fsS http://api:8080/api/topology
	docker compose run --rm --entrypoint curl seed -fsS http://api:8080/api/analytics
	docker compose run --rm --entrypoint curl seed -fsS -X POST http://api:8080/api/config/recommendations
	docker compose run --rm --entrypoint curl seed -fsS -X POST http://api:8080/api/security/analyze
	docker compose run --rm --entrypoint curl seed -fsS -X POST http://api:8080/api/servicenow/tickets
	docker compose run --rm --entrypoint curl seed -fsS "http://api:8080/api/assets/search?q=Search"

security-export:
	mkdir -p reports
	docker scout cves --format sarif --output reports/docker-scout-api.sarif.json local://infra-inventory-stream-api:latest
	docker scout cves --format sarif --output reports/docker-scout-indexer.sarif.json local://infra-inventory-stream-indexer:latest
	docker scout cves --format sarif --output reports/docker-scout-web.sarif.json local://infra-inventory-stream-web:latest
	docker scout cves --format sarif --output reports/docker-scout-redpanda.sarif.json local://docker.redpanda.com/redpandadata/redpanda:v26.1.5
	docker scout cves --format sarif --output reports/docker-scout-opensearch.sarif.json local://opensearchproject/opensearch:3.5.0
	docker scout cves --format sarif --output reports/docker-scout-seed.sarif.json local://curlimages/curl:8.19.0
	docker scout cves --format markdown --output reports/docker-scout-api.md local://infra-inventory-stream-api:latest
	docker scout cves --format markdown --output reports/docker-scout-indexer.md local://infra-inventory-stream-indexer:latest
	docker scout cves --format markdown --output reports/docker-scout-web.md local://infra-inventory-stream-web:latest
	docker scout cves --format markdown --output reports/docker-scout-redpanda.md local://docker.redpanda.com/redpandadata/redpanda:v26.1.5
	docker scout cves --format markdown --output reports/docker-scout-opensearch.md local://opensearchproject/opensearch:3.5.0
	docker scout cves --format markdown --output reports/docker-scout-seed.md local://curlimages/curl:8.19.0

security-login:
	docker login

terraform-check:
	terraform fmt -check -recursive deploy/terraform
	terraform -chdir=deploy/terraform init -backend=false
	terraform -chdir=deploy/terraform validate

azure-down:
	terraform -chdir=deploy/terraform destroy
