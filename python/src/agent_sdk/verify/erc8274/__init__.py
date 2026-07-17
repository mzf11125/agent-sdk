"""ERC-8274: Agent Verifiable & Proof Verifier."""

from .client import AgentVerifierClient, ProofVerifierClient, get_trusted_verifier

__all__ = [
    "AgentVerifierClient",
    "ProofVerifierClient",
    "get_trusted_verifier",
]
