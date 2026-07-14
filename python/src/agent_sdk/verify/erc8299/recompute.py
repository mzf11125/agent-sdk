from eth_hash.auto import keccak
from eth_utils import to_bytes, to_hex


def compute_raw_input_hash(raw_input_hex: str) -> str:
    """
    Compute the raw input hash for a WYRIWE attestation.

    ERC-8299 Section 45: raw_input_hash = keccak256(raw_user_input).

    Args:
        raw_input_hex: The raw user input as a 0x-prefixed hex string.

    Returns:
        The 32-byte keccak256 hash of the raw input, 0x-prefixed.
    """
    raw_bytes = to_bytes(hexstr=raw_input_hex)
    return to_hex(keccak(raw_bytes))


def compute_sanitization_pipeline_hash(spec_cid: str, raw_input_hash: str) -> str:
    """
    Compute the sanitization pipeline hash for a WYRIWE attestation.

    ERC-8299 Section 46: sanitization_pipeline_hash =
        keccak256(utf8(cid) || raw_input_hash).

    The spec_cid is converted to UTF-8 bytes, then concatenated with the
    raw_input_hash bytes before hashing.

    Args:
        spec_cid: The sanitization spec CID string (e.g. "ipfs://Qm...").
        raw_input_hash: The raw input hash (32 bytes, 0x-prefixed).

    Returns:
        The 32-byte keccak256 hash of the concatenated bytes, 0x-prefixed.
    """
    cid_bytes = spec_cid.encode("utf-8")
    raw_bytes = to_bytes(hexstr=raw_input_hash)
    return to_hex(keccak(cid_bytes + raw_bytes))
