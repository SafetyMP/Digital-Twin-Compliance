#!/usr/bin/env bash
# Inserts 51 payments for the same source account within the last hour to trigger INT-M001.
set -euo pipefail

CORE_URL="${CORE_BANKING_DB_URL:-postgres://core:core@localhost:5433/core_banking?sslmode=disable}"
BURST_COUNT="${BURST_COUNT:-51}"

ACCOUNT_ID=$(psql "$CORE_URL" -tA -c "SELECT account_id FROM accounts LIMIT 1")
DEST_ID=$(psql "$CORE_URL" -tA -c "SELECT account_id FROM accounts WHERE account_id != '$ACCOUNT_ID' LIMIT 1")

if [[ -z "$ACCOUNT_ID" || -z "$DEST_ID" ]]; then
  echo "No accounts found for payment burst" >&2
  exit 1
fi

echo "Bursting $BURST_COUNT payments from account $ACCOUNT_ID..."
for i in $(seq 1 "$BURST_COUNT"); do
  psql "$CORE_URL" -v ON_ERROR_STOP=1 -c "
    INSERT INTO payments (source_account_id, destination_account_id, amount, currency, status, initiated_at, updated_at)
    VALUES ('$ACCOUNT_ID', '$DEST_ID', 100.00, 'EUR', 'Pending', now() - interval '30 minutes', now());
  " >/dev/null
done

echo "Payment burst complete."
