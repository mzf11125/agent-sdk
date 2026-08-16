"""ERC-8354 Confidential Agent Policy Verdicts — guard and registry clients."""

from __future__ import annotations

from dataclasses import dataclass

from eth_account.signers.local import LocalAccount
from web3 import Web3
from web3.middleware import SignAndSendRawMiddlewareBuilder

from .abi import CONFIDENTIAL_POLICY_VERDICT_ABI, POLICY_DOMAIN_REGISTRY_ABI

_CONFIDENTIAL_POLICY_VERDICT_INTERFACE_ID = "0xd6da8150"


@dataclass(frozen=True)
class _DomainOutput:
    registrar: str
    verifier: str
    program_key: bytes
    max_root_age: int
    active: bool


@dataclass
class Verdict:
    agent_id: int
    domain_id: bytes
    policy_root: bytes
    action_commitment: bytes
    executor: str
    expiry: int
    nullifier: bytes
    decision: int
    policy_kind: int

    def _as_dict(self):
        return {
            "agentId": self.agent_id,
            "domainId": self.domain_id,
            "policyRoot": self.policy_root,
            "actionCommitment": self.action_commitment,
            "executor": self.executor,
            "expiry": self.expiry,
            "nullifier": self.nullifier,
            "decision": self.decision,
            "policyKind": self.policy_kind,
        }


class ConfidentialPolicyVerdictClient:
    """ERC-8354 Confidential Policy Verdicts, guard side.

    Consumes a zero-knowledge verdict that an action was evaluated against a
    committed secret policy and permitted. ``verify`` is a read-only view
    call; ``consume`` burns the verdict's single-use nullifier and gates
    execution.
    """

    def __init__(self, rpc_url: str, address: str, account: LocalAccount):
        self._w3 = Web3(Web3.HTTPProvider(rpc_url))
        self._w3.middleware_onion.add(SignAndSendRawMiddlewareBuilder.build(account))
        self._w3.eth.default_account = account.address
        self._contract = self._w3.eth.contract(
            address=Web3.to_checksum_address(address),
            abi=CONFIDENTIAL_POLICY_VERDICT_ABI,
        )
        self._account = account

    def verify(self, verdict: Verdict, proof: bytes) -> bool:
        """Verify a verdict without state change. Returns False on a well-formed
        but invalid verdict, and never reverts on a malformed proof."""
        return self._contract.functions.verify(verdict._as_dict(), proof).call()

    def verdict_digest(self, verdict: Verdict) -> bytes:
        """The EIP-712 digest an executor signs to authorize a relayer."""
        return self._contract.functions.verdictDigest(verdict._as_dict()).call()

    def is_consumed(self, domain_id: bytes, nullifier: bytes) -> bool:
        """Whether the verdict's nullifier has been burned for the domain."""
        return self._contract.functions.isConsumed(domain_id, nullifier).call()

    def supports_interface(self) -> bool:
        """ERC-165: does the contract advertise IConfidentialPolicyVerdict?"""
        return self._contract.functions.supportsInterface(
            bytes.fromhex(_CONFIDENTIAL_POLICY_VERDICT_INTERFACE_ID[2:])
        ).call()

    def consume(self, verdict: Verdict, proof: bytes) -> dict:
        """Verify and burn a verdict directly. The caller (tx sender) must be
        the executor. Returns the transaction receipt."""
        tx_hash = self._contract.functions.consume(verdict._as_dict(), proof).transact()
        receipt = self._w3.eth.wait_for_transaction_receipt(tx_hash)
        return dict(receipt)

    def consume_relayed(self, verdict: Verdict, proof: bytes, executor_auth: bytes) -> dict:
        """Verify and burn a verdict via a relayer. executor_auth is a valid
        EIP-712 signature by the executor over verdictDigest(v)."""
        tx_hash = self._contract.functions.consume(
            verdict._as_dict(), proof, executor_auth
        ).transact()
        receipt = self._w3.eth.wait_for_transaction_receipt(tx_hash)
        return dict(receipt)


class PolicyDomainRegistryClient:
    """ERC-8354 recommended companion registry. Read-only view calls for the
    guard's checks. No account needed (no writes)."""

    def __init__(self, rpc_url: str, address: str):
        self._w3 = Web3(Web3.HTTPProvider(rpc_url))
        self._contract = self._w3.eth.contract(
            address=Web3.to_checksum_address(address),
            abi=POLICY_DOMAIN_REGISTRY_ABI,
        )

    def domain(self, domain_id: bytes) -> _DomainOutput:
        """The Domain record for a domain id."""
        result = self._contract.functions.domain(domain_id).call()
        return _DomainOutput(
            registrar=result[0],
            verifier=result[1],
            program_key=result[2],
            max_root_age=result[3],
            active=result[4],
        )

    def current_root(self, domain_id: bytes) -> tuple[bytes, int, int]:
        """The current root, version, and update timestamp."""
        root, version, updated_at = self._contract.functions.currentRoot(domain_id).call()
        return root, version, updated_at

    def is_root_acceptable(self, domain_id: bytes, root: bytes) -> bool:
        """Whether a root is current or superseded within the grace window."""
        return self._contract.functions.isRootAcceptable(domain_id, root).call()