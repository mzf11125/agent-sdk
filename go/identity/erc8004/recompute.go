// Package erc8004 implements the ERC-8004 Agent Identity SDK.
//
// ERC-8004 (Trustless Agents, Identity Registry) assigns every registered
// agent an integer id and exposes it on-chain as the 32-byte agentId. The
// derivation is a left-padded zero-extension of the registry id — not a hash
// — so it is exactly recomputable off-chain from public data.
package erc8004

import (
	"encoding/binary"

	"github.com/ethereum/go-ethereum/common"
)

// ComputeAgentId computes the on-chain agentId for a registry-assigned id
// (ERC-8004 / ERC-8299).
//
// agentId = bytes32(uint256(registryId)) — a left-padded zero-extension of
// the registry id, NOT a hash.
//
// Golden vector: registryId 860 →
// 0x000000000000000000000000000000000000000000000000000000000000035c.
//
// Never panics and cannot fail — the registry id always fits in bytes32.
func ComputeAgentId(registryId uint64) common.Hash {
	var b common.Hash
	binary.BigEndian.PutUint64(b[24:], registryId)
	return b
}
