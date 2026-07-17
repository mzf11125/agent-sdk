import time

import pytest
from eth_account import Account
from eth_account.messages import encode_defunct
from web3 import Web3
from web3.exceptions import ContractLogicError

from agent_sdk.settlement.erc8203.client import ConsultEscrowClient, JobStatus
from agent_sdk.settlement.erc8203.recompute import compute_verdict_hash


def _make_client(deploy_contract, anvil_rpc_url, anvil_acct_fn):
    contract_address = Web3.to_checksum_address(
        deploy_contract("settlement/ERC8203", "DeployERC8203")
    )
    # Account #1 acts as the consumer.
    account = Account.from_key(anvil_acct_fn(1)["privateKey"])
    client = ConsultEscrowClient(anvil_rpc_url, contract_address, account)
    return client, contract_address, anvil_acct_fn


def _compute_job_id(label: str) -> str:
    return "0x" + Web3.keccak(text=label).hex()


class TestConsultEscrowClient:
    """Integration tests for ERC-8203 ConsultEscrow client."""

    def test_open_and_get_job(self, deploy_contract, anvil_rpc_url, anvil_account):
        client, _, acct_fn = _make_client(deploy_contract, anvil_rpc_url, anvil_account)
        provider = acct_fn(2)["address"]
        attestor = acct_fn(0)["address"]
        job_id = _compute_job_id("test-open-1")
        deadline = int(time.time()) + 3600
        value = Web3.to_wei(1, "ether")

        receipt = client.open(job_id, provider, attestor, deadline, value)

        opened_events = client._get_opened_events(receipt)
        assert len(opened_events) == 1
        assert opened_events[0].job_id == job_id
        assert opened_events[0].provider.lower() == provider.lower()
        assert opened_events[0].attestor.lower() == attestor.lower()
        assert opened_events[0].amount == value
        assert opened_events[0].deadline == deadline

        job = client.get_job(job_id)
        assert job.consumer.lower() == acct_fn(1)["address"].lower()
        assert job.provider.lower() == provider.lower()
        assert job.attestor.lower() == attestor.lower()
        assert job.amount == value
        assert job.deadline == deadline
        assert job.status == JobStatus.OPEN

    def test_revert_on_duplicate_job(self, deploy_contract, anvil_rpc_url, anvil_account):
        client, _, acct_fn = _make_client(deploy_contract, anvil_rpc_url, anvil_account)
        provider = acct_fn(2)["address"]
        attestor = acct_fn(0)["address"]
        job_id = _compute_job_id("test-duplicate")
        deadline = int(time.time()) + 3600
        value = Web3.to_wei(1, "ether")

        client.open(job_id, provider, attestor, deadline, value)

        with pytest.raises(ContractLogicError):
            client.open(job_id, provider, attestor, deadline, value)

    def test_revert_on_zero_value(self, deploy_contract, anvil_rpc_url, anvil_account):
        client, _, acct_fn = _make_client(deploy_contract, anvil_rpc_url, anvil_account)
        provider = acct_fn(2)["address"]
        attestor = acct_fn(0)["address"]
        job_id = _compute_job_id("test-zero-value")
        deadline = int(time.time()) + 3600

        with pytest.raises(ContractLogicError):
            client.open(job_id, provider, attestor, deadline, 0)

    def test_release_happy_path(self, deploy_contract, anvil_rpc_url, anvil_account):
        client, _, acct_fn = _make_client(deploy_contract, anvil_rpc_url, anvil_account)
        provider = acct_fn(2)["address"]
        attestor_key = acct_fn(0)["privateKey"]
        attestor_addr = acct_fn(0)["address"]
        job_id = _compute_job_id("test-release")
        deadline = int(time.time()) + 3600
        value = Web3.to_wei(1, "ether")
        result_text = "Task completed successfully."

        client.open(job_id, provider, attestor_addr, deadline, value)

        result_hash = Web3.to_hex(Web3.keccak(text=result_text))
        commitment_hash = compute_verdict_hash(job_id, result_text)

        # Sign with attestor key
        attestor_acct = Account.from_key(attestor_key)
        signable_message = encode_defunct(hexstr=commitment_hash)
        signature = attestor_acct.sign_message(signable_message)

        receipt = client.release(job_id, result_hash, bytes(signature.signature))

        released_events = client._get_released_events(receipt)
        assert len(released_events) == 1
        assert released_events[0].job_id == job_id
        assert released_events[0].result_hash == result_hash
        assert released_events[0].commitment_hash == commitment_hash
        assert released_events[0].provider.lower() == provider.lower()
        assert released_events[0].amount == value

        job = client.get_job(job_id)
        assert job.status == JobStatus.RELEASED

    def test_release_rejects_wrong_signer(self, deploy_contract, anvil_rpc_url, anvil_account):
        client, _, acct_fn = _make_client(deploy_contract, anvil_rpc_url, anvil_account)
        provider = acct_fn(2)["address"]
        attestor_addr = acct_fn(0)["address"]
        job_id = _compute_job_id("test-wrong-sig")
        deadline = int(time.time()) + 3600
        value = Web3.to_wei(1, "ether")
        result_text = "Tampered result."

        client.open(job_id, provider, attestor_addr, deadline, value)

        result_hash = Web3.to_hex(Web3.keccak(text=result_text))
        commitment_hash = compute_verdict_hash(job_id, result_text)

        # Sign with WRONG key (account #1) — not the attestor
        wrong_acct = Account.from_key(acct_fn(1)["privateKey"])
        wrong_signable = encode_defunct(hexstr=commitment_hash)
        wrong_sig = wrong_acct.sign_message(wrong_signable)

        with pytest.raises(ContractLogicError):
            client.release(job_id, result_hash, bytes(wrong_sig.signature))

    def test_refund_reverts_before_deadline(self, deploy_contract, anvil_rpc_url, anvil_account):
        client, _, acct_fn = _make_client(deploy_contract, anvil_rpc_url, anvil_account)
        provider = acct_fn(2)["address"]
        attestor = acct_fn(0)["address"]
        job_id = _compute_job_id("test-refund-early")
        deadline = int(time.time()) + 3600
        value = Web3.to_wei(1, "ether")

        client.open(job_id, provider, attestor, deadline, value)

        with pytest.raises(ContractLogicError):
            client.refund(job_id)

    def test_verify_accepts_genuine_commitment(self, deploy_contract, anvil_rpc_url, anvil_account):
        client, _, _ = _make_client(deploy_contract, anvil_rpc_url, anvil_account)
        job_id = "0xbc01b40fe7a3509f35470053d4bc1844d50c9782546cf0fc11154adcb90caa56"
        result_text = "No intermediaries required, cryptographic verification only."
        commitment_hash = compute_verdict_hash(job_id, result_text)

        assert client.verify(commitment_hash, job_id, result_text) is True

    def test_verify_rejects_tampered_hash(self, deploy_contract, anvil_rpc_url, anvil_account):
        client, _, _ = _make_client(deploy_contract, anvil_rpc_url, anvil_account)
        job_id = "0xbc01b40fe7a3509f35470053d4bc1844d50c9782546cf0fc11154adcb90caa56"
        result_text = "No intermediaries required, cryptographic verification only."
        tampered_hash = "0x" + "00" * 32

        assert client.verify(tampered_hash, job_id, result_text) is False

    def test_verify_detects_tampered_text(self, deploy_contract, anvil_rpc_url, anvil_account):
        client, _, _ = _make_client(deploy_contract, anvil_rpc_url, anvil_account)
        job_id = "0xbc01b40fe7a3509f35470053d4bc1844d50c9782546cf0fc11154adcb90caa56"
        original_text = "No intermediaries required, cryptographic verification only."
        tampered_text = "Tampered result text."
        commitment_hash = compute_verdict_hash(job_id, original_text)

        assert client.verify(commitment_hash, job_id, tampered_text) is False
