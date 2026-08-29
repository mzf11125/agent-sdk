"""Pure recompute functions for ERC-8354 Confidential Agent Policy Verdicts.

These functions reproduce the deterministic computations the ERC specifies,
without any blockchain dependency. Verified against golden conformance vectors.
"""

from __future__ import annotations

from dataclasses import asdict, is_dataclass

from eth_utils import keccak, to_bytes, to_checksum_address

MECHANISM_ZK_SECRET_POLICY = "0xa843829a78c66c29679817606d0c8a9fa26575b6c2ed0f9f97079d7c46577ac6"

_VERDICT_TYPE = (
    "Verdict(uint256 agentId,bytes32 domainId,bytes32 policyRoot,bytes32 "
    "actionCommitment,address executor,uint64 expiry,bytes32 nullifier,"
    "uint8 decision,uint8 policyKind)"
)

_DOMAIN_TYPE = (
    "EIP712Domain(string name,string version,uint256 chainId,address verifyingContract)"
)


def _pad32(value: bytes) -> bytes:
    """Left-pad bytes to 32 bytes."""
    return value.rjust(32, b"\x00")


def _encode_fixed(types: list[str], values: list) -> bytes:
    """Replicate Solidity's abi.encode for fixed-size types only.

    Covers uint8, uint64, uint256, bytes32, and address. Each value is a
    32-byte segment, concatenated in order, with no length prefix.
    """
    encoded = b""
    for typ, val in zip(types, values):
        if typ in ("uint8", "uint64", "uint256"):
            encoded += int(val).to_bytes(32, "big")
        elif typ == "bytes32":
            if isinstance(val, str):
                encoded += to_bytes(hexstr=val)
            elif isinstance(val, bytes):
                encoded += val.rjust(32, b"\x00")
            else:
                encoded += val
        elif typ == "address":
            if isinstance(val, str):
                encoded += _pad32(to_bytes(hexstr=to_checksum_address(val)))
            else:
                encoded += _pad32(val)
        else:
            raise ValueError(f"Unsupported ABI type: {typ}")
    return encoded


def compute_action_commitment(
    chain_id: int,
    domain_id: str,
    agent_id: int,
    target: str,
    value: int,
    call_data: str,
    action_nonce: int,
) -> str:
    """Compute the canonical action commitment for a policy verdict.

    ERC-8354 Action commitment:
        actionCommitment = keccak256(abi.encode(chainId, domainId, agentId,
            target, value, keccak256(callData), actionNonce))

    Args:
        chain_id: Chain id the action executes on.
        domain_id: Policy domain id (0x-prefixed 32-byte hex).
        agent_id: ERC-8004 Identity Registry token id.
        target: Call target address (0x-prefixed hex).
        value: Wei forwarded.
        call_data: Call data bytes (0x-prefixed hex).
        action_nonce: Monotonic per (domain, agent) counter.

    Returns:
        0x-prefixed 32-byte hex string of the action commitment.
    """
    call_data_hash = keccak(to_bytes(hexstr=call_data))
    encoded = _encode_fixed(
        ["uint256", "bytes32", "uint256", "address", "uint256", "bytes32", "uint256"],
        [chain_id, domain_id, agent_id, target, value, call_data_hash, action_nonce],
    )
    return "0x" + keccak(encoded).hex()


def compute_verdict_digest(
    verdict: dict,
    chain_id: int,
    verifying_contract: str,
) -> str:
    """Compute the EIP-712 digest an executor signs to authorize a relayer.

    ERC-8354 verdictDigest over the Verdict type, with EIP-712 domain name
    "ConfidentialPolicyVerdict" and version "1".

    Args:
        verdict: Either a dict with keys agent_id, domain_id, policy_root,
            action_commitment, executor, expiry, nullifier, decision,
            policy_kind, or an instance of the exported `Verdict` dataclass
            (which carries the same field names). Both forms compose with the
            public API exported by this package.
        chain_id: Chain id of the verifying guard contract.
        verifying_contract: Address of the guard contract.

    Returns:
        0x-prefixed 32-byte hex string of the EIP-712 digest.
    """
    if is_dataclass(verdict):
        verdict = asdict(verdict)

    verdict_typehash = keccak(text=_VERDICT_TYPE)
    hash_struct = keccak(
        _encode_fixed(
            [
                "bytes32", "uint256", "bytes32", "bytes32", "bytes32",
                "address", "uint64", "bytes32", "uint8", "uint8",
            ],
            [
                verdict_typehash,
                verdict["agent_id"],
                verdict["domain_id"],
                verdict["policy_root"],
                verdict["action_commitment"],
                verdict["executor"],
                verdict["expiry"],
                verdict["nullifier"],
                verdict["decision"],
                verdict["policy_kind"],
            ],
        )
    )

    domain_typehash = keccak(text=_DOMAIN_TYPE)
    name_hash = keccak(text="ConfidentialPolicyVerdict")
    version_hash = keccak(text="1")
    domain_separator = keccak(
        _encode_fixed(
            ["bytes32", "bytes32", "bytes32", "uint256", "address"],
            [domain_typehash, name_hash, version_hash, chain_id, verifying_contract],
        )
    )

    digest = keccak(b"\x19\x01" + domain_separator + hash_struct)
    return "0x" + digest.hex()
