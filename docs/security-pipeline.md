# Security Pipeline

The project keeps image and dependency drift visible in three places: local commands, CI, and the app's CVE scoring workflow.

## Local Container Scan Export

Docker Scout requires a CLI login even when Docker Desktop is open.

```bash
make security-login
make security-export
```

Reports are written to `reports/`:

- `docker-scout-api.md`
- `docker-scout-indexer.md`
- `docker-scout-web.md`
- `docker-scout-redpanda.md`
- `docker-scout-opensearch.md`
- `docker-scout-seed.md`
- matching SARIF files for upload into code scanning tools

The directory is ignored by Git because scan reports can contain local image metadata.

## Runtime CVE Scoring

The UI's **Start CVE analysis** action calls:

```bash
POST /api/security/analyze
```

The backend reads indexed assets from OpenSearch, scores risk with `internal/core/ml`, and returns ranked findings with remediation text and ServiceNow evidence targets.

## How This Prevents Repeat Issues

- Dockerfiles pin current base images.
- `make verify` proves the app still works after dependency or image updates.
- `make security-export` produces a reviewable list instead of relying on screenshots.
- The imported asset graph becomes data for CVE scoring and configuration recommendations.
- The Go API rejects dependency cycles before publishing new inventory events.

## Before Opening a PR

```bash
make demo
make verify
make test
terraform -chdir=deploy/terraform validate
make security-export
```

If Scout asks for login, run `make security-login` once and rerun the export.
