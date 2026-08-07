package erc8323

import (
	"strings"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi"
)

// sourceBindingABIJSON is the ERC-8323 IAgentSourceBinding interface as
// JSON, hand-written to match the interface exactly (see
// agent-ercs/contracts/identity/ERC8323/IAgentSourceBinding.sol). It covers
// the write side (registerWithSource), the read side (boundCollection,
// getSourceNFT, hasSourceNFT, isSourceNFTOwnershipValid), the inherited
// ERC-721 ownerOf, the ERC-165 supportsInterface used by
// SupportsSourceBinding, and the SourceNFTLinked event parsed by Register.
const sourceBindingABIJSON = `[
  {
    "type": "function",
    "name": "boundCollection",
    "stateMutability": "view",
    "inputs": [],
    "outputs": [{"name": "", "type": "address"}]
  },
  {
    "type": "function",
    "name": "registerWithSource",
    "stateMutability": "payable",
    "inputs": [{"name": "sourceTokenId", "type": "uint256"}],
    "outputs": [{"name": "agentId", "type": "uint256"}]
  },
  {
    "type": "function",
    "name": "getSourceNFT",
    "stateMutability": "view",
    "inputs": [{"name": "agentId", "type": "uint256"}],
    "outputs": [
      {"name": "sourceContract", "type": "address"},
      {"name": "sourceTokenId", "type": "uint256"}
    ]
  },
  {
    "type": "function",
    "name": "hasSourceNFT",
    "stateMutability": "view",
    "inputs": [{"name": "agentId", "type": "uint256"}],
    "outputs": [{"name": "", "type": "bool"}]
  },
  {
    "type": "function",
    "name": "isSourceNFTOwnershipValid",
    "stateMutability": "view",
    "inputs": [{"name": "agentId", "type": "uint256"}],
    "outputs": [{"name": "", "type": "bool"}]
  },
  {
    "type": "function",
    "name": "ownerOf",
    "stateMutability": "view",
    "inputs": [{"name": "tokenId", "type": "uint256"}],
    "outputs": [{"name": "", "type": "address"}]
  },
  {
    "type": "function",
    "name": "supportsInterface",
    "stateMutability": "view",
    "inputs": [{"name": "interfaceId", "type": "bytes4"}],
    "outputs": [{"name": "", "type": "bool"}]
  },
  {
    "type": "event",
    "name": "SourceNFTLinked",
    "inputs": [
      {"name": "agentId", "type": "uint256", "indexed": true},
      {"name": "sourceContract", "type": "address", "indexed": true},
      {"name": "sourceTokenId", "type": "uint256", "indexed": false}
    ]
  }
]`

var (
	sbABI     abi.ABI
	sbABIOnce sync.Once
	sbABIErr  error
)

// SourceBindingABI returns the parsed ERC-8323 IAgentSourceBinding ABI.
// Parsed once and cached; errors are returned, never panicked.
func SourceBindingABI() (abi.ABI, error) {
	sbABIOnce.Do(func() {
		sbABI, sbABIErr = abi.JSON(strings.NewReader(sourceBindingABIJSON))
	})
	return sbABI, sbABIErr
}
