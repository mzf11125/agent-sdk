from __future__ import annotations

from dataclasses import dataclass
from typing import Any

from eth_account.signers.local import LocalAccount
from web3 import Web3
from web3.logs import DISCARD
from web3.middleware import SignAndSendRawMiddlewareBuilder

from .abi import AGENT_WORKFLOW_ABI


@dataclass(frozen=True)
class AgentTask:
    stage: int
    task_seq: int
    input_hash: str
    input: bytes
    timestamp: int
    expires_at: int
    prev_reply_hashes: list[str]
    workflow_run_id: str


@dataclass(frozen=True)
class AgentReply:
    output_hash: str
    output: bytes
    timestamp: int
    replier: str
    prev_task_hashes: list[str]
    workflow_run_id: str


@dataclass(frozen=True)
class RunResult:
    status: int
    final_task_hash: str
    completed_at: int


@dataclass(frozen=True)
class TaskResult:
    task: AgentTask
    proven: bool


@dataclass(frozen=True)
class ReplyResult:
    reply: AgentReply
    verifier: str
    proven: bool
    verification_digest: str


@dataclass(frozen=True)
class NewTaskEvent:
    workflow_run_id: str
    stage: int
    task_hash: str


class AgentWorkflowClient:
    """Client for the ERC-8301 IAgentWorkflow contract."""

    def __init__(self, rpc_url: str, address: str, account: LocalAccount):
        self._w3 = Web3(Web3.HTTPProvider(rpc_url))
        self._w3.middleware_onion.add(SignAndSendRawMiddlewareBuilder.build(account))
        self._w3.eth.default_account = account.address
        self._contract = self._w3.eth.contract(
            address=Web3.to_checksum_address(address), abi=AGENT_WORKFLOW_ABI
        )

    def run(self, input_hash: str, input_data: bytes, expires_at: int) -> NewTaskEvent:
        """Start a new workflow run.

        Args:
            input_hash: keccak256 of the input (0x-prefixed hex).
            input_data: Input plaintext; MAY be empty (b"").
            expires_at: Unix timestamp after which the initial task expires.

        Returns:
            NewTaskEvent with workflow_run_id, task_hash, and stage.
        """
        tx_hash = self._contract.functions.run(
            Web3.to_bytes(hexstr=input_hash),
            input_data,
            expires_at,
        ).transact()
        receipt = self._w3.eth.wait_for_transaction_receipt(tx_hash)
        events = self._contract.events.NewAgentTask().process_receipt(receipt, errors=DISCARD)
        if not events:
            raise RuntimeError("run: NewAgentTask event not found in transaction receipt")
        args = events[0]["args"]
        return NewTaskEvent(
            workflow_run_id=_to_hex_str(args["workflowRunId"]),
            stage=args["stage"],
            task_hash=_to_hex_str(args["taskHash"]),
        )

    def result(self, workflow_run_id: str) -> RunResult:
        """Query the result of a run.

        Args:
            workflow_run_id: The run identifier (0x-prefixed hex).

        Returns:
            RunResult with status, final_task_hash, and completed_at.
        """
        status, final_task_hash, completed_at = self._contract.functions.result(
            Web3.to_bytes(hexstr=workflow_run_id),
        ).call()
        return RunResult(
            status=status,
            final_task_hash=_to_hex_str(final_task_hash),
            completed_at=completed_at,
        )

    def get_task(self, task_hash: str) -> TaskResult:
        """Returns the stored AgentTask and its proven status.

        Args:
            task_hash: The task hash (0x-prefixed hex).

        Returns:
            TaskResult with task and proven flag.
        """
        raw_task, proven = self._contract.functions.getAgentTask(
            Web3.to_bytes(hexstr=task_hash),
        ).call()
        task = _raw_to_task(raw_task)
        return TaskResult(task=task, proven=proven)

    def get_reply(self, reply_hash: str) -> ReplyResult:
        """Returns the stored AgentReply and its verification status.

        Args:
            reply_hash: The reply hash (0x-prefixed hex).

        Returns:
            ReplyResult with reply, verifier, proven, and verification_digest.
        """
        raw_reply, verifier, proven, verification_digest = self._contract.functions.getAgentReply(
            Web3.to_bytes(hexstr=reply_hash),
        ).call()
        reply = _raw_to_reply(raw_reply)
        return ReplyResult(
            reply=reply,
            verifier=verifier,
            proven=proven,
            verification_digest=_to_hex_str(verification_digest),
        )

    def on_agent_reply(self, reply: AgentReply) -> None:
        """Agent submits a reply to a dispatched task."""
        raw = _reply_to_raw(reply)
        self._send("onAgentReply", raw)

    def on_agent_prove(self, reply_hashes: list[str], proof: bytes) -> None:
        """Submit a cryptographic proof covering one or more anchored replies."""
        raw_hashes = [Web3.to_bytes(hexstr=h) for h in reply_hashes]
        self._send("onAgentProve", raw_hashes, proof)

    def _send(self, function_name: str, *args: Any) -> None:
        tx_hash = getattr(self._contract.functions, function_name)(*args).transact()
        self._w3.eth.wait_for_transaction_receipt(tx_hash)


# ── Helper functions ──────────────────────────────────────────────────────


def _to_hex_str(value: bytes | str) -> str:
    """Convert bytes to a 0x-prefixed hex string (idempotent for hex strings)."""
    if isinstance(value, str):
        return value
    return "0x" + value.hex()


def _raw_to_task(raw: tuple) -> AgentTask:
    """Convert a raw tuple from the contract into an AgentTask."""
    return AgentTask(
        stage=raw[0],
        task_seq=raw[1],
        input_hash=_to_hex_str(raw[2]),
        input=raw[3],
        timestamp=raw[4],
        expires_at=raw[5],
        prev_reply_hashes=[_to_hex_str(h) for h in raw[6]],
        workflow_run_id=_to_hex_str(raw[7]),
    )


def _raw_to_reply(raw: tuple) -> AgentReply:
    """Convert a raw tuple from the contract into an AgentReply."""
    return AgentReply(
        output_hash=_to_hex_str(raw[0]),
        output=raw[1],
        timestamp=raw[2],
        replier=raw[3],
        prev_task_hashes=[_to_hex_str(h) for h in raw[4]],
        workflow_run_id=_to_hex_str(raw[5]),
    )


def _reply_to_raw(reply: AgentReply) -> tuple:
    """Convert an AgentReply to a raw tuple for contract submission."""
    return (
        Web3.to_bytes(hexstr=reply.output_hash),
        reply.output,
        reply.timestamp,
        Web3.to_checksum_address(reply.replier),
        [Web3.to_bytes(hexstr=h) for h in reply.prev_task_hashes],
        Web3.to_bytes(hexstr=reply.workflow_run_id),
    )
