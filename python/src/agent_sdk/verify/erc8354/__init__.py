"""ERC-8354 Confidential Agent Policy Verdicts — pre-execution ZK allow/deny."""

from .client import (
    ConfidentialPolicyVerdictClient,
    PolicyDomainRegistryClient,
    Verdict,
)
from .recompute import (
    MECHANISM_ZK_SECRET_POLICY,
    compute_action_commitment,
    compute_verdict_digest,
)

__all__ = [
    "ConfidentialPolicyVerdictClient",
    "PolicyDomainRegistryClient",
    "Verdict",
    "compute_action_commitment",
    "compute_verdict_digest",
    "MECHANISM_ZK_SECRET_POLICY",
]
