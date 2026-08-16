// Package erc8354 implements the ERC-8354 Confidential Agent Policy Verdicts
// SDK.
//
// ERC-8354 defines a pre-execution allow/deny verdict, proven in zero
// knowledge against a policy that is never disclosed on-chain. This package
// provides the pure recompute layer (Layer 2): the canonical action
// commitment and the EIP-712 verdict digest, both reproduced from public
// inputs with no blockchain dependency.
package erc8354

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

// Verdict is the ERC-8354 verdict envelope. Every field is a public input of
// the proving program.
type Verdict struct {
	AgentId          *big.Int
	DomainId         common.Hash
	PolicyRoot       common.Hash
	ActionCommitment common.Hash
	Executor         common.Address
	Expiry           uint64
	Nullifier        common.Hash
	Decision         uint8
	PolicyKind       uint8
}

// MechanismZkSecretPolicy is the source-class tag a Guard writes into the
// ERC-8004 attestation for a consumed verdict: keccak256("zk-secret-policy").
var MechanismZkSecretPolicy = crypto.Keccak256Hash([]byte("zk-secret-policy"))

var (
	verdictTypeHash = crypto.Keccak256Hash([]byte(
		"Verdict(uint256 agentId,bytes32 domainId,bytes32 policyRoot,bytes32 actionCommitment,address executor,uint64 expiry,bytes32 nullifier,uint8 decision,uint8 policyKind)",
	))
	domainTypeHash = crypto.Keccak256Hash([]byte(
		"EIP712Domain(string name,string version,uint256 chainId,address verifyingContract)",
	))
	nameHash    = crypto.Keccak256Hash([]byte("ConfidentialPolicyVerdict"))
	versionHash = crypto.Keccak256Hash([]byte("1"))
)

// ComputeActionCommitment computes the canonical action commitment for a
// policy verdict (ERC-8354 Action commitment).
//
//	actionCommitment = keccak256(abi.encode(chainId, domainId, agentId,
//	    target, value, keccak256(callData), actionNonce))
//
// The Guard recomputes this from the action it is about to execute and
// compares it to Verdict.ActionCommitment. callData is hashed with
// keccak256 first, so an empty callData yields keccak256("") as the
// callDataHash, never bytes32 zero.
func ComputeActionCommitment(
	chainId *big.Int,
	domainId common.Hash,
	agentId *big.Int,
	target common.Address,
	value *big.Int,
	callData []byte,
	actionNonce *big.Int,
) (common.Hash, error) {
	callDataHash := crypto.Keccak256Hash(callData)

	uint256Type, err := abi.NewType("uint256", "", nil)
	if err != nil {
		return common.Hash{}, fmt.Errorf("erc8354: new uint256 type: %w", err)
	}
	bytes32Type, err := abi.NewType("bytes32", "", nil)
	if err != nil {
		return common.Hash{}, fmt.Errorf("erc8354: new bytes32 type: %w", err)
	}
	addressType, err := abi.NewType("address", "", nil)
	if err != nil {
		return common.Hash{}, fmt.Errorf("erc8354: new address type: %w", err)
	}

	args := abi.Arguments{
		{Type: uint256Type},
		{Type: bytes32Type},
		{Type: uint256Type},
		{Type: addressType},
		{Type: uint256Type},
		{Type: bytes32Type},
		{Type: uint256Type},
	}
	packed, err := args.Pack(chainId, domainId, agentId, target, value, callDataHash, actionNonce)
	if err != nil {
		return common.Hash{}, fmt.Errorf("erc8354: pack action commitment: %w", err)
	}
	return crypto.Keccak256Hash(packed), nil
}

// ComputeVerdictDigest computes the EIP-712 digest an executor signs to
// authorize a relayer to submit a verdict on their behalf (ERC-8354
// verdictDigest). The EIP-712 domain is name "ConfidentialPolicyVerdict",
// version "1", plus chainId and verifyingContract.
func ComputeVerdictDigest(v Verdict, chainId *big.Int, verifyingContract common.Address) (common.Hash, error) {
	uint256Type, err := abi.NewType("uint256", "", nil)
	if err != nil {
		return common.Hash{}, fmt.Errorf("erc8354: new uint256 type: %w", err)
	}
	bytes32Type, err := abi.NewType("bytes32", "", nil)
	if err != nil {
		return common.Hash{}, fmt.Errorf("erc8354: new bytes32 type: %w", err)
	}
	addressType, err := abi.NewType("address", "", nil)
	if err != nil {
		return common.Hash{}, fmt.Errorf("erc8354: new address type: %w", err)
	}
	uint64Type, err := abi.NewType("uint64", "", nil)
	if err != nil {
		return common.Hash{}, fmt.Errorf("erc8354: new uint64 type: %w", err)
	}
	uint8Type, err := abi.NewType("uint8", "", nil)
	if err != nil {
		return common.Hash{}, fmt.Errorf("erc8354: new uint8 type: %w", err)
	}

	verdictArgs := abi.Arguments{
		{Type: bytes32Type},
		{Type: uint256Type},
		{Type: bytes32Type},
		{Type: bytes32Type},
		{Type: bytes32Type},
		{Type: addressType},
		{Type: uint64Type},
		{Type: bytes32Type},
		{Type: uint8Type},
		{Type: uint8Type},
	}
	hashStruct, err := verdictArgs.Pack(
		verdictTypeHash,
		v.AgentId,
		v.DomainId,
		v.PolicyRoot,
		v.ActionCommitment,
		v.Executor,
		v.Expiry,
		v.Nullifier,
		v.Decision,
		v.PolicyKind,
	)
	if err != nil {
		return common.Hash{}, fmt.Errorf("erc8354: pack verdict hash struct: %w", err)
	}

	domainArgs := abi.Arguments{
		{Type: bytes32Type},
		{Type: bytes32Type},
		{Type: bytes32Type},
		{Type: uint256Type},
		{Type: addressType},
	}
	domainSeparatorPacked, err := domainArgs.Pack(domainTypeHash, nameHash, versionHash, chainId, verifyingContract)
	if err != nil {
		return common.Hash{}, fmt.Errorf("erc8354: pack domain separator: %w", err)
	}

	// EIP-712: digest = keccak256("\x19\x01" || domainSeparator || hashStruct),
	// where domainSeparator and hashStruct are the 32-byte keccak hashes of
	// the abi-encoded domain and verdict struct respectively.
	domainSeparator := crypto.Keccak256Hash(domainSeparatorPacked)
	hashStructHash := crypto.Keccak256Hash(hashStruct)
	digest := crypto.Keccak256Hash(
		append(append([]byte{0x19, 0x01}, domainSeparator[:]...), hashStructHash[:]...),
	)
	return digest, nil
}
