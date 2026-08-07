package erc8203

import (
	"strings"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi"
)

// consultEscrowABIJSON is the ERC-8203 IConsultEscrow interface as JSON,
// hand-written to match the interface exactly (see
// agent-ercs/contracts/settlement/ConsultEscrow/IConsultEscrow.sol).
const consultEscrowABIJSON = `[
  {
    "type": "function",
    "name": "open",
    "stateMutability": "payable",
    "inputs": [
      {"name": "jobId", "type": "bytes32"},
      {"name": "provider", "type": "address"},
      {"name": "attestor", "type": "address"},
      {"name": "deadline", "type": "uint256"}
    ],
    "outputs": []
  },
  {
    "type": "function",
    "name": "release",
    "stateMutability": "nonpayable",
    "inputs": [
      {"name": "jobId", "type": "bytes32"},
      {"name": "resultHash", "type": "bytes32"},
      {"name": "signature", "type": "bytes"}
    ],
    "outputs": []
  },
  {
    "type": "function",
    "name": "refund",
    "stateMutability": "nonpayable",
    "inputs": [{"name": "jobId", "type": "bytes32"}],
    "outputs": []
  },
  {
    "type": "function",
    "name": "jobs",
    "stateMutability": "view",
    "inputs": [{"name": "jobId", "type": "bytes32"}],
    "outputs": [
      {"name": "consumer", "type": "address"},
      {"name": "provider", "type": "address"},
      {"name": "attestor", "type": "address"},
      {"name": "amount", "type": "uint256"},
      {"name": "deadline", "type": "uint256"},
      {"name": "status", "type": "uint8"}
    ]
  },
  {
    "type": "event",
    "name": "Opened",
    "inputs": [
      {"name": "jobId", "type": "bytes32", "indexed": true},
      {"name": "consumer", "type": "address", "indexed": true},
      {"name": "provider", "type": "address", "indexed": true},
      {"name": "attestor", "type": "address", "indexed": false},
      {"name": "amount", "type": "uint256", "indexed": false},
      {"name": "deadline", "type": "uint256", "indexed": false}
    ]
  },
  {
    "type": "event",
    "name": "Released",
    "inputs": [
      {"name": "jobId", "type": "bytes32", "indexed": true},
      {"name": "resultHash", "type": "bytes32", "indexed": false},
      {"name": "commitmentHash", "type": "bytes32", "indexed": false},
      {"name": "provider", "type": "address", "indexed": false},
      {"name": "amount", "type": "uint256", "indexed": false}
    ]
  },
  {
    "type": "event",
    "name": "Refunded",
    "inputs": [
      {"name": "jobId", "type": "bytes32", "indexed": true},
      {"name": "consumer", "type": "address", "indexed": false},
      {"name": "amount", "type": "uint256", "indexed": false}
    ]
  }
]`

var (
	escrowABI     abi.ABI
	escrowABIOnce sync.Once
	escrowABIErr  error
)

// ConsultEscrowABI returns the parsed ERC-8203 IConsultEscrow ABI.
// Parsed once and cached; errors are returned, never panicked.
func ConsultEscrowABI() (abi.ABI, error) {
	escrowABIOnce.Do(func() {
		escrowABI, escrowABIErr = abi.JSON(strings.NewReader(consultEscrowABIJSON))
	})
	return escrowABI, escrowABIErr
}
