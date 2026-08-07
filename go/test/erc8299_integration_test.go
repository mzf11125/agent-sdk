// Package test holds integration tests that run against a local anvil node
// deployed via the testkit harness.
//
// Usage:
//
//	testkit/scripts/start-anvil.sh
//	ERC8299_ADDRESSES=$(testkit/scripts/deploy.sh verify/ERC8299 DeployERC8299)
//	ERC8299_ADDRESSES="$ERC8299_ADDRESSES" go test -v ./go/test/ -run TestERC8299
//	testkit/scripts/stop-anvil.sh
//
// deploy.sh prints one contract-creation address per line in broadcast
// order: the WyriweAttestation mock first, then the JudgmentExecution
// mock. ERC8299_ADDRESSES carries those two addresses separated by
// newlines (shell command substitution preserves internal newlines).
package test

import (
	"context"
	"math/big"
	"os"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/trustless-ai/agent-sdk/go/verify/erc8299"
)

func TestERC8299WyriweAttestationVerify(t *testing.T) {
	wyriweAddr, _ := erc8299Addresses(t)

	rpc, err := ethclient.Dial(anvilRPC)
	if err != nil {
		t.Fatalf("ethclient.Dial(%s): %v", anvilRPC, err)
	}
	defer rpc.Close()

	client := erc8299.NewWyriweAttestationClient(rpc, wyriweAddr)

	system, err := client.ProofSystem(context.Background())
	if err != nil {
		t.Fatalf("ProofSystem: %v", err)
	}
	t.Logf("proofSystem() = %q", system)
	if system != "attestation/wyriwe" {
		t.Errorf("proofSystem() = %q, want %q", system, "attestation/wyriwe")
	}

	attestation := erc8299.WyriweAttestation{
		AgentId:                  crypto.Keccak256Hash([]byte("agent-1")),
		Registry:                 common.HexToAddress("0x000000000000000000000000000000000000c0de"),
		ModelHash:                crypto.Keccak256Hash([]byte("model-v1")),
		RawInputHash:             crypto.Keccak256Hash([]byte("raw input")),
		SanitizationPipelineHash: crypto.Keccak256Hash([]byte("sanitization pipeline")),
		InputHash:                crypto.Keccak256Hash([]byte("sanitized input")),
		OutputHash:               crypto.Keccak256Hash([]byte("model output")),
		Timestamp:                big.NewInt(1000000),
	}

	// The mock accepts signature == raw keccak256(abi.encode(attestation)).
	valid, err := client.Verify(context.Background(), attestation, wyriweSignature(t, attestation))
	if err != nil {
		t.Fatalf("Verify(valid signature): %v", err)
	}
	t.Logf("verify(valid) = %v", valid)
	if !valid {
		t.Error("verify(valid signature) = false, want true")
	}

	invalid, err := client.Verify(context.Background(), attestation, []byte{0xff, 0x00, 0xaa, 0xbb})
	if err != nil {
		t.Fatalf("Verify(invalid signature): %v", err)
	}
	t.Logf("verify(invalid) = %v", invalid)
	if invalid {
		t.Error("verify(invalid signature) = true, want false")
	}
}

func TestERC8299JudgmentExecutionVerify(t *testing.T) {
	_, judgmentAddr := erc8299Addresses(t)

	rpc, err := ethclient.Dial(anvilRPC)
	if err != nil {
		t.Fatalf("ethclient.Dial(%s): %v", anvilRPC, err)
	}
	defer rpc.Close()

	client := erc8299.NewJudgmentExecutionClient(rpc, judgmentAddr)

	system, err := client.ProofSystem(context.Background())
	if err != nil {
		t.Fatalf("ProofSystem: %v", err)
	}
	t.Logf("proofSystem() = %q", system)
	if system != "attestation/judgment" {
		t.Errorf("proofSystem() = %q, want %q", system, "attestation/judgment")
	}

	attestation := erc8299.JudgmentExecutionAttestation{
		AgentId:            crypto.Keccak256Hash([]byte("executing-agent")),
		Registry:           common.HexToAddress("0x000000000000000000000000000000000000c0de"),
		ValidatorId:        crypto.Keccak256Hash([]byte("validator-1")),
		RawProposalHash:    crypto.Keccak256Hash([]byte("proposal")),
		VerdictHash:        crypto.Keccak256Hash([]byte("verdict")),
		ExecutedActionHash: crypto.Keccak256Hash([]byte("executed action")),
		VerdictTimestamp:   big.NewInt(1000000),
		ExecutedTimestamp:  big.NewInt(2000000),
		RecordPointer:      "ipfs://QmJudgmentRecord",
	}

	valid, err := client.Verify(context.Background(), attestation, judgmentSignature(t, attestation))
	if err != nil {
		t.Fatalf("Verify(valid signature): %v", err)
	}
	t.Logf("verify(valid) = %v", valid)
	if !valid {
		t.Error("verify(valid signature) = false, want true")
	}

	invalid, err := client.Verify(context.Background(), attestation, []byte{0xde, 0xad, 0xbe, 0xef})
	if err != nil {
		t.Fatalf("Verify(invalid signature): %v", err)
	}
	t.Logf("verify(invalid) = %v", invalid)
	if invalid {
		t.Error("verify(invalid signature) = true, want false")
	}
}

// erc8299Addresses reads the ERC8299_ADDRESSES env var set from
// deploy.sh output: two addresses, newline-separated in broadcast order
// (wyriweAttestation, then judgmentExecutionAttestation).
func erc8299Addresses(t *testing.T) (common.Address, common.Address) {
	t.Helper()
	fields := strings.Fields(os.Getenv("ERC8299_ADDRESSES"))
	if len(fields) != 2 {
		t.Fatalf("ERC8299_ADDRESSES must hold 2 addresses (got %d) — deploy first via testkit/scripts/deploy.sh verify/ERC8299 DeployERC8299", len(fields))
	}
	return common.HexToAddress(fields[0]), common.HexToAddress(fields[1])
}

// wyriweSignature replicates the mock's signature check:
// keccak256(abi.encode(attestation)). The struct is packed as a tuple type
// (all members static), producing the bare 256-byte member concatenation —
// identical to Solidity's abi.encode(struct).
func wyriweSignature(t *testing.T, a erc8299.WyriweAttestation) []byte {
	t.Helper()
	args, err := erc8299TupleArgument(t, "WyriweAttestation", []abi.ArgumentMarshaling{
		{Name: "agentId", Type: "bytes32"},
		{Name: "registry", Type: "address"},
		{Name: "modelHash", Type: "bytes32"},
		{Name: "rawInputHash", Type: "bytes32"},
		{Name: "sanitizationPipelineHash", Type: "bytes32"},
		{Name: "inputHash", Type: "bytes32"},
		{Name: "outputHash", Type: "bytes32"},
		{Name: "timestamp", Type: "uint256"},
	})
	if err != nil {
		t.Fatalf("build wyriwe tuple arg: %v", err)
	}
	packed, err := args.Pack(a)
	if err != nil {
		t.Fatalf("pack wyriwe attestation: %v", err)
	}
	return crypto.Keccak256(packed)
}

// judgmentSignature replicates the mock's signature check:
// keccak256(abi.encode(attestation)). recordPointer makes the tuple
// dynamic, so the pack prepends the 32-byte offset word before the
// head+tail block — exactly Solidity's abi.encode(struct) for a struct
// with a dynamic member.
func judgmentSignature(t *testing.T, a erc8299.JudgmentExecutionAttestation) []byte {
	t.Helper()
	args, err := erc8299TupleArgument(t, "JudgmentExecutionAttestation", []abi.ArgumentMarshaling{
		{Name: "agentId", Type: "bytes32"},
		{Name: "registry", Type: "address"},
		{Name: "validatorId", Type: "bytes32"},
		{Name: "rawProposalHash", Type: "bytes32"},
		{Name: "verdictHash", Type: "bytes32"},
		{Name: "executedActionHash", Type: "bytes32"},
		{Name: "verdictTimestamp", Type: "uint256"},
		{Name: "executedTimestamp", Type: "uint256"},
		{Name: "recordPointer", Type: "string"},
	})
	if err != nil {
		t.Fatalf("build judgment tuple arg: %v", err)
	}
	packed, err := args.Pack(a)
	if err != nil {
		t.Fatalf("pack judgment attestation: %v", err)
	}
	return crypto.Keccak256(packed)
}

// erc8299TupleArgument builds a single-argument abi.Arguments whose type
// is the given tuple, mirroring the Solidity struct member order (see
// IWyriweAttestation.sol / IJudgmentExecutionAttestation.sol). Packing the
// struct value with it reproduces abi.encode(struct).
func erc8299TupleArgument(t *testing.T, name string, components []abi.ArgumentMarshaling) (abi.Arguments, error) {
	t.Helper()
	typ, err := abi.NewType("tuple", name, components)
	if err != nil {
		return nil, err
	}
	return abi.Arguments{{Type: typ}}, nil
}
