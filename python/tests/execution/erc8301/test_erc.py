from __future__ import annotations

import time

import pytest
from eth_account import Account
from web3 import Web3
from web3.exceptions import ContractLogicError

from agent_sdk.execution.erc8301.client import AgentWorkflowClient, AgentReply


def _keccak_hex(text: str) -> str:
    """Return 0x-prefixed keccak256 hex string of text."""
    return Web3.to_hex(Web3.keccak(text=text))


@pytest.fixture
def client(deploy_contract, anvil_rpc_url, anvil_account):
    contract_address = Web3.to_checksum_address(
        deploy_contract("execution/ERC8301", "DeployERC8301")
    )
    account = Account.from_key(anvil_account(1)["privateKey"])
    return AgentWorkflowClient(anvil_rpc_url, contract_address, account), account, contract_address


def test_starts_workflow_run(client):
    cl, _, _ = client
    input_hash = _keccak_hex("test input")
    expires_at = int(time.time()) + 1000

    result = cl.run(input_hash, b"test input", expires_at)

    assert result.workflow_run_id is not None
    assert result.workflow_run_id != "0x" + "00" * 32
    assert result.task_hash is not None
    assert result.stage == 1


def test_retrieves_task_after_run(client):
    cl, _, _ = client
    input_hash = _keccak_hex("test retrieval")
    expires_at = int(time.time()) + 1000

    event = cl.run(input_hash, b"test retrieval", expires_at)
    task_result = cl.get_task(event.task_hash)

    assert task_result.task.stage == 1
    assert task_result.task.input_hash == input_hash
    assert task_result.task.workflow_run_id == event.workflow_run_id
    assert task_result.proven is True  # initial task with empty prevReplyHashes is proven


def test_retrieves_run_result(client):
    cl, _, _ = client
    input_hash = _keccak_hex("test result")
    expires_at = int(time.time()) + 1000

    event = cl.run(input_hash, b"test result", expires_at)
    run_result = cl.result(event.workflow_run_id)

    assert run_result.status == 0  # Pending


def test_submits_reply_and_retrieves_it(client):
    cl, account, _ = client
    input_hash = _keccak_hex("test reply")
    expires_at = int(time.time()) + 1000

    event = cl.run(input_hash, b"test reply", expires_at)
    now = int(time.time())
    output_hash = _keccak_hex("reply output")

    reply = AgentReply(
        output_hash=output_hash,
        output=b"reply output",
        timestamp=now,
        replier=account.address,
        prev_task_hashes=[],
        workflow_run_id=event.workflow_run_id,
    )

    cl.on_agent_reply(reply)

    # Compute the expected replyHash
    from agent_sdk.execution.erc8301.recompute import compute_reply_hash

    reply_hash = compute_reply_hash(
        output_hash, now, account.address, "0x", event.workflow_run_id,
    )

    reply_result = cl.get_reply(reply_hash)
    assert reply_result.reply.output_hash == output_hash
    assert reply_result.reply.replier.lower() == account.address.lower()
    assert reply_result.reply.workflow_run_id == event.workflow_run_id
    assert reply_result.verifier == "0x0000000000000000000000000000000000000000"
    assert reply_result.proven is False


def test_marks_reply_as_proven(client):
    cl, account, _ = client
    input_hash = _keccak_hex("test prove")
    expires_at = int(time.time()) + 1000

    event = cl.run(input_hash, b"test prove", expires_at)
    now = int(time.time())
    output_hash = _keccak_hex("prove output")

    reply = AgentReply(
        output_hash=output_hash,
        output=b"prove output",
        timestamp=now,
        replier=account.address,
        prev_task_hashes=[],
        workflow_run_id=event.workflow_run_id,
    )

    cl.on_agent_reply(reply)

    from agent_sdk.execution.erc8301.recompute import compute_reply_hash

    reply_hash = compute_reply_hash(
        output_hash, now, account.address, "0x", event.workflow_run_id,
    )

    cl.on_agent_prove([reply_hash], b"proof")

    reply_result = cl.get_reply(reply_hash)
    assert reply_result.proven is True
    assert reply_result.verifier.lower() == account.address.lower()
    assert reply_result.verification_digest is not None


def test_reverts_on_nonexistent_task(client):
    cl, _, _ = client
    fake_hash = "0x" + "00" * 31 + "01"

    with pytest.raises(ContractLogicError):
        cl.get_task(fake_hash)


def test_reverts_on_nonexistent_reply(client):
    cl, _, _ = client
    fake_hash = "0x" + "00" * 31 + "02"

    with pytest.raises(ContractLogicError):
        cl.get_reply(fake_hash)
