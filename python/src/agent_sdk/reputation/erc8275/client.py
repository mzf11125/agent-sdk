from __future__ import annotations

from dataclasses import dataclass

from eth_account.signers.local import LocalAccount
from web3 import Web3

from .abi import AGENT_REPUTATION_ABI


@dataclass(frozen=True)
class ReputationData:
    completed_orders: int
    disputed_orders: int
    total_volume: int
    last_active_at: int
    score: int


class AgentReputationClient:
    """Client for the ERC-8275 Agent Reputation contract."""

    def __init__(self, rpc_url: str, address: str, account: LocalAccount):
        self._w3 = Web3(Web3.HTTPProvider(rpc_url))
        self._account = account
        self._contract = self._w3.eth.contract(
            address=Web3.to_checksum_address(address), abi=AGENT_REPUTATION_ABI
        )

    def get_reputation(self, agent_id: str) -> ReputationData:
        """Read the current reputation snapshot for an agent.

        Args:
            agent_id: The agent's identifier (hex-encoded bytes32).

        Returns:
            ReputationData with completed/disputed orders, volume,
            last active timestamp, and score.
        """
        result = self._contract.functions.getReputation(
            Web3.to_bytes(hexstr=agent_id)
        ).call()
        return ReputationData(
            completed_orders=result[0],
            disputed_orders=result[1],
            total_volume=result[2],
            last_active_at=result[3],
            score=result[4],
        )

    def get_decay_weight(self, agent_id: str) -> int:
        """Read the recency-decay weight for an agent's score.

        Args:
            agent_id: The agent's identifier (hex-encoded bytes32).

        Returns:
            Decay weight in basis points (10000 = no decay).
        """
        return self._contract.functions.getDecayWeight(
            Web3.to_bytes(hexstr=agent_id)
        ).call()

    def verify_outcome(self, order_id: str, proof: bytes) -> bool:
        """Verify a settled order's outcome proof against the public record.

        This is a read-only call (no gas, no broadcast) — anyone can derive
        the answer without spending gas.

        Args:
            order_id: Identifier of the settled order (hex-encoded bytes32).
            proof: Implementation-defined proof of the settled outcome.

        Returns:
            True if the outcome is valid against public on-chain data.
        """
        return self._contract.functions.verifyOutcome(
            Web3.to_bytes(hexstr=order_id),
            proof,
        ).call()
