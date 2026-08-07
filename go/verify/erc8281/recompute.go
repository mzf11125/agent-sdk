// Package erc8281 implements the ERC-8281 Observation Commitment Protocol
// (OCP) SDK.
//
// ERC-8281 anchors an opaque commitment digest on-chain via record(digest),
// emitting the Recorded(digest, committer) event as tamper-evident,
// timestamped proof-of-existence. The observation itself stays off-chain;
// verification is recompute-based — a verifier re-derives the digest from
// the primary artifact and confirms the matching Recorded log exists.
package erc8281

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

// ComputeObservationDigest computes digest = keccak256(observation)
// (ERC-8281 §1) — the core OCP commitment step.
//
// The observation bytes are hashed to produce the opaque digest that is
// anchored on-chain via record(digest). Keccak-256 never fails, so the
// function returns a single value rather than (result, error), and never
// panics.
//
// Golden vector: ComputeObservationDigest([]byte("hello")) =
// 0x1c8aff950685c2ed4bc3174f3472287b56d9517b9c948127319a09a7a36deac8.
func ComputeObservationDigest(observation []byte) common.Hash {
	return crypto.Keccak256Hash(observation)
}
