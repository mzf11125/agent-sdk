from __future__ import annotations

from dataclasses import dataclass

from eth_account.signers.local import LocalAccount
from web3 import Web3
from web3.logs import DISCARD
from web3.middleware import SignAndSendRawMiddlewareBuilder

from .abi import OBSERVATION_COMMITMENT_ABI

# ERC-165 interface id of IObservationCommitment (0xb5c645bd).
OBSERVATION_COMMITMENT_INTERFACE_ID = "0xb5c645bd"


@dataclass(frozen=True)
class RecordedEvent:
    digest: bytes
    committer: str


class ObservationCommitmentClient:
    """ERC-8281 Observation Commitment Protocol (OCP) — write side.

    Anchors an opaque commitment digest on-chain via ``record()``, emitting a
    ``Recorded`` event as tamper-evident proof. The observation itself stays
    off-chain; verification is recompute-based — a verifier re-derives the
    digest from the primary artifact and confirms the matching ``Recorded``
    log exists at the claimed chain/block position.

    There is no on-chain getter: the event log IS the ledger.
    """

    def __init__(self, rpc_url: str, address: str, account: LocalAccount):
        self._w3 = Web3(Web3.HTTPProvider(rpc_url))
        self._w3.middleware_onion.add(SignAndSendRawMiddlewareBuilder.build(account))
        self._w3.eth.default_account = account.address
        self._contract = self._w3.eth.contract(
            address=Web3.to_checksum_address(address), abi=OBSERVATION_COMMITMENT_ABI
        )
        self._account = account

    def record(self, digest: bytes) -> dict:
        """Commit a digest on-chain. Emits ``Recorded(digest, committer)``.

        Returns the transaction receipt so the caller can extract chain/block/log
        position for the proof envelope.
        """
        tx_hash = self._contract.functions.record(digest).transact()
        receipt = self._w3.eth.wait_for_transaction_receipt(tx_hash)
        return dict(receipt)

    def parse_recorded_event(self, receipt: dict) -> RecordedEvent:
        """Extract the ``Recorded`` event from a transaction receipt.

        Raises ``RuntimeError`` if no matching event is found.
        """
        events = self._contract.events.Recorded().process_receipt(receipt, errors=DISCARD)
        if not events:
            raise RuntimeError("Recorded event not found in transaction receipt")
        return RecordedEvent(
            digest=events[0]["args"]["digest"],
            committer=events[0]["args"]["committer"],
        )

    def supports_observation_commitment(self) -> bool:
        """ERC-165: does this contract advertise IObservationCommitment (0xb5c645bd)?"""
        return self._contract.functions.supportsInterface(
            bytes.fromhex(OBSERVATION_COMMITMENT_INTERFACE_ID[2:])
        ).call()
