// Package erc8203 implements the ERC-8203 ConsultEscrow Settlement SDK.
//
// ERC-8203 (ConsultEscrow) locks payment for an agent consultation and
// releases it to the provider only when the named attestor signs the
// result's commitment. The release commitment is recomputed ON-CHAIN from
// `jobId` — `commitmentHash = keccak256(abi.encode(jobId, resultHash))`
// with `resultHash = keccak256(utf8(resultText))` — so the outcome is fully
// re-derivable from public data: both jobId and resultHash are visible in
// the Released event, and any verifier holding the result text can
// recompute the verdict without trusting a third party.
package erc8203

import (
	"fmt"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

// ComputeVerdictHash computes the release commitment for an ERC-8203
// ConsultEscrow job:
//
//	resultHash  = keccak256(utf8(resultText))
//	verdictHash = keccak256(abi.encode(bytes32 jobId, bytes32 resultHash))
//
// This is exactly what ConsultEscrow.release() recomputes on-chain before
// checking the attestor's EIP-191 signature — the recompute-kit
// "8203/settlement-proof" recipe. The result is fully re-derivable from
// public data (jobId and resultText are both posted on-chain in the
// Opened/Released events).
//
// Golden vector (recompute-kit 8203/settlement-proof, cross-verified
// against the TypeScript and Rust SDKs):
//
//	jobId      0xbc01b40fe7a3509f35470053d4bc1844d50c9782546cf0fc11154adcb90caa56
//	resultText "No intermediaries required, cryptographic verification only."
//	verdict    0xdc568bd1cbacdd1ead8231e9d3d6f4e475f5168f3cc9f72b31935d46cfdd48f7
//
// Returns an error if the bytes32 ABI type cannot be constructed or the
// arguments fail to pack; never panics.
func ComputeVerdictHash(jobID common.Hash, resultText string) (common.Hash, error) {
	resultHash := crypto.Keccak256Hash([]byte(resultText))

	bytes32Type, err := abi.NewType("bytes32", "", nil)
	if err != nil {
		return common.Hash{}, fmt.Errorf("erc8203: bytes32 type: %w", err)
	}
	// abi.encode(bytes32 jobId, bytes32 resultHash) — the on-chain
	// commitment construction. Unlike a method call, no 4-byte selector is
	// prepended here: this is raw abi.encode, not calldata.
	encoded, err := abi.Arguments{{Type: bytes32Type}, {Type: bytes32Type}}.Pack(jobID, resultHash)
	if err != nil {
		return common.Hash{}, fmt.Errorf("erc8203: pack (jobId, resultHash): %w", err)
	}
	return crypto.Keccak256Hash(encoded), nil
}
