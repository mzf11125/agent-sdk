package erc8004

import (
	"strings"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi"
)

// identityRegistryABIJSON is the ERC-8004 IIdentityRegistry interface as
// JSON, hand-written to match the interface exactly (see
// agent-ercs/contracts/identity/ERC8004/IIdentityRegistry.sol). It covers
// the view functions read by IdentityRegistryClient plus the two-argument
// register overload, which integration tests use to mint an agent before
// reading its state.
const identityRegistryABIJSON = `[
  {
    "type": "function",
    "name": "register",
    "stateMutability": "nonpayable",
    "inputs": [
      {"name": "agentURI", "type": "string"},
      {
        "name": "metadata",
        "type": "tuple[]",
        "components": [
          {"name": "metadataKey", "type": "string"},
          {"name": "metadataValue", "type": "bytes"}
        ]
      }
    ],
    "outputs": [{"name": "agentId", "type": "uint256"}]
  },
  {
    "type": "function",
    "name": "tokenURI",
    "stateMutability": "view",
    "inputs": [{"name": "agentId", "type": "uint256"}],
    "outputs": [{"name": "", "type": "string"}]
  },
  {
    "type": "function",
    "name": "getMetadata",
    "stateMutability": "view",
    "inputs": [
      {"name": "agentId", "type": "uint256"},
      {"name": "metadataKey", "type": "string"}
    ],
    "outputs": [{"name": "", "type": "bytes"}]
  },
  {
    "type": "function",
    "name": "getAgentWallet",
    "stateMutability": "view",
    "inputs": [{"name": "agentId", "type": "uint256"}],
    "outputs": [{"name": "", "type": "address"}]
  },
  {
    "type": "function",
    "name": "ownerOf",
    "stateMutability": "view",
    "inputs": [{"name": "tokenId", "type": "uint256"}],
    "outputs": [{"name": "", "type": "address"}]
  }
]`

var (
	idABI     abi.ABI
	idABIOnce sync.Once
	idABIErr  error
)

// IdentityRegistryABI returns the parsed ERC-8004 IIdentityRegistry ABI.
// Parsed once and cached; errors are returned, never panicked.
func IdentityRegistryABI() (abi.ABI, error) {
	idABIOnce.Do(func() {
		idABI, idABIErr = abi.JSON(strings.NewReader(identityRegistryABIJSON))
	})
	return idABI, idABIErr
}
