from __future__ import annotations

from dataclasses import dataclass
from enum import IntEnum

from eth_account.signers.local import LocalAccount
from web3 import Web3
from web3.logs import DISCARD
from web3.middleware import SignAndSendRawMiddlewareBuilder

from agent_sdk.settlement.erc8203.recompute import compute_verdict_hash

from .abi import CONSULT_ESCROW_ABI


class JobStatus(IntEnum):
    NONE = 0
    OPEN = 1
    RELEASED = 2
    REFUNDED = 3


@dataclass(frozen=True)
class Job:
    consumer: str
    provider: str
    attestor: str
    amount: int
    deadline: int
    status: int


@dataclass(frozen=True)
class OpenedEvent:
    job_id: str
    consumer: str
    provider: str
    attestor: str
    amount: int
    deadline: int


@dataclass(frozen=True)
class ReleasedEvent:
    job_id: str
    result_hash: str
    commitment_hash: str
    provider: str
    amount: int


@dataclass(frozen=True)
class RefundedEvent:
    job_id: str
    consumer: str
    amount: int


class ConsultEscrowClient:
    """Client for the ERC-8203 ConsultEscrow settlement contract."""

    def __init__(self, rpc_url: str, address: str, account: LocalAccount):
        self._w3 = Web3(Web3.HTTPProvider(rpc_url))
        self._w3.middleware_onion.add(SignAndSendRawMiddlewareBuilder.build(account))
        self._w3.eth.default_account = account.address
        self._contract = self._w3.eth.contract(
            address=Web3.to_checksum_address(address), abi=CONSULT_ESCROW_ABI
        )

    def open(self, job_id: str, provider: str, attestor: str, deadline: int, value: int) -> bytes:
        """Open a new escrow job. Sends value as msg.value."""
        tx_hash = self._contract.functions.open(
            Web3.to_bytes(hexstr=job_id),
            Web3.to_checksum_address(provider),
            Web3.to_checksum_address(attestor),
            deadline,
        ).transact({"value": value})
        return self._w3.eth.wait_for_transaction_receipt(tx_hash)

    def release(self, job_id: str, result_hash: str, signature: bytes) -> bytes:
        """Release escrowed funds to the provider."""
        tx_hash = self._contract.functions.release(
            Web3.to_bytes(hexstr=job_id),
            Web3.to_bytes(hexstr=result_hash),
            signature,
        ).transact()
        return self._w3.eth.wait_for_transaction_receipt(tx_hash)

    def refund(self, job_id: str) -> bytes:
        """Refund the consumer after the deadline."""
        tx_hash = self._contract.functions.refund(
            Web3.to_bytes(hexstr=job_id),
        ).transact()
        return self._w3.eth.wait_for_transaction_receipt(tx_hash)

    def get_job(self, job_id: str) -> Job:
        """Read the escrowed job details."""
        result = self._contract.functions.jobs(
            Web3.to_bytes(hexstr=job_id),
        ).call()
        return Job(
            consumer=result[0],
            provider=result[1],
            attestor=result[2],
            amount=result[3],
            deadline=result[4],
            status=result[5],
        )

    def verify(self, commitment_hash: str, job_id: str, result_text: str) -> bool:
        """Verify a claimed commitment hash against an independently recomputed one.

        Pure recompute-to-verify (Layer 2) — no contract call or gas needed.
        """
        return compute_verdict_hash(job_id, result_text) == commitment_hash

    def _get_opened_events(self, receipt) -> list[OpenedEvent]:
        """Parse Opened events from a transaction receipt."""
        events = self._contract.events.Opened().process_receipt(receipt, errors=DISCARD)
        return [
            OpenedEvent(
                job_id=Web3.to_hex(e["args"]["jobId"]),
                consumer=e["args"]["consumer"],
                provider=e["args"]["provider"],
                attestor=e["args"]["attestor"],
                amount=e["args"]["amount"],
                deadline=e["args"]["deadline"],
            )
            for e in events
        ]

    def _get_released_events(self, receipt) -> list[ReleasedEvent]:
        """Parse Released events from a transaction receipt."""
        events = self._contract.events.Released().process_receipt(receipt, errors=DISCARD)
        return [
            ReleasedEvent(
                job_id=Web3.to_hex(e["args"]["jobId"]),
                result_hash=Web3.to_hex(e["args"]["resultHash"]),
                commitment_hash=Web3.to_hex(e["args"]["commitmentHash"]),
                provider=e["args"]["provider"],
                amount=e["args"]["amount"],
            )
            for e in events
        ]

    def _get_refunded_events(self, receipt) -> list[RefundedEvent]:
        """Parse Refunded events from a transaction receipt."""
        events = self._contract.events.Refunded().process_receipt(receipt, errors=DISCARD)
        return [
            RefundedEvent(
                job_id=Web3.to_hex(e["args"]["jobId"]),
                consumer=e["args"]["consumer"],
                amount=e["args"]["amount"],
            )
            for e in events
        ]
