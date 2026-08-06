package erc8299

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

// WyriweAttestation mirrors the IWyriweAttestation.WyriweAttestation struct
// (ERC-8299, L3 input provenance). Field names match the ABI component names
// (PascalCase of the Solidity names) so go-ethereum can pack the struct into
// the verify(tuple, bytes) calldata. The EIP-712 field ordering below is
// normative — reordering changes the typeHash.
type WyriweAttestation struct {
	AgentId                  common.Hash    // ERC-8004 agent identity anchor; MAY be zero if unregistered
	Registry                 common.Address // ERC-8004 registry address
	ModelHash                common.Hash    // hash of model weights or manifest
	RawInputHash             common.Hash    // keccak256(raw_user_input)
	SanitizationPipelineHash common.Hash    // keccak256(utf8(spec_cid) || rawInputHash)
	InputHash                common.Hash    // keccak256(sanitized_input); equals rawInputHash under the identity sentinel
	OutputHash               common.Hash    // keccak256(model_output)
	Timestamp                *big.Int       // unix timestamp of execution
}

// JudgmentExecutionAttestation mirrors the
// IJudgmentExecutionAttestation.JudgmentExecutionAttestation struct
// (ERC-8299, L4 judgment validator chain-of-custody). Field names match the
// ABI component names so go-ethereum can pack the struct into the
// verify(tuple, bytes) calldata. Commit-reveal invariant:
// verdictTimestamp < executedTimestamp MUST hold.
type JudgmentExecutionAttestation struct {
	AgentId            common.Hash    // ERC-8004 identity of the EXECUTING agent
	Registry           common.Address // ERC-8004 registry address
	ValidatorId        common.Hash    // ERC-8004 identity of the judgment validator
	RawProposalHash    common.Hash    // keccak256(canonical proposed-action artifact, pre-review)
	VerdictHash        common.Hash    // keccak256(verdict_artifact_ref || rawProposalHash)
	ExecutedActionHash common.Hash    // keccak256(canonical executed-action record), revealed at settlement
	VerdictTimestamp   *big.Int       // verdict issuance — the commit, strictly pre-execution
	ExecutedTimestamp  *big.Int       // execution — the reveal
	RecordPointer      string         // URI resolving to a RecordPointer-shaped record
}
