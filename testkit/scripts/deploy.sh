#!/usr/bin/env bash
set -euo pipefail

if [ "$#" -ne 2 ]; then
  echo "usage: deploy.sh <category/ERCXXXX> <ScriptContractName>" >&2
  exit 1
fi

ERC_PATH="$1"
CONTRACT_NAME="$2"
PORT="${ANVIL_PORT:-8545}"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TESTKIT_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"
ACCOUNTS_FILE="${TESTKIT_DIR}/.anvil-accounts.json"

DEPLOYER_KEY="$(jq -r '.accounts[0].privateKey' "$ACCOUNTS_FILE")"

cd "$TESTKIT_DIR"

forge script "script/${ERC_PATH}/${CONTRACT_NAME}.s.sol:${CONTRACT_NAME}" \
  --broadcast \
  --rpc-url "http://127.0.0.1:${PORT}" \
  --private-key "$DEPLOYER_KEY" \
  > /dev/null

CHAIN_ID="$(cast chain-id --rpc-url "http://127.0.0.1:${PORT}")"
# Print every contract-creation address from this run, one per line, in
# broadcast order. A script deploying a single contract yields one line
# (existing callers take that line as "the" address); a script deploying
# several related contracts (e.g. one ERC needing multiple wired-together
# contracts) yields one line per contract, in the order `new` was called.
jq -r '.transactions[] | select(.contractAddress != null) | .contractAddress' \
  "broadcast/${CONTRACT_NAME}.s.sol/${CHAIN_ID}/run-latest.json"
