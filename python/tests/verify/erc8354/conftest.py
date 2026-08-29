"""Pytest fixtures for ERC-8354 integration tests."""

from pathlib import Path

import pytest
from eth_account import Account
from web3 import Web3

TESTKIT_DIR = Path(__file__).resolve().parents[4] / "testkit"
ANVIL_RPC_URL = "http://127.0.0.1:8545"

DOMAIN_ID = Web3.keccak(text="test-domain")
POLICY_ROOT = Web3.keccak(text="root-v1")
PROGRAM_KEY = Web3.keccak(text="interpreter-vkey")

REGISTRY_ADMIN_ABI = [
    {
        "type": "function",
        "name": "registerDomain",
        "stateMutability": "nonpayable",
        "inputs": [
            {"name": "domainId", "type": "bytes32"},
            {"name": "registrar", "type": "address"},
            {"name": "verifier", "type": "address"},
            {"name": "programKey", "type": "bytes32"},
            {"name": "maxRootAge", "type": "uint64"},
        ],
        "outputs": [],
    },
    {
        "type": "function",
        "name": "updateRoot",
        "stateMutability": "nonpayable",
        "inputs": [
            {"name": "domainId", "type": "bytes32"},
            {"name": "newRoot", "type": "bytes32"},
        ],
        "outputs": [],
    },
]


def _read_account(index):
    import json

    with open(TESTKIT_DIR / ".anvil-accounts.json") as f:
        data = json.load(f)
    return data["accounts"][index]


@pytest.fixture(scope="module")
def erc8354():
    """Deploy the ERC-8354 mocks and set up a registered domain and root.

    Returns a dict with verifier, registry, and guard addresses.
    """
    import subprocess

    result = subprocess.run(
        [str(TESTKIT_DIR / "scripts" / "deploy.sh"), "verify/ERC8354", "DeployERC8354"],
        capture_output=True,
        text=True,
        check=True,
        cwd=str(TESTKIT_DIR),
    )
    verifier, registry, guard = [line for line in result.stdout.strip().splitlines() if line]
    verifier = Web3.to_checksum_address(verifier)
    registry = Web3.to_checksum_address(registry)
    guard = Web3.to_checksum_address(guard)

    admin = Account.from_key(_read_account(0)["privateKey"])
    w3 = Web3(Web3.HTTPProvider(ANVIL_RPC_URL))
    w3.eth.default_account = admin.address
    registry_contract = w3.eth.contract(address=registry, abi=REGISTRY_ADMIN_ABI)

    tx = registry_contract.functions.registerDomain(
        DOMAIN_ID,
        "0x000000000000000000000000000000000000a11c",
        verifier,
        PROGRAM_KEY,
        3600,
    ).transact({"from": admin.address})
    # Wait for the domain to exist before the root update, which depends on it.
    w3.eth.wait_for_transaction_receipt(tx)
    tx = registry_contract.functions.updateRoot(DOMAIN_ID, POLICY_ROOT).transact(
        {"from": admin.address}
    )
    w3.eth.wait_for_transaction_receipt(tx)

    return {"verifier": verifier, "registry": registry, "guard": guard}


@pytest.fixture(scope="module")
def executor_account():
    return Account.from_key(_read_account(1)["privateKey"])
