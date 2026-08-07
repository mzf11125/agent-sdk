package erc8263

import "github.com/ethereum/go-ethereum/common"

// AnchorProofEvent is the IOnChainProof.AnchorProof event (ERC-8263),
// emitted exactly once for every successful anchor across both entrypoints.
// The event log is the ledger — the interface exposes no on-chain getter —
// so IsAnchored scans the contract's AnchorProof events keyed by proofHash
// (topics[2]).
//
// agentIdScheme (uint8, data segment — deliberately not indexed) and aux
// (bytes, data segment) are decoded from the event's data field by the
// consumer; this struct carries the indexed topics plus aux for callers that
// parse events themselves.
type AnchorProofEvent struct {
	AgentID   common.Hash    // 32-byte agent identifier per agentIdScheme (topics[1]).
	ProofHash common.Hash    // Non-zero 32-byte commitment to the action (topics[2]).
	Operator  common.Address // Transaction submitter (topics[3]); NOT an authorization claim.
	Aux       []byte         // Opaque, explicitly non-normative extension bytes.
}
