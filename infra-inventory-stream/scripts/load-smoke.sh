#!/usr/bin/env sh
set -eu

API_URL="${API_URL:-http://127.0.0.1:8080}"
COUNT="${COUNT:-5000}"
start="$(date +%s)"
i=1
while [ "$i" -le "$COUNT" ]; do
  curl -fsS -X POST "$API_URL/api/assets" \
    -H "Content-Type: application/json" \
    -H "X-Tenant: demo" \
    -d "{\"id\":\"asset-$i\",\"type\":\"service\",\"name\":\"Generated Service $i\",\"owner\":\"load\",\"environment\":\"dev\",\"region\":\"eastus\",\"service\":\"load-test\",\"version\":$i,\"dependencies\":[\"queue-inventory-events\"],\"risk\":\"low\"}" >/dev/null
  i=$((i + 1))
done
end="$(date +%s)"
elapsed=$((end - start))
[ "$elapsed" -eq 0 ] && elapsed=1
echo "sent=$COUNT elapsed_seconds=$elapsed approx_rps=$((COUNT / elapsed))"
