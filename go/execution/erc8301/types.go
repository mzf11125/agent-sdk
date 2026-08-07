package erc8301

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

// RunStatus is the terminal status of a workflow run (ERC-8301 RunStatus).
type RunStatus uint8

const (
	RunStatusPending RunStatus = iota // Run is active; no result yet
	RunStatusSuccess                  // FSM reached a terminal stage with all gates satisfied
	RunStatusFailed                   // Run aborted
)

// AgentTask is a task dispatched by the workflow contract to agents
// (ERC-8301 AgentTask). The contract stores the full struct and derives
// taskHash from its fields, so every value is recomputable from public data.
type AgentTask struct {
	Stage           uint8         // FSM stage (developer-defined enum cast to uint8).
	TaskSeq         *big.Int      // Per-run monotonic counter; starts at 0.
	InputHash       common.Hash   // keccak256(input).
	Input           []byte        // Input plaintext; MAY be empty if conveyed off-chain.
	Timestamp       *big.Int      // block.timestamp at emission.
	ExpiresAt       *big.Int      // Unix timestamp after which this task is no longer valid.
	PrevReplyHashes []common.Hash // Replies that triggered this task; empty for the initial task.
	WorkflowRunId   common.Hash   // Run identifier.
}

// AgentReply is a reply submitted by an agent in response to a dispatched
// task (ERC-8301 AgentReply).
type AgentReply struct {
	OutputHash     common.Hash    // keccak256(output).
	Output         []byte         // Reply output plaintext; MAY be empty.
	Timestamp      *big.Int       // Off-chain execution time (Unix).
	Replier        common.Address // Agent address; MUST equal msg.sender when submitted.
	PrevTaskHashes []common.Hash  // Tasks this reply responds to; MUST be non-empty.
	WorkflowRunId  common.Hash    // Run identifier.
}
