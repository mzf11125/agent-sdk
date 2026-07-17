import { encodeAbiParameters, keccak256 } from 'viem'
import type { Hex } from 'viem'

/**
 * Compute the inner hash for an AgentTask or AgentReply.
 *
 * ERC-8301 §AgentTask / §AgentReply:
 *   innerHash = keccak256(abi.encodePacked(hashesPacked))
 *
 * When hashesPacked is empty ("0x"), the result is keccak256("") =
 * 0xc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470.
 * An implementation that returns bytes32(0) for the empty case is WRONG.
 *
 * @param hashesPacked - The concatenated prevReplyHashes or prevTaskHashes as packed bytes.
 * @returns The 32-byte inner hash.
 */
function computeInnerHash(hashesPacked: Hex): Hex {
  return keccak256(hashesPacked)
}

/**
 * Compute the task hash for an AgentTask.
 *
 * ERC-8301 §AgentTask:
 *   taskHash = keccak256(abi.encode(
 *       stage, taskSeq, inputHash, timestamp, expiresAt,
 *       keccak256(abi.encodePacked(prevReplyHashes)),
 *       workflowRunId))
 *
 * When prevReplyHashesPacked is empty ("0x"), the inner hash is
 * keccak256("") = 0xc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470
 * — NOT bytes32(0).
 *
 * @param stage                - FSM stage (developer-defined enum cast to uint8).
 * @param taskSeq              - Per-run monotonic counter.
 * @param inputHash            - keccak256(input) as 32-byte hex.
 * @param timestamp            - block.timestamp at emission.
 * @param expiresAt            - Unix timestamp after which this task expires.
 * @param prevReplyHashesPacked - Concatenated prevReplyHashes (empty → "0x").
 * @param workflowRunId        - Run identifier (32-byte hex).
 * @returns The 32-byte task hash.
 */
export function computeTaskHash(
  stage: number,
  taskSeq: bigint | number,
  inputHash: Hex,
  timestamp: bigint | number,
  expiresAt: bigint | number,
  prevReplyHashesPacked: Hex,
  workflowRunId: Hex,
): Hex {
  const innerHash = computeInnerHash(prevReplyHashesPacked)

  const encoded = encodeAbiParameters(
    [
      { type: 'uint8' },
      { type: 'uint256' },
      { type: 'bytes32' },
      { type: 'uint256' },
      { type: 'uint256' },
      { type: 'bytes32' },
      { type: 'bytes32' },
    ],
    [
      stage,
      BigInt(taskSeq),
      inputHash,
      BigInt(timestamp),
      BigInt(expiresAt),
      innerHash,
      workflowRunId,
    ],
  )

  return keccak256(encoded)
}

/**
 * Compute the reply hash for an AgentReply.
 *
 * ERC-8301 §AgentReply:
 *   replyHash = keccak256(abi.encode(
 *       outputHash, timestamp, replier,
 *       keccak256(abi.encodePacked(prevTaskHashes)),
 *       workflowRunId))
 *
 * When prevTaskHashesPacked is empty ("0x"), the inner hash is
 * keccak256("") = 0xc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470
 * — NOT bytes32(0).
 *
 * @param outputHash           - keccak256(output) as 32-byte hex.
 * @param timestamp            - Off-chain execution time (Unix).
 * @param replier              - Agent address (20-byte hex, 0x-prefixed).
 * @param prevTaskHashesPacked - Concatenated prevTaskHashes (empty → "0x").
 * @param workflowRunId        - Run identifier (32-byte hex).
 * @returns The 32-byte reply hash.
 */
export function computeReplyHash(
  outputHash: Hex,
  timestamp: bigint | number,
  replier: Hex,
  prevTaskHashesPacked: Hex,
  workflowRunId: Hex,
): Hex {
  const innerHash = computeInnerHash(prevTaskHashesPacked)

  const encoded = encodeAbiParameters(
    [
      { type: 'bytes32' },
      { type: 'uint256' },
      { type: 'address' },
      { type: 'bytes32' },
      { type: 'bytes32' },
    ],
    [
      outputHash,
      BigInt(timestamp),
      replier,
      innerHash,
      workflowRunId,
    ],
  )

  return keccak256(encoded)
}
