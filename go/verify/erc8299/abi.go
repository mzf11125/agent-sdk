package erc8299

import (
	"strings"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi"
)

// wyriweAttestationABIJSON is the ERC-8299 IWyriweAttestation interface as
// JSON, hand-written to match the interface exactly (see
// agent-ercs/contracts/verify/ERC8299/IWyriweAttestation.sol). Both
// functions are view — verify checks the EIP-712-style signature against
// the known attestor and proofSystem exposes the proof-system identifier.
const wyriweAttestationABIJSON = `[
  {
    "type": "function",
    "name": "verify",
    "stateMutability": "view",
    "inputs": [
      {
        "name": "attestation",
        "type": "tuple",
        "components": [
          {"name": "agentId", "type": "bytes32"},
          {"name": "registry", "type": "address"},
          {"name": "modelHash", "type": "bytes32"},
          {"name": "rawInputHash", "type": "bytes32"},
          {"name": "sanitizationPipelineHash", "type": "bytes32"},
          {"name": "inputHash", "type": "bytes32"},
          {"name": "outputHash", "type": "bytes32"},
          {"name": "timestamp", "type": "uint256"}
        ]
      },
      {"name": "signature", "type": "bytes"}
    ],
    "outputs": [{"name": "valid", "type": "bool"}]
  },
  {
    "type": "function",
    "name": "proofSystem",
    "stateMutability": "view",
    "inputs": [],
    "outputs": [{"name": "", "type": "string"}]
  }
]`

// judgmentExecutionABIJSON is the ERC-8299 IJudgmentExecutionAttestation
// interface as JSON, hand-written to match the interface exactly (see
// agent-ercs/contracts/verify/ERC8299/IJudgmentExecutionAttestation.sol).
const judgmentExecutionABIJSON = `[
  {
    "type": "function",
    "name": "verify",
    "stateMutability": "view",
    "inputs": [
      {
        "name": "attestation",
        "type": "tuple",
        "components": [
          {"name": "agentId", "type": "bytes32"},
          {"name": "registry", "type": "address"},
          {"name": "validatorId", "type": "bytes32"},
          {"name": "rawProposalHash", "type": "bytes32"},
          {"name": "verdictHash", "type": "bytes32"},
          {"name": "executedActionHash", "type": "bytes32"},
          {"name": "verdictTimestamp", "type": "uint256"},
          {"name": "executedTimestamp", "type": "uint256"},
          {"name": "recordPointer", "type": "string"}
        ]
      },
      {"name": "signature", "type": "bytes"}
    ],
    "outputs": [{"name": "valid", "type": "bool"}]
  },
  {
    "type": "function",
    "name": "proofSystem",
    "stateMutability": "view",
    "inputs": [],
    "outputs": [{"name": "", "type": "string"}]
  }
]`

var (
	wyriweABI     abi.ABI
	wyriweABIOnce sync.Once
	wyriweABIErr  error

	judgmentABI     abi.ABI
	judgmentABIOnce sync.Once
	judgmentABIErr  error
)

// WyriweAttestationABI returns the parsed ERC-8299 IWyriweAttestation ABI.
// Parsed once and cached; errors are returned, never panicked.
func WyriweAttestationABI() (abi.ABI, error) {
	wyriweABIOnce.Do(func() {
		wyriweABI, wyriweABIErr = abi.JSON(strings.NewReader(wyriweAttestationABIJSON))
	})
	return wyriweABI, wyriweABIErr
}

// JudgmentExecutionABI returns the parsed ERC-8299
// IJudgmentExecutionAttestation ABI. Parsed once and cached; errors are
// returned, never panicked.
func JudgmentExecutionABI() (abi.ABI, error) {
	judgmentABIOnce.Do(func() {
		judgmentABI, judgmentABIErr = abi.JSON(strings.NewReader(judgmentExecutionABIJSON))
	})
	return judgmentABI, judgmentABIErr
}
