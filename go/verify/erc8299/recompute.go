// Package erc8299 implements the ERC-8299 WYRIWE — Input Provenance for AI
// Inference (Go) SDK.
//
// ERC-8299 binds what a user submitted (`rawInputHash`) to what the model
// actually received (`inputHash`) via a public, replayable sanitization
// pipeline (`sanitizationPipelineHash`). Both hashes are pure recomputes:
// any verifier holding the raw input and the pinned sanitization spec CID
// can re-derive them without trusting a third party. The on-chain layer
// (IWyriweAttestation / IJudgmentExecutionAttestation) authenticates
// attestations via an EIP-712-style signature checked by the contract's
// verify(attestation, signature) view.
package erc8299

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

// ComputeRawInputHash computes rawInputHash = keccak256(raw_user_input)
// (ERC-8299 §45) — the first step of the WYRIWE triple-hash construction.
//
// The raw user input bytes are hashed to produce the hash the attestation
// pins. Keccak-256 never fails, so the function returns a single value
// rather than (result, error), and never panics.
//
// Golden vector: ComputeRawInputHash([]byte("hello")) =
// 0x1c8aff950685c2ed4bc3174f3472287b56d9517b9c948127319a09a7a36deac8.
func ComputeRawInputHash(rawInput []byte) common.Hash {
	return crypto.Keccak256Hash(rawInput)
}

// ComputeSanitizationPipelineHash computes
// sanitizationPipelineHash = keccak256(utf8(cid) || rawInputHash)
// (ERC-8299 §46) — the second step of the WYRIWE triple-hash construction.
//
// The spec CID is converted to UTF-8 bytes, concatenated with the
// rawInputHash bytes (the plan's `append(rawHash[:], innerHash...)`
// pattern), and hashed. Keccak-256 never fails, so the function returns a
// single value rather than (result, error), and never panics.
//
// Golden vector:
// ComputeSanitizationPipelineHash("ipfs://QmccvoM6aRVgZ2dtFWvT6Wm3DmTvoAUHHotK7uQufnStVR",
// hexToHash("0x1c8aff950685c2ed4bc3174f3472287b56d9517b9c948127319a09a7a36deac8")) =
// 0x5798efed4aa92f96a0622fc30268042b067294bdb5fd06f599bf8d84fd5d734b.
func ComputeSanitizationPipelineHash(cid string, rawInputHash common.Hash) common.Hash {
	buf := make([]byte, 0, len(cid)+common.HashLength)
	buf = append(buf, cid...)
	buf = append(buf, rawInputHash[:]...)
	return crypto.Keccak256Hash(buf)
}
