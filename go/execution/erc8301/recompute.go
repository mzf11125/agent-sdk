// Package erc8301 implements the ERC-8301 AI Agent Execution SDK.
//
// ERC-8301 (Agent Workflow) links every dispatched task and submitted reply
// through deterministic hashes:
//
//	taskHash = keccak256(abi.encode(stage, taskSeq, inputHash, timestamp,
//	    expiresAt, keccak256(abi.encodePacked(prevReplyHashes)), workflowRunId))
//	replyHash = keccak256(abi.encode(outputHash, timestamp, replier,
//	    keccak256(abi.encodePacked(prevTaskHashes)), workflowRunId))
//
// Both are pure functions of public struct fields — derived-not-stored. The
// contract exposes the full structs via getAgentTask/getAgentReply, so anyone
// can independently recompute either hash without trusting the party that
// submitted the task or reply.
package erc8301

import (
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

// ComputeInnerHash computes the inner hash of an AgentTask or AgentReply:
// keccak256 of the raw concatenated prev-hash list (abi.encodePacked).
//
// CRITICAL: when prevHashesPacked is empty the result is keccak256("") =
// 0xc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470 — NOT
// bytes32(0). Special-casing the empty list to the zero hash produces wrong
// task/reply hashes and silently breaks the evidence chain.
func ComputeInnerHash(prevHashesPacked []byte) common.Hash {
	return crypto.Keccak256Hash(prevHashesPacked)
}

// ComputeTaskHash computes the task hash for an ERC-8301 AgentTask.
//
//	taskHash = keccak256(abi.encode(stage, taskSeq, inputHash, timestamp,
//	    expiresAt, innerHash, workflowRunId))
//
// where innerHash = keccak256(abi.encodePacked(prevReplyHashesPacked)) —
// pass nil (or empty) for a task with no previous replies; the empty-list
// inner hash is keccak256(""), not bytes32(0).
//
// Golden vector (recompute-kit "8301/task-hash", cross-verified against the
// TypeScript and Rust SDKs): stage=1, taskSeq=0, inputHash
// 0x1c8aff950685c2ed4bc3174f3472287b56d9517b9c948127319a09a7a36deac8,
// timestamp=1700000000, expiresAt=1700001000, prevReplyHashesPacked empty,
// workflowRunId 0x00000000000000000000000000000000000000000000000000000000deadbeef
// → 0xf1f404c844a4aff1d0d7d17cebb518a2d386197aad09ab86517eaa01448301ec.
//
// The ABI encode is the mixed-type pattern: uint8, uint256, bytes32, uint256,
// uint256, bytes32, bytes32 — every value is static, so the packed encoding
// is a bare 224-byte concatenation. Returns an error if a type or the pack
// fails; never panics.
func ComputeTaskHash(stage uint8, taskSeq uint64, inputHash common.Hash, timestamp uint64, expiresAt uint64, prevReplyHashesPacked []byte, workflowRunId common.Hash) (common.Hash, error) {
	innerHash := ComputeInnerHash(prevReplyHashesPacked)

	uint8Type, err := abi.NewType("uint8", "", nil)
	if err != nil {
		return common.Hash{}, err
	}
	uint256Type, err := abi.NewType("uint256", "", nil)
	if err != nil {
		return common.Hash{}, err
	}
	bytes32Type, err := abi.NewType("bytes32", "", nil)
	if err != nil {
		return common.Hash{}, err
	}
	args := abi.Arguments{
		{Type: uint8Type},
		{Type: uint256Type},
		{Type: bytes32Type},
		{Type: uint256Type},
		{Type: uint256Type},
		{Type: bytes32Type},
		{Type: bytes32Type},
	}
	packed, err := args.Pack(
		stage,
		new(big.Int).SetUint64(taskSeq),
		inputHash,
		new(big.Int).SetUint64(timestamp),
		new(big.Int).SetUint64(expiresAt),
		innerHash,
		workflowRunId,
	)
	if err != nil {
		return common.Hash{}, err
	}
	return crypto.Keccak256Hash(packed), nil
}

// ComputeReplyHash computes the reply hash for an ERC-8301 AgentReply.
//
//	replyHash = keccak256(abi.encode(outputHash, timestamp, replier,
//	    innerHash, workflowRunId))
//
// where innerHash = keccak256(abi.encodePacked(prevTaskHashesPacked)).
//
// ABI encode types: bytes32, uint256, address, bytes32, bytes32 — the
// address is left-padded to 32 bytes like any other static value. Returns an
// error if a type or the pack fails; never panics.
func ComputeReplyHash(outputHash common.Hash, timestamp uint64, replier common.Address, prevTaskHashesPacked []byte, workflowRunId common.Hash) (common.Hash, error) {
	innerHash := ComputeInnerHash(prevTaskHashesPacked)

	uint256Type, err := abi.NewType("uint256", "", nil)
	if err != nil {
		return common.Hash{}, err
	}
	bytes32Type, err := abi.NewType("bytes32", "", nil)
	if err != nil {
		return common.Hash{}, err
	}
	addressType, err := abi.NewType("address", "", nil)
	if err != nil {
		return common.Hash{}, err
	}
	args := abi.Arguments{
		{Type: bytes32Type},
		{Type: uint256Type},
		{Type: addressType},
		{Type: bytes32Type},
		{Type: bytes32Type},
	}
	packed, err := args.Pack(
		outputHash,
		new(big.Int).SetUint64(timestamp),
		replier,
		innerHash,
		workflowRunId,
	)
	if err != nil {
		return common.Hash{}, err
	}
	return crypto.Keccak256Hash(packed), nil
}
