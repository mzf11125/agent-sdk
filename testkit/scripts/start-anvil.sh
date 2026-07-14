#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TESTKIT_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"
PID_FILE="${TESTKIT_DIR}/.anvil.pid"
LOG_FILE="${TESTKIT_DIR}/.anvil.log"
ACCOUNTS_FILE="${TESTKIT_DIR}/.anvil-accounts.json"
PORT="${ANVIL_PORT:-8545}"

if [ -f "$PID_FILE" ] && kill -0 "$(cat "$PID_FILE")" 2>/dev/null; then
  echo "anvil already running (pid $(cat "$PID_FILE"))" >&2
  exit 0
fi

anvil --port "$PORT" --mnemonic-random > "$LOG_FILE" 2>&1 &
echo $! > "$PID_FILE"

READY=0
for _ in $(seq 1 50); do
  if curl -s -o /dev/null -X POST -H "Content-Type: application/json" \
       --data '{"jsonrpc":"2.0","method":"eth_chainId","params":[],"id":1}' \
       "http://127.0.0.1:${PORT}"; then
    READY=1
    break
  fi
  sleep 0.2
done

if [ "$READY" -ne 1 ]; then
  echo "anvil did not become ready on port ${PORT}; see ${LOG_FILE}" >&2
  exit 1
fi

ADDRESSES="$(awk '/^Available Accounts/{flag=1; next} /^Private Keys/{flag=0} flag' "$LOG_FILE" | grep -oE '0x[0-9a-fA-F]{40}')"
KEYS="$(awk '/^Private Keys/{flag=1; next} /^Wallet/{flag=0} flag' "$LOG_FILE" | grep -oE '0x[0-9a-fA-F]{64}')"

if [ -z "$ADDRESSES" ] || [ -z "$KEYS" ]; then
  echo "failed to parse anvil accounts from ${LOG_FILE}" >&2
  exit 1
fi

paste -d' ' <(echo "$ADDRESSES") <(echo "$KEYS") \
  | jq -R 'split(" ") | {address: .[0], privateKey: .[1]}' \
  | jq -s '{accounts: .}' \
  > "$ACCOUNTS_FILE"
