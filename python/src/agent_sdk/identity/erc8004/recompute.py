def compute_agent_id(registry_id: int) -> str:
    """
    Compute the agentId for a given registryId.

    ERC-8004 / ERC-8299 (PR #1810): agentId = bytes32(uint256(registryId)).
    This is a left-padded zero-extension of the registry id — not a hash.
    Pure Python — no Web3/eth_utils dependency needed for this trivial case.

    Args:
        registry_id: The registry-assigned agent id (positive integer).

    Returns:
        0x-prefixed 32-byte hex string.
    """
    return "0x" + registry_id.to_bytes(32, "big").hex()
