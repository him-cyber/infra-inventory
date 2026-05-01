#!/usr/bin/env sh
set -eu

API_URL="${API_URL:-http://127.0.0.1:8080}"

curl -fsS -X POST "$API_URL/api/org/import" \
  -H "Content-Type: application/json" \
  -H "X-Tenant: demo" \
  -d '{
    "provider":"azure",
    "account_id":"azure-verify-connectivity",
    "tenant_id":"demo-tenant",
    "subscription_id":"sub-verify-connectivity",
    "region":"eastus",
    "assets":[
      {
        "id":"az-verify-fw",
        "type":"network",
        "name":"Azure Verify Firewall",
        "owner":"it-ops",
        "environment":"prod",
        "region":"eastus",
        "service":"network-security",
        "version":17,
        "dependencies":["network-edge-router"],
        "risk":"high"
      }
    ]
  }'
