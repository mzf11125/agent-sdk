package erc8301

import (
	"strings"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi"
)

// agentWorkflowABIJSON is the ERC-8301 IAgentWorkflow interface as JSON,
// hand-written to match the interface exactly (see
// agent-ercs/contracts/execution/ERC8301/IAgentWorkflow.sol).
const agentWorkflowABIJSON = `[
  {
    "type": "function",
    "name": "run",
    "stateMutability": "nonpayable",
    "inputs": [
      {"name": "inputHash", "type": "bytes32"},
      {"name": "input", "type": "bytes"},
      {"name": "expiresAt", "type": "uint256"}
    ],
    "outputs": [{"name": "workflowRunId", "type": "bytes32"}]
  },
  {
    "type": "function",
    "name": "result",
    "stateMutability": "view",
    "inputs": [{"name": "workflowRunId", "type": "bytes32"}],
    "outputs": [
      {"name": "status", "type": "uint8"},
      {"name": "finalTaskHash", "type": "bytes32"},
      {"name": "completedAt", "type": "uint256"}
    ]
  },
  {
    "type": "function",
    "name": "getAgentTask",
    "stateMutability": "view",
    "inputs": [{"name": "taskHash", "type": "bytes32"}],
    "outputs": [
      {
        "name": "task",
        "type": "tuple",
        "components": [
          {"name": "stage", "type": "uint8"},
          {"name": "taskSeq", "type": "uint256"},
          {"name": "inputHash", "type": "bytes32"},
          {"name": "input", "type": "bytes"},
          {"name": "timestamp", "type": "uint256"},
          {"name": "expiresAt", "type": "uint256"},
          {"name": "prevReplyHashes", "type": "bytes32[]"},
          {"name": "workflowRunId", "type": "bytes32"}
        ]
      },
      {"name": "proven", "type": "bool"}
    ]
  },
  {
    "type": "function",
    "name": "getAgentReply",
    "stateMutability": "view",
    "inputs": [{"name": "replyHash", "type": "bytes32"}],
    "outputs": [
      {
        "name": "reply",
        "type": "tuple",
        "components": [
          {"name": "outputHash", "type": "bytes32"},
          {"name": "output", "type": "bytes"},
          {"name": "timestamp", "type": "uint256"},
          {"name": "replier", "type": "address"},
          {"name": "prevTaskHashes", "type": "bytes32[]"},
          {"name": "workflowRunId", "type": "bytes32"}
        ]
      },
      {"name": "verifier", "type": "address"},
      {"name": "proven", "type": "bool"},
      {"name": "verificationDigest", "type": "bytes32"}
    ]
  },
  {
    "type": "function",
    "name": "onAgentReply",
    "stateMutability": "nonpayable",
    "inputs": [
      {
        "name": "reply",
        "type": "tuple",
        "components": [
          {"name": "outputHash", "type": "bytes32"},
          {"name": "output", "type": "bytes"},
          {"name": "timestamp", "type": "uint256"},
          {"name": "replier", "type": "address"},
          {"name": "prevTaskHashes", "type": "bytes32[]"},
          {"name": "workflowRunId", "type": "bytes32"}
        ]
      }
    ],
    "outputs": []
  },
  {
    "type": "function",
    "name": "onAgentProve",
    "stateMutability": "nonpayable",
    "inputs": [
      {"name": "replyHashes", "type": "bytes32[]"},
      {"name": "proof", "type": "bytes"}
    ],
    "outputs": []
  },
  {
    "type": "event",
    "name": "NewAgentTask",
    "anonymous": false,
    "inputs": [
      {"name": "workflowRunId", "type": "bytes32", "indexed": true},
      {"name": "stage", "type": "uint8", "indexed": true},
      {"name": "taskHash", "type": "bytes32", "indexed": true}
    ]
  }
]`

var (
	workflowABI     abi.ABI
	workflowABIOnce sync.Once
	workflowABIErr  error
)

// AgentWorkflowABI returns the parsed ERC-8301 IAgentWorkflow ABI.
// Parsed once and cached; errors are returned, never panicked.
func AgentWorkflowABI() (abi.ABI, error) {
	workflowABIOnce.Do(func() {
		workflowABI, workflowABIErr = abi.JSON(strings.NewReader(agentWorkflowABIJSON))
	})
	return workflowABI, workflowABIErr
}
