"""ERC-8004: Agent Identity Registry."""

from .client import IdentityRegistryClient, MetadataEntry
from .recompute import compute_agent_id

__all__ = [
    "IdentityRegistryClient",
    "MetadataEntry",
    "compute_agent_id",
]
