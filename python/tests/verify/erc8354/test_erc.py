"""Integration tests for ERC-8354 Confidential Policy Verdicts (Layer 1)."""

import pytest
from eth_account import Account
from web3 import Web3

from agent_sdk.verify.erc8354.client import (
    ConfidentialPolicyVerdictClient,
    PolicyDomainRegistryClient,
    Verdict,
)
from agent_sdk.verify.erc8354.recompute import compute_action_commitment

from .conftest import DOMAIN_ID, POLICY_ROOT, _read_account

ANVIL_RPC_URL = "http://127.0.0.1:8545"

DOMAIN = {
    "name": "ConfidentialPolicyVerdict",
    "version": "1",
    "chainId": 31337,
    "verifyingContract": "0x0000000000000000000000000000000000000000",  # replaced per test
}

VERDICT_TYPES = {
    "EIP712Domain": [
        {"name": "name", "type": "string"},
        {"name": "version", "type": "string"},
        {"name": "chainId", "type": "uint256"},
        {"name": "verifyingContract", "type": "address"},
    ],
    "Verdict": [
        {"name": "agentId", "type": "uint256"},
        {"name": "domainId", "type": "bytes32"},
        {"name": "policyRoot", "type": "bytes32"},
        {"name": "actionCommitment", "type": "bytes32"},
        {"name": "executor", "type": "address"},
        {"name": "expiry", "type": "uint64"},
        {"name": "nullifier", "type": "bytes32"},
        {"name": "decision", "type": "uint8"},
        {"name": "policyKind", "type": "uint8"},
    ],
}

MOCK_VERIFIER_ABI = [
    {
        "type": "function",
        "name": "setResult",
        "stateMutability": "nonpayable",
        "inputs": [{"name": "r", "type": "bool"}],
        "outputs": [],
    }
]


def _verdict(executor_address: str, nullifier: bytes, commitment: bytes) -> Verdict:
    return Verdict(
        agent_id=1,
        domain_id=DOMAIN_ID,
        policy_root=POLICY_ROOT,
        action_commitment=commitment,
        executor=executor_address,
        expiry=4_000_000_000,
        nullifier=nullifier,
        decision=1,
        policy_kind=0,
    )


def _commitment() -> bytes:
    commitment_hex = compute_action_commitment(
        31337,
        DOMAIN_ID,
        1,
        "0x0000000000000000000000000000000000000001",
        0,
        "0x",
        0,
    )
    return bytes.fromhex(commitment_hex[2:])


def _message_dict(verdict: Verdict) -> dict:
    return {
        "agentId": verdict.agent_id,
        "domainId": verdict.domain_id,
        "policyRoot": verdict.policy_root,
        "actionCommitment": verdict.action_commitment,
        "executor": verdict.executor,
        "expiry": verdict.expiry,
        "nullifier": verdict.nullifier,
        "decision": verdict.decision,
        "policyKind": verdict.policy_kind,
    }


def _executor_auth(guard_address: str, executor_key: str, verdict: Verdict) -> bytes:
    domain = dict(DOMAIN)
    domain["verifyingContract"] = guard_address
    full_message = {
        "types": VERDICT_TYPES,
        "primaryType": "Verdict",
        "domain": domain,
        "message": _message_dict(verdict),
    }
    signed = Account.sign_typed_data(executor_key, full_message=full_message)
    return signed.signature


def test_supports_interface(erc8354, executor_account):
    client = ConfidentialPolicyVerdictClient(
        ANVIL_RPC_URL, erc8354["guard"], executor_account
    )
    assert client.supports_interface() is True


def test_registry_reads(erc8354):
    client = PolicyDomainRegistryClient(ANVIL_RPC_URL, erc8354["registry"])
    domain = client.domain(DOMAIN_ID)
    assert domain.active is True
    assert domain.verifier.lower() == erc8354["verifier"].lower()
    assert client.is_root_acceptable(DOMAIN_ID, POLICY_ROOT) is True


def test_verify(erc8354, executor_account):
    client = ConfidentialPolicyVerdictClient(
        ANVIL_RPC_URL, erc8354["guard"], executor_account
    )
    verdict = _verdict(executor_account.address, Web3.keccak(text="nf-verify"), _commitment())
    assert client.verify(verdict, b"proof") is True


def test_consume_and_replay(erc8354, executor_account):
    client = ConfidentialPolicyVerdictClient(
        ANVIL_RPC_URL, erc8354["guard"], executor_account
    )
    verdict = _verdict(
        executor_account.address, Web3.keccak(text="nf-consume"), _commitment()
    )
    receipt = client.consume(verdict, b"proof")
    assert receipt.get("status") == 1
    assert client.is_consumed(DOMAIN_ID, verdict.nullifier) is True

    with pytest.raises(Exception):
        client.consume(verdict, b"proof")


def test_consume_relayed(erc8354, executor_account):
    # The executor (account 1) authorizes a relayer (account 0) by signing the
    # EIP-712 verdict digest. The relayer submits the transaction.
    relayer_account = Account.from_key(_read_account(0)["privateKey"])
    client = ConfidentialPolicyVerdictClient(
        ANVIL_RPC_URL, erc8354["guard"], relayer_account
    )
    verdict = _verdict(
        executor_account.address, Web3.keccak(text="nf-relayed"), _commitment()
    )
    executor_auth = _executor_auth(
        erc8354["guard"], _read_account(1)["privateKey"], verdict
    )
    receipt = client.consume_relayed(verdict, b"proof", executor_auth)
    assert receipt.get("status") == 1
    assert client.is_consumed(DOMAIN_ID, verdict.nullifier) is True


def test_rejects_bad_executor_signature(erc8354, executor_account):
    relayer_account = Account.from_key(_read_account(0)["privateKey"])
    client = ConfidentialPolicyVerdictClient(
        ANVIL_RPC_URL, erc8354["guard"], relayer_account
    )
    verdict = _verdict(
        executor_account.address, Web3.keccak(text="nf-bad-sig"), _commitment()
    )
    with pytest.raises(Exception):
        client.consume_relayed(verdict, b"proof", b"not-a-signature")


def test_rejects_invalid_proof(erc8354, executor_account):
    admin_account = Account.from_key(_read_account(0)["privateKey"])
    w3 = Web3(Web3.HTTPProvider(ANVIL_RPC_URL))
    w3.eth.default_account = admin_account.address
    verifier_contract = w3.eth.contract(address=erc8354["verifier"], abi=MOCK_VERIFIER_ABI)
    tx = verifier_contract.functions.setResult(False).transact(
        {"from": admin_account.address}
    )
    w3.eth.wait_for_transaction_receipt(tx)

    client = ConfidentialPolicyVerdictClient(
        ANVIL_RPC_URL, erc8354["guard"], executor_account
    )
    verdict = _verdict(
        executor_account.address, Web3.keccak(text="nf-invalid-proof"), _commitment()
    )
    with pytest.raises(Exception):
        client.consume(verdict, b"proof")
