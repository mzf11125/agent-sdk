import hashlib
import json


def compute_raw_proposal_hash(artifact: str) -> str:
    """
    Compute the raw proposal hash for a JudgmentExecutionAttestation (L4).

    ERC-8299 "L4 Judgment validator binding": rawProposalHash = sha256(artifact).

    Unlike the base WYRIWE triple-hash (L1-L3, keccak256-based -- see recompute.py),
    the L4 layer is designed to anchor off-chain (Nostr-relay-published) verdicts as
    well as on-chain ones, so it uses sha256 over the raw artifact string directly --
    matching invinoveritas's reference implementation (services/proof_signing.py:
    `artifact_hash = sha256(artifact)`), not the EVM-native keccak256 the L1-L3
    layer uses for its on-chain contract calls.

    Args:
        artifact: The exact proposed-action text as submitted for review (a diff,
            command, plan, etc. -- whatever string the producer reviewed).

    Returns:
        The 32-byte sha256 hash of the UTF-8 artifact bytes, 0x-prefixed.
    """
    return "0x" + hashlib.sha256(artifact.encode("utf-8")).hexdigest()


def compute_verdict_hash(fields: dict, preimage_fields: tuple) -> str:
    """
    Compute the verdict hash for a JudgmentExecutionAttestation (L4).

    ERC-8299 "L4 Judgment validator binding": verdictHash binds the verdict metadata
    to rawProposalHash so a verdict cannot be replayed against a different proposal
    than the one it judged ("verdict-shopping"). invinoveritas's reference
    implementation computes this as `decision_ref = sha256(JCS({preimage fields}))`
    over a fixed, sorted set of named fields -- NOT `keccak256(verdict_event_id ||
    rawProposalHash)` as a literal byte-concatenation. JCS (RFC 8785: sorted keys,
    no extraneous whitespace, literal UTF-8 -- NOT \\uXXXX-escaped) so the preimage
    bytes recompute identically across implementations/languages.

    `fields` MUST include whichever key pins the proposal (e.g. `raw_proposal_hash`)
    among `preimage_fields` if the producer wants the anti-replay binding; this
    function does not hardcode a specific field set -- pass the producer's own
    `decision_ref_preimage_fields` (published on each proof) so a recompute matches
    the policy version that was actually in force when the verdict was issued.

    Args:
        fields: The verdict's named metadata fields (string or None values).
        preimage_fields: Which keys of `fields` to include, in any order (they get
            sorted before hashing regardless).

    Returns:
        `sha256:<hex>` -- matching invinoveritas's `decision_ref` format.
    """
    preimage = {k: fields.get(k) for k in preimage_fields}
    canon = json.dumps(preimage, sort_keys=True, separators=(",", ":"), ensure_ascii=False)
    return "sha256:" + hashlib.sha256(canon.encode("utf-8")).hexdigest()
