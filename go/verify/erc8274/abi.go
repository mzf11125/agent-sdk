package erc8274

import (
	"strings"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi"
)

// proofVerifierABIJSON is the ERC-8274 IProofVerifier interface as JSON,
// hand-written to match the interface exactly (see
// agent-ercs/contracts/verify/ERC8274/IProofVerifier.sol). verify is the
// stateless cryptographic proof check; proofSystem/proofProfile expose the
// proof-system identifier and compact profile hash.
const proofVerifierABIJSON = `[
  {
    "type": "function",
    "name": "verify",
    "stateMutability": "nonpayable",
    "inputs": [
      {"name": "inputHash", "type": "bytes32"},
      {"name": "outputHash", "type": "bytes32"},
      {"name": "metadata", "type": "bytes"},
      {"name": "proof", "type": "bytes"}
    ],
    "outputs": [{"name": "valid", "type": "bool"}]
  },
  {
    "type": "function",
    "name": "proofSystem",
    "stateMutability": "view",
    "inputs": [],
    "outputs": [{"name": "", "type": "string"}]
  },
  {
    "type": "function",
    "name": "proofProfile",
    "stateMutability": "view",
    "inputs": [],
    "outputs": [{"name": "", "type": "bytes32"}]
  }
]`

// agentVerifierABIJSON is the ERC-8274 IAgentVerifier interface as JSON,
// hand-written to match the interface exactly (see
// agent-ercs/contracts/verify/ERC8274/IAgentVerifier.sol). verify is
// stateful — it emits VerificationCompleted with the digest preimage — so
// it must be broadcast, unlike IProofVerifier.verify.
const agentVerifierABIJSON = `[
  {
    "type": "function",
    "name": "verify",
    "stateMutability": "nonpayable",
    "inputs": [
      {"name": "taskId", "type": "bytes32"},
      {"name": "agentId", "type": "bytes32"},
      {"name": "inputHash", "type": "bytes32"},
      {"name": "outputHash", "type": "bytes32"},
      {"name": "proof", "type": "bytes"}
    ],
    "outputs": [
      {"name": "valid", "type": "bool"},
      {"name": "verificationDigest", "type": "bytes32"}
    ]
  },
  {
    "type": "event",
    "name": "VerificationCompleted",
    "inputs": [
      {"name": "taskId", "type": "bytes32", "indexed": true},
      {"name": "agentId", "type": "bytes32", "indexed": true},
      {"name": "inputHash", "type": "bytes32", "indexed": false},
      {"name": "outputHash", "type": "bytes32", "indexed": false},
      {"name": "valid", "type": "bool", "indexed": false},
      {"name": "verificationDigest", "type": "bytes32", "indexed": false}
    ]
  }
]`

// agentVerifiableABIJSON is the ERC-8274 IAgentVerifiable interface as
// JSON, hand-written to match the interface exactly (see
// agent-ercs/contracts/verify/ERC8274/IAgentVerifiable.sol). A settlement
// or execution contract implements it to declare which IAgentVerifier it
// trusts.
const agentVerifiableABIJSON = `[
  {
    "type": "function",
    "name": "agentVerifier",
    "stateMutability": "view",
    "inputs": [],
    "outputs": [{"name": "", "type": "address"}]
  },
  {
    "type": "event",
    "name": "AgentVerifierUpdated",
    "inputs": [
      {"name": "oldVerifier", "type": "address", "indexed": true},
      {"name": "newVerifier", "type": "address", "indexed": true}
    ]
  }
]`

var (
	proofVerifierABI     abi.ABI
	proofVerifierABIOnce sync.Once
	proofVerifierABIErr  error

	agentVerifierABI     abi.ABI
	agentVerifierABIOnce sync.Once
	agentVerifierABIErr  error

	agentVerifiableABI     abi.ABI
	agentVerifiableABIOnce sync.Once
	agentVerifiableABIErr  error
)

// ProofVerifierABI returns the parsed ERC-8274 IProofVerifier ABI. Parsed
// once and cached; errors are returned, never panicked.
func ProofVerifierABI() (abi.ABI, error) {
	proofVerifierABIOnce.Do(func() {
		proofVerifierABI, proofVerifierABIErr = abi.JSON(strings.NewReader(proofVerifierABIJSON))
	})
	return proofVerifierABI, proofVerifierABIErr
}

// AgentVerifierABI returns the parsed ERC-8274 IAgentVerifier ABI. Parsed
// once and cached; errors are returned, never panicked.
func AgentVerifierABI() (abi.ABI, error) {
	agentVerifierABIOnce.Do(func() {
		agentVerifierABI, agentVerifierABIErr = abi.JSON(strings.NewReader(agentVerifierABIJSON))
	})
	return agentVerifierABI, agentVerifierABIErr
}

// AgentVerifiableABI returns the parsed ERC-8274 IAgentVerifiable ABI.
// Parsed once and cached; errors are returned, never panicked.
func AgentVerifiableABI() (abi.ABI, error) {
	agentVerifiableABIOnce.Do(func() {
		agentVerifiableABI, agentVerifiableABIErr = abi.JSON(strings.NewReader(agentVerifiableABIJSON))
	})
	return agentVerifiableABI, agentVerifiableABIErr
}
