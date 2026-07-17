from __future__ import annotations

from eth_utils import keccak, to_bytes, to_checksum_address


def _pad_to_32(value: bytes) -> bytes:
    """Left-pad bytes to 32 bytes.

    Solidity's abi.encode left-pads uint and address values to 32 bytes.
    """
    if len(value) > 32:
        raise ValueError(f"Value exceeds 32 bytes: {len(value)}")
    return b"\x00" * (32 - len(value)) + value


def _encode_fixed_types(types: list[str], values: list) -> bytes:
    """Replicate Solidity's abi.encode for fixed-size types only.

    This is a minimal implementation covering the types used by ERC-8301:
    uint8, uint256, bytes32, and address. Each value is returned as a
    32-byte segment, concatenated in order — no length prefix, no dynamic
    encoding.

    Args:
        types: List of Solidity type strings.
        values: List of Python values matching the types.

    Returns:
        ABI-encoded bytes (concatenation of 32-byte segments).
    """
    encoded = b""
    for typ, val in zip(types, values):
        if typ in ("uint8", "uint256"):
            encoded += val.to_bytes(32, "big")
        elif typ == "bytes32":
            if isinstance(val, str):
                encoded += to_bytes(hexstr=val)
            elif isinstance(val, bytes):
                if len(val) > 32:
                    raise ValueError(f"bytes32 value exceeds 32 bytes: {len(val)}")
                encoded += val.rjust(32, b"\x00")
            else:
                encoded += val
        elif typ == "address":
            if isinstance(val, str):
                addr_bytes = to_bytes(hexstr=to_checksum_address(val))
            else:
                addr_bytes = val
            encoded += _pad_to_32(addr_bytes)
        else:
            raise ValueError(f"Unsupported ABI type: {typ}")
    return encoded


def compute_task_hash(
    stage: int,
    task_seq: int,
    input_hash: str,
    timestamp: int,
    expires_at: int,
    prev_reply_hashes_packed: str,
    workflow_run_id: str,
) -> str:
    """Compute the task hash for an AgentTask.

    ERC-8301 §AgentTask:
        taskHash = keccak256(abi.encode(
            stage, taskSeq, inputHash, timestamp, expiresAt,
            keccak256(abi.encodePacked(prevReplyHashes)),
            workflowRunId))

    When prev_reply_hashes_packed is empty ("0x"), the inner hash is
    keccak256("") = 0xc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470
    — NOT bytes32(0).

    Args:
        stage: FSM stage (developer-defined enum cast to uint8).
        task_seq: Per-run monotonic counter.
        input_hash: keccak256(input) as 0x-prefixed 32-byte hex string.
        timestamp: block.timestamp at emission.
        expires_at: Unix timestamp after which this task expires.
        prev_reply_hashes_packed: Concatenated prevReplyHashes (empty -> "0x").
        workflow_run_id: Run identifier (0x-prefixed 32-byte hex string).

    Returns:
        0x-prefixed 32-byte hex string of the task hash.
    """
    # innerHash = keccak256(abi.encodePacked(prevReplyHashesPacked))
    inner_hash = keccak(to_bytes(hexstr=prev_reply_hashes_packed))

    # taskHash = keccak256(abi.encode(stage, taskSeq, inputHash, ...))
    encoded = _encode_fixed_types(
        ["uint8", "uint256", "bytes32", "uint256", "uint256", "bytes32", "bytes32"],
        [stage, task_seq, input_hash, timestamp, expires_at, inner_hash, workflow_run_id],
    )
    task_hash = keccak(encoded)
    return "0x" + task_hash.hex()


def compute_reply_hash(
    output_hash: str,
    timestamp: int,
    replier: str,
    prev_task_hashes_packed: str,
    workflow_run_id: str,
) -> str:
    """Compute the reply hash for an AgentReply.

    ERC-8301 §AgentReply:
        replyHash = keccak256(abi.encode(
            outputHash, timestamp, replier,
            keccak256(abi.encodePacked(prevTaskHashes)),
            workflowRunId))

    When prev_task_hashes_packed is empty ("0x"), the inner hash is
    keccak256("") = 0xc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470
    — NOT bytes32(0).

    Args:
        output_hash: keccak256(output) as 0x-prefixed 32-byte hex string.
        timestamp: Off-chain execution time (Unix).
        replier: Agent address (0x-prefixed 20-byte hex string).
        prev_task_hashes_packed: Concatenated prevTaskHashes (empty -> "0x").
        workflow_run_id: Run identifier (0x-prefixed 32-byte hex string).

    Returns:
        0x-prefixed 32-byte hex string of the reply hash.
    """
    # innerHash = keccak256(abi.encodePacked(prevTaskHashesPacked))
    inner_hash = keccak(to_bytes(hexstr=prev_task_hashes_packed))

    # replyHash = keccak256(abi.encode(outputHash, timestamp, replier, ...))
    encoded = _encode_fixed_types(
        ["bytes32", "uint256", "address", "bytes32", "bytes32"],
        [output_hash, timestamp, replier, inner_hash, workflow_run_id],
    )
    reply_hash = keccak(encoded)
    return "0x" + reply_hash.hex()
