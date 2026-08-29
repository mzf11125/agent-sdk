"""Recompute tests for ERC-8354 Confidential Agent Policy Verdicts (Layer 2)."""

import json
import pathlib

import pytest

from agent_sdk.verify.erc8354.client import Verdict
from agent_sdk.verify.erc8354.recompute import (
    MECHANISM_ZK_SECRET_POLICY,
    compute_action_commitment,
    compute_verdict_digest,
)

DOMAIN_ID = "0x34a63641b78652cdd53505da4f32cac6058bd148e3ff543f39f75997a89c2815"
ACTION_COMMITMENT = "0xcc8e5dc414db5ed2340be02c3d7fdc725fe5f1463b382a7ed13f8036a4a0b7b1"


def test_action_commitment_golden_vector():
    result = compute_action_commitment(
        chain_id=31337,
        domain_id=DOMAIN_ID,
        agent_id=1,
        target="0x0000000000000000000000000000000000000001",
        value=0,
        call_data="0x",
        action_nonce=0,
    )
    assert result == ACTION_COMMITMENT


def test_action_commitment_empty_call_data_hashes_keccak_empty():
    result = compute_action_commitment(
        chain_id=31337,
        domain_id=DOMAIN_ID,
        agent_id=1,
        target="0x0000000000000000000000000000000000000001",
        value=0,
        call_data="0x",
        action_nonce=0,
    )
    assert result != "0x0000000000000000000000000000000000000000000000000000000000000000"


def test_action_commitment_deterministic():
    args = dict(
        chain_id=31337,
        domain_id=DOMAIN_ID,
        agent_id=1,
        target="0x0000000000000000000000000000000000000001",
        value=0,
        call_data="0x",
        action_nonce=0,
    )
    assert compute_action_commitment(**args) == compute_action_commitment(**args)


def test_action_commitment_changes_with_nonce():
    args = dict(
        chain_id=31337,
        domain_id=DOMAIN_ID,
        agent_id=1,
        target="0x0000000000000000000000000000000000000001",
        value=0,
        call_data="0x",
    )
    a = compute_action_commitment(action_nonce=0, **args)
    b = compute_action_commitment(action_nonce=1, **args)
    assert a != b


def test_verdict_digest_golden_vector():
    result = compute_verdict_digest(
        {
            "agent_id": 1,
            "domain_id": DOMAIN_ID,
            "policy_root": DOMAIN_ID,
            "action_commitment": ACTION_COMMITMENT,
            "executor": "0x0000000000000000000000000000000000000002",
            "expiry": 2000000000,
            "nullifier": "0x6e47261c83f90eed41cda2b00caad094c33daa0a09fec22396b3e2bfe5e222b2",
            "decision": 1,
            "policy_kind": 0,
        },
        chain_id=31337,
        verifying_contract="0x0000000000000000000000000000000000000003",
    )
    assert result == "0xf2345f63ba9e78a068eb4f74640e6543289010540b457d8016771175ad460f32"


def test_verdict_digest_accepts_exported_dataclass():
    # The exported Verdict dataclass must compose with compute_verdict_digest
    # directly, without forcing callers to convert to a dict first.
    verdict = Verdict(
        agent_id=1,
        domain_id=bytes.fromhex(DOMAIN_ID[2:]),
        policy_root=bytes.fromhex(DOMAIN_ID[2:]),
        action_commitment=bytes.fromhex(ACTION_COMMITMENT[2:]),
        executor="0x0000000000000000000000000000000000000002",
        expiry=2000000000,
        nullifier=bytes.fromhex(
            "6e47261c83f90eed41cda2b00caad094c33daa0a09fec22396b3e2bfe5e222b2"
        ),
        decision=1,
        policy_kind=0,
    )
    result = compute_verdict_digest(
        verdict,
        chain_id=31337,
        verifying_contract="0x0000000000000000000000000000000000000003",
    )
    assert result == "0xf2345f63ba9e78a068eb4f74640e6543289010540b457d8016771175ad460f32"


def test_verdict_digest_depends_on_verifying_contract():
    verdict = {
        "agent_id": 1,
        "domain_id": DOMAIN_ID,
        "policy_root": DOMAIN_ID,
        "action_commitment": ACTION_COMMITMENT,
        "executor": "0x0000000000000000000000000000000000000002",
        "expiry": 2000000000,
        "nullifier": "0x6e47261c83f90eed41cda2b00caad094c33daa0a09fec22396b3e2bfe5e222b2",
        "decision": 1,
        "policy_kind": 0,
    }
    a = compute_verdict_digest(verdict, chain_id=31337, verifying_contract="0x0000000000000000000000000000000000000003")
    b = compute_verdict_digest(verdict, chain_id=31337, verifying_contract="0x0000000000000000000000000000000000000004")
    assert a != b


def test_mechanism_constant():
    assert MECHANISM_ZK_SECRET_POLICY == "0xa843829a78c66c29679817606d0c8a9fa26575b6c2ed0f9f97079d7c46577ac6"


VECTORS = pathlib.Path(__file__).resolve().parents[4] / "testkit" / "vectors" / "erc8354-verdict.vectors.json"


def _load_vectors():
    if not VECTORS.exists():
        return []
    return json.loads(VECTORS.read_text(encoding="utf-8"))["vectors"]


@pytest.mark.parametrize("v", _load_vectors(), ids=lambda v: v.get("step", ""))
def test_golden_vector(v):
    if v["step"] == "8354/action-commitment":
        inputs = v["inputs"]
        assert compute_action_commitment(
            chain_id=int(inputs["chainId"]),
            domain_id=inputs["domainId"],
            agent_id=int(inputs["agentId"]),
            target=inputs["target"],
            value=int(inputs["value"]),
            call_data=inputs["callData"],
            action_nonce=int(inputs["actionNonce"]),
        ) == v["expected"]
    elif v["step"] == "8354/verdict-digest":
        inputs = v["inputs"]
        assert compute_verdict_digest(
            {
                "agent_id": int(inputs["agentId"]),
                "domain_id": inputs["domainId"],
                "policy_root": inputs["policyRoot"],
                "action_commitment": inputs["actionCommitment"],
                "executor": inputs["executor"],
                "expiry": int(inputs["expiry"]),
                "nullifier": inputs["nullifier"],
                "decision": int(inputs["decision"]),
                "policy_kind": int(inputs["policyKind"]),
            },
            chain_id=int(inputs["chainId"]),
            verifying_contract=inputs["verifyingContract"],
        ) == v["expected"]
    else:
        pytest.fail(f"unknown step {v['step']}, a vector exists that no function covers")
