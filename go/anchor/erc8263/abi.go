package erc8263

import (
	"strings"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi"
)

// onChainProofABIJSON is the ERC-8263 IOnChainProof interface as JSON,
// hand-written to match the interface exactly (see
// agent-ercs/contracts/anchor/ERC8263/IOnChainProof.sol).
const onChainProofABIJSON = `[
  {
    "type": "function",
    "name": "anchor",
    "stateMutability": "nonpayable",
    "inputs": [
      {"name": "agentIdScheme", "type": "uint8"},
      {"name": "agentId", "type": "bytes32"},
      {"name": "proofHash", "type": "bytes32"}
    ],
    "outputs": []
  },
  {
    "type": "function",
    "name": "anchorWithAux",
    "stateMutability": "nonpayable",
    "inputs": [
      {"name": "agentIdScheme", "type": "uint8"},
      {"name": "agentId", "type": "bytes32"},
      {"name": "proofHash", "type": "bytes32"},
      {"name": "aux", "type": "bytes"}
    ],
    "outputs": []
  },
  {
    "type": "event",
    "name": "AnchorProof",
    "inputs": [
      {"name": "agentIdScheme", "type": "uint8", "indexed": false},
      {"name": "agentId", "type": "bytes32", "indexed": true},
      {"name": "proofHash", "type": "bytes32", "indexed": true},
      {"name": "operator", "type": "address", "indexed": true},
      {"name": "aux", "type": "bytes", "indexed": false}
    ]
  }
]`

var (
	ocpABI     abi.ABI
	ocpABIOnce sync.Once
	ocpABIErr  error
)

// OnChainProofABI returns the parsed ERC-8263 IOnChainProof ABI. Parsed once
// and cached; errors are returned, never panicked.
func OnChainProofABI() (abi.ABI, error) {
	ocpABIOnce.Do(func() {
		ocpABI, ocpABIErr = abi.JSON(strings.NewReader(onChainProofABIJSON))
	})
	return ocpABI, ocpABIErr
}
