#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TESTKIT_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"
PID_FILE="${TESTKIT_DIR}/.anvil.pid"

if [ -f "$PID_FILE" ]; then
  PID="$(cat "$PID_FILE")"
  kill "$PID" 2>/dev/null || true
  rm -f "$PID_FILE"
fi

rm -f "${TESTKIT_DIR}/.anvil-accounts.json" "${TESTKIT_DIR}/.anvil.log"
