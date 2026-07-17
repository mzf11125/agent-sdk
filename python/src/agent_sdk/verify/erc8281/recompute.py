"""Pure recompute functions for ERC-8281 Observation Commitment Protocol (OCP).

These functions reproduce the deterministic computations the ERC specifies,
without any blockchain dependency. Verified against golden conformance vectors.
"""

from eth_utils import keccak


def compute_observation_digest(observation: bytes) -> bytes:
    """Compute ``digest = keccak256(observation)``.

    This is the core OCP commitment step (ERC-8281 §1): the observation
    bytes are hashed to produce the opaque digest that is anchored on-chain
    via ``record(digest)``.

    Args:
        observation: The raw observation bytes to commit.

    Returns:
        The 32-byte keccak256 digest.
    """
    return keccak(observation)
