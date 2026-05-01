#!/usr/bin/env sh
set -eu

API_URL="${API_URL:-http://127.0.0.1:8080}"

post_asset() {
  curl -fsS -X POST "$API_URL/api/assets" \
    -H "Content-Type: application/json" \
    -H "X-Tenant: demo" \
    -d "$1" >/dev/null
}

post_asset '{"id":"network-edge-router","type":"network","name":"Mountain View Edge Router","owner":"it-ops","environment":"prod","region":"mountain-view-office","service":"office-network","version":12,"dependencies":[],"risk":"low"}'
post_asset '{"id":"network-branch-wifi","type":"network","name":"Branch Wi-Fi Mesh","owner":"it-ops","environment":"prod","region":"mountain-view-office","service":"office-network","version":22,"dependencies":["network-edge-router"],"risk":"medium"}'
post_asset '{"id":"endpoint-laptop-fleet","type":"endpoint","name":"Employee Laptop Fleet","owner":"helpdesk","environment":"prod","region":"mountain-view-office","service":"employee-support","version":108,"dependencies":["network-branch-wifi","svc-catalog-api"],"risk":"medium"}'
post_asset '{"id":"servicenow-support-view","type":"frontend","name":"ServiceNow Support View","owner":"product","environment":"prod","region":"eastus","service":"support-workflow","version":9,"dependencies":["svc-catalog-api","opensearch-assets"],"risk":"low"}'
post_asset '{"id":"svc-catalog-api","type":"service","name":"Catalog API","owner":"platform","environment":"prod","region":"eastus","service":"catalog","version":41,"dependencies":["queue-inventory-events","db-catalog"],"risk":"medium"}'
post_asset '{"id":"queue-inventory-events","type":"queue","name":"Inventory Events","owner":"core-infra","environment":"prod","region":"eastus","service":"inventory","version":7,"dependencies":[],"risk":"low"}'
post_asset '{"id":"worker-indexer","type":"worker","name":"Search Indexer","owner":"search","environment":"prod","region":"eastus","service":"inventory","version":18,"dependencies":["queue-inventory-events","opensearch-assets"],"risk":"high"}'
post_asset '{"id":"opensearch-assets","type":"database","name":"Inventory Search Index","owner":"search","environment":"prod","region":"eastus","service":"inventory","version":3,"dependencies":[],"risk":"medium"}'
post_asset '{"id":"db-catalog","type":"database","name":"Catalog Metadata Store","owner":"platform","environment":"prod","region":"eastus","service":"catalog","version":15,"dependencies":[],"risk":"medium"}'
post_asset '{"id":"svc-now-assist-router","type":"service","name":"Now Assist Context Router","owner":"ml-platform","environment":"prod","region":"eastus","service":"now-assist","version":27,"dependencies":["opensearch-assets","model-cve-risk"],"risk":"medium"}'
post_asset '{"id":"model-cve-risk","type":"worker","name":"CVE Risk Model Scorer","owner":"ml-platform","environment":"prod","region":"eastus","service":"security-ai","version":11,"dependencies":["feature-store-risk","opensearch-assets"],"risk":"high"}'
post_asset '{"id":"feature-store-risk","type":"database","name":"Risk Feature Store","owner":"data","environment":"prod","region":"eastus","service":"security-ai","version":8,"dependencies":[],"risk":"medium"}'
post_asset '{"id":"svc-ticket-webhook","type":"service","name":"ServiceNow Ticket Webhook","owner":"core-infra","environment":"prod","region":"eastus","service":"support-workflow","version":33,"dependencies":["queue-inventory-events","svc-now-assist-router"],"risk":"medium"}'
post_asset '{"id":"endpoint-exec-laptops","type":"endpoint","name":"Executive Laptop Ring","owner":"helpdesk","environment":"prod","region":"mountain-view-office","service":"employee-support","version":44,"dependencies":["network-branch-wifi","svc-ticket-webhook"],"risk":"high"}'
post_asset '{"id":"network-conf-room","type":"network","name":"Conference Room Network","owner":"it-ops","environment":"prod","region":"mountain-view-office","service":"office-network","version":19,"dependencies":["network-edge-router"],"risk":"low"}'
post_asset '{"id":"svc-change-policy","type":"service","name":"Change Policy Evaluator","owner":"platform","environment":"prod","region":"eastus","service":"change-management","version":24,"dependencies":["db-catalog","opensearch-assets"],"risk":"medium"}'
post_asset '{"id":"queue-cve-analysis","type":"queue","name":"CVE Analysis Jobs","owner":"data","environment":"prod","region":"eastus","service":"security-ai","version":6,"dependencies":[],"risk":"low"}'
post_asset '{"id":"worker-cve-enricher","type":"worker","name":"CVE Evidence Enricher","owner":"security","environment":"prod","region":"eastus","service":"security-ai","version":14,"dependencies":["queue-cve-analysis","feature-store-risk"],"risk":"medium"}'
post_asset '{"id":"db-servicenow-cache","type":"database","name":"ServiceNow Evidence Cache","owner":"core-infra","environment":"prod","region":"eastus","service":"support-workflow","version":5,"dependencies":[],"risk":"medium"}'

echo "seeded inventory assets into $API_URL"
