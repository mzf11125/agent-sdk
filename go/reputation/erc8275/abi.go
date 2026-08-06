package erc8275

import (
	"strings"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi"
)

// agentReputationABIJSON is the ERC-8275 IAgentReputation interface as JSON,
// hand-written to match the interface exactly (see
// agent-ercs/contracts/reputation/ERC8275/IAgentReputation.sol).
const agentReputationABIJSON = `[
  {
    "type": "function",
    "name": "getReputation",
    "stateMutability": "view",
    "inputs": [{"name": "agentId", "type": "bytes32"}],
    "outputs": [
      {"name": "completedOrders", "type": "uint64"},
      {"name": "disputedOrders", "type": "uint64"},
      {"name": "totalVolume", "type": "uint64"},
      {"name": "lastActiveAt", "type": "uint64"},
      {"name": "score", "type": "uint16"}
    ]
  },
  {
    "type": "function",
    "name": "getDecayWeight",
    "stateMutability": "view",
    "inputs": [{"name": "agentId", "type": "bytes32"}],
    "outputs": [{"name": "weight", "type": "uint16"}]
  },
  {
    "type": "function",
    "name": "verifyOutcome",
    "stateMutability": "view",
    "inputs": [
      {"name": "orderId", "type": "bytes32"},
      {"name": "proof", "type": "bytes"}
    ],
    "outputs": [{"name": "valid", "type": "bool"}]
  }
]`

var (
	repABI     abi.ABI
	repABIOnce sync.Once
	repABIErr  error
)

// AgentReputationABI returns the parsed ERC-8275 IAgentReputation ABI.
// Parsed once and cached; errors are returned, never panicked.
func AgentReputationABI() (abi.ABI, error) {
	repABIOnce.Do(func() {
		repABI, repABIErr = abi.JSON(strings.NewReader(agentReputationABIJSON))
	})
	return repABI, repABIErr
}
