package erc8312

import (
	"strings"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi"
)

// boundedAgentActionABIJSON is the ERC-8312 IBoundedAgentAction interface as
// JSON, hand-written to match the interface exactly (see
// agent-ercs/contracts/metering/ERC8312/IBoundedAgentAction.sol). Status is
// an enum and encodes as uint8; the getEnvelope tuple component names match
// the Solidity struct so go-ethereum can unpack them into Envelope.
const boundedAgentActionABIJSON = `[
  {
    "type": "function",
    "name": "registerEnvelope",
    "stateMutability": "nonpayable",
    "inputs": [
      {"name": "principal", "type": "address"},
      {"name": "capabilityRoot", "type": "bytes32"},
      {"name": "expiresAt", "type": "uint64"},
      {"name": "initData", "type": "bytes"}
    ],
    "outputs": [{"name": "id", "type": "bytes32"}]
  },
  {
    "type": "function",
    "name": "getEnvelope",
    "stateMutability": "view",
    "inputs": [{"name": "id", "type": "bytes32"}],
    "outputs": [
      {
        "name": "envelope",
        "type": "tuple",
        "components": [
          {"name": "id", "type": "bytes32"},
          {"name": "principal", "type": "address"},
          {"name": "capabilityRoot", "type": "bytes32"},
          {"name": "cursorRoot", "type": "bytes32"},
          {"name": "createdAt", "type": "uint64"},
          {"name": "expiresAt", "type": "uint64"},
          {"name": "status", "type": "uint8"}
        ]
      }
    ]
  },
  {
    "type": "function",
    "name": "getCursor",
    "stateMutability": "view",
    "inputs": [{"name": "id", "type": "bytes32"}],
    "outputs": [{"name": "cursorRoot", "type": "bytes32"}]
  },
  {
    "type": "function",
    "name": "getStatus",
    "stateMutability": "view",
    "inputs": [{"name": "id", "type": "bytes32"}],
    "outputs": [{"name": "status", "type": "uint8"}]
  },
  {
    "type": "function",
    "name": "isActive",
    "stateMutability": "view",
    "inputs": [{"name": "id", "type": "bytes32"}],
    "outputs": [{"name": "active", "type": "bool"}]
  },
  {
    "type": "function",
    "name": "advanceCursor",
    "stateMutability": "nonpayable",
    "inputs": [
      {"name": "id", "type": "bytes32"},
      {"name": "witness", "type": "bytes"}
    ],
    "outputs": [{"name": "newCursor", "type": "bytes32"}]
  },
  {
    "type": "function",
    "name": "setStatus",
    "stateMutability": "nonpayable",
    "inputs": [
      {"name": "id", "type": "bytes32"},
      {"name": "newStatus", "type": "uint8"}
    ],
    "outputs": []
  },
  {
    "type": "event",
    "name": "EnvelopeRegistered",
    "anonymous": false,
    "inputs": [
      {"name": "id", "type": "bytes32", "indexed": true},
      {"name": "principal", "type": "address", "indexed": true},
      {"name": "capabilityRoot", "type": "bytes32", "indexed": true}
    ]
  },
  {
    "type": "event",
    "name": "EnvelopeAdvanced",
    "anonymous": false,
    "inputs": [
      {"name": "id", "type": "bytes32", "indexed": true},
      {"name": "prevCursor", "type": "bytes32", "indexed": false},
      {"name": "newCursor", "type": "bytes32", "indexed": false}
    ]
  },
  {
    "type": "event",
    "name": "EnvelopeStatusChanged",
    "anonymous": false,
    "inputs": [
      {"name": "id", "type": "bytes32", "indexed": true},
      {"name": "fromStatus", "type": "uint8", "indexed": false},
      {"name": "toStatus", "type": "uint8", "indexed": false}
    ]
  }
]`

// budgetSubstrateABIJSON is the ERC-8312 IBudgetSubstrate typed read surface
// as JSON (see
// agent-ercs/contracts/metering/ERC8312/IBudgetSubstrate.sol). Only the
// three profile functions; the inherited IBoundedAgentAction surface is
// served by BoundedAgentActionClient against the same contract address.
const budgetSubstrateABIJSON = `[
  {
    "type": "function",
    "name": "bound",
    "stateMutability": "view",
    "inputs": [{"name": "id", "type": "bytes32"}],
    "outputs": [
      {"name": "cap", "type": "uint256"},
      {"name": "asset", "type": "address"}
    ]
  },
  {
    "type": "function",
    "name": "spent",
    "stateMutability": "view",
    "inputs": [{"name": "id", "type": "bytes32"}],
    "outputs": [{"name": "spent", "type": "uint256"}]
  },
  {
    "type": "function",
    "name": "remaining",
    "stateMutability": "view",
    "inputs": [{"name": "id", "type": "bytes32"}],
    "outputs": [{"name": "remaining", "type": "uint256"}]
  }
]`

// contestableEnvelopeABIJSON is the ERC-8312 IContestableEnvelope extension
// as JSON (see
// agent-ercs/contracts/metering/ERC8312/IContestableEnvelope.sol).
const contestableEnvelopeABIJSON = `[
  {
    "type": "function",
    "name": "contest",
    "stateMutability": "nonpayable",
    "inputs": [
      {"name": "id", "type": "bytes32"},
      {"name": "evidence", "type": "bytes"}
    ],
    "outputs": []
  },
  {
    "type": "function",
    "name": "resolve",
    "stateMutability": "nonpayable",
    "inputs": [
      {"name": "id", "type": "bytes32"},
      {"name": "outcome", "type": "uint8"},
      {"name": "resolution", "type": "bytes"}
    ],
    "outputs": []
  },
  {
    "type": "event",
    "name": "EnvelopeContested",
    "anonymous": false,
    "inputs": [
      {"name": "id", "type": "bytes32", "indexed": true},
      {"name": "challenger", "type": "address", "indexed": true}
    ]
  },
  {
    "type": "event",
    "name": "EnvelopeResolved",
    "anonymous": false,
    "inputs": [
      {"name": "id", "type": "bytes32", "indexed": true},
      {"name": "outcome", "type": "uint8", "indexed": false}
    ]
  }
]`

var (
	boundedAgentActionABI     abi.ABI
	boundedAgentActionABIOnce sync.Once
	boundedAgentActionABIErr  error
)

// BoundedAgentActionABI returns the parsed ERC-8312 IBoundedAgentAction ABI.
// Parsed once and cached; errors are returned, never panicked.
func BoundedAgentActionABI() (abi.ABI, error) {
	boundedAgentActionABIOnce.Do(func() {
		boundedAgentActionABI, boundedAgentActionABIErr = abi.JSON(strings.NewReader(boundedAgentActionABIJSON))
	})
	return boundedAgentActionABI, boundedAgentActionABIErr
}

var (
	budgetSubstrateABI     abi.ABI
	budgetSubstrateABIOnce sync.Once
	budgetSubstrateABIErr  error
)

// BudgetSubstrateABI returns the parsed ERC-8312 IBudgetSubstrate ABI.
// Parsed once and cached; errors are returned, never panicked.
func BudgetSubstrateABI() (abi.ABI, error) {
	budgetSubstrateABIOnce.Do(func() {
		budgetSubstrateABI, budgetSubstrateABIErr = abi.JSON(strings.NewReader(budgetSubstrateABIJSON))
	})
	return budgetSubstrateABI, budgetSubstrateABIErr
}

var (
	contestableEnvelopeABI     abi.ABI
	contestableEnvelopeABIOnce sync.Once
	contestableEnvelopeABIErr  error
)

// ContestableEnvelopeABI returns the parsed ERC-8312 IContestableEnvelope
// ABI. Parsed once and cached; errors are returned, never panicked.
func ContestableEnvelopeABI() (abi.ABI, error) {
	contestableEnvelopeABIOnce.Do(func() {
		contestableEnvelopeABI, contestableEnvelopeABIErr = abi.JSON(strings.NewReader(contestableEnvelopeABIJSON))
	})
	return contestableEnvelopeABI, contestableEnvelopeABIErr
}
