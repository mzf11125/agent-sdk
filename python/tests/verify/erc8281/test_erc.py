"""Integration tests for ERC-8281 Observation Commitment client (Layer 1)."""

from eth_account import Account
from eth_utils import keccak

from agent_sdk.verify.erc8281.client import ObservationCommitmentClient
from agent_sdk.verify.erc8281.recompute import compute_observation_digest


def test_supports_observation_commitment(deploy_erc8281, anvil_rpc_url, anvil_account):
    """Supports the observation commitment interface."""
    addr = deploy_erc8281
    account = Account.from_key(anvil_account(1)["privateKey"])
    client = ObservationCommitmentClient(anvil_rpc_url, addr, account)
    assert client.supports_observation_commitment() is True


def test_record_and_parse_event(deploy_erc8281, anvil_rpc_url, anvil_account):
    """Records a digest and emits Recorded event."""
    addr = deploy_erc8281
    account = Account.from_key(anvil_account(1)["privateKey"])
    client = ObservationCommitmentClient(anvil_rpc_url, addr, account)

    digest = keccak(b"test-observation")
    receipt = client.record(digest)

    assert receipt.get("status") == 1  # success

    event = client.parse_recorded_event(receipt)
    assert event.digest == digest
    assert event.committer.lower() == account.address.lower()


def test_verify_via_recompute(deploy_erc8281, anvil_rpc_url):
    """Verifies via recompute: digest matches keccak256(observation)."""
    observation = bytes.fromhex("deadbeef")
    digest = compute_observation_digest(observation)
    expected = keccak(observation)
    assert digest == expected


def test_record_multiple_digests(deploy_erc8281, anvil_rpc_url, anvil_account):
    """Records multiple digests with different values."""
    addr = deploy_erc8281
    account = Account.from_key(anvil_account(1)["privateKey"])
    client = ObservationCommitmentClient(anvil_rpc_url, addr, account)

    digest1 = keccak(b"obs-1")
    digest2 = keccak(b"obs-2")

    receipt1 = client.record(digest1)
    receipt2 = client.record(digest2)

    assert receipt1["status"] == 1
    assert receipt2["status"] == 1

    event1 = client.parse_recorded_event(receipt1)
    event2 = client.parse_recorded_event(receipt2)
    assert event1.digest == digest1
    assert event2.digest == digest2
