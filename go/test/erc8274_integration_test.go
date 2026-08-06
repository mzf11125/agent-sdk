// Package test holds integration tests that run against a local anvil node
// deployed via the testkit harness.
//
// Usage:
//
//	testkit/scripts/start-anvil.sh
//	ERC8274_ADDRESSES=$(testkit/scripts/deploy.sh verify/ERC8274 DeployERC8274)
//	ERC8274_ADDRESSES="$ERC8274_ADDRESSES" go test -v ./go/test/ -run TestERC8274
//	testkit/scripts/stop-anvil.sh
//
// deploy.sh prints one contract-creation address per line in broadcast
// order: the proofVerifier mock first, then agentVerifier, then
// agentVerifiable. ERC8274_ADDRESSES carries those three addresses
// separated by newlines (shell command substitution preserves internal
// newlines).
package test

import (
	"os"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/trustless-ai/agent-sdk/go/verify/erc8274"
)

func TestERC8274ProofVerifier(t *testing.T) {
	proofVerifierAddr, _, _ := erc8274Addresses(t)

	rpc, err := ethclient.Dial(anvilRPC)
	if err != nil {
		t.Fatalf("ethclient.Dial(%s): %v", anvilRPC, err)
	}
	defer rpc.Close()

	client := erc8274.NewProofVerifierClient(rpc, proofVerifierAddr)

	system, err := client.ProofSystem()
	if err != nil {
		t.Fatalf("ProofSystem: %v", err)
	}
	t.Logf("proofSystem() = %q", system)
	if system != "mock-test-only" {
		t.Errorf("proofSystem() = %q, want %q", system, "mock-test-only")
	}

	profile, err := client.ProofProfile()
	if err != nil {
		t.Fatalf("ProofProfile: %v", err)
	}
	wantProfile := crypto.Keccak256Hash([]byte("mock-test-only-v1"))
	t.Logf("proofProfile() = %s", profile.Hex())
	if profile != wantProfile {
		t.Errorf("proofProfile() = %s, want %s", profile.Hex(), wantProfile.Hex())
	}

	// The mock accepts proof == keccak256(abi.encodePacked(inputHash,
	// outputHash, metadata)).
	inputHash := crypto.Keccak256Hash([]byte("input"))
	outputHash := crypto.Keccak256Hash([]byte("output"))
	metadata := []byte("model-v1")
	validProof := crypto.Keccak256Hash(encodePackedDigest(inputHash, outputHash, metadata))

	valid, err := client.Verify(inputHash, outputHash, metadata, validProof)
	if err != nil {
		t.Fatalf("Verify(valid proof): %v", err)
	}
	t.Logf("verify(valid) = %v", valid)
	if !valid {
		t.Error("verify(valid proof) = false, want true")
	}

	invalid, err := client.Verify(inputHash, outputHash, metadata, crypto.Keccak256Hash([]byte("garbage")))
	if err != nil {
		t.Fatalf("Verify(invalid proof): %v", err)
	}
	t.Logf("verify(invalid) = %v", invalid)
	if invalid {
		t.Error("verify(invalid proof) = true, want false")
	}
}

func TestERC8274AgentVerifier(t *testing.T) {
	proofVerifierAddr, agentVerifierAddr, _ := erc8274Addresses(t)

	rpc, err := ethclient.Dial(anvilRPC)
	if err != nil {
		t.Fatalf("ethclient.Dial(%s): %v", anvilRPC, err)
	}
	defer rpc.Close()

	key := anvilSignerKey(t)
	client := erc8274.NewAgentVerifierClient(rpc, agentVerifierAddr, key)

	taskId := crypto.Keccak256Hash([]byte("task-1"))
	agentId := crypto.Keccak256Hash([]byte("agent-1"))
	inputHash := crypto.Keccak256Hash([]byte("input"))
	outputHash := crypto.Keccak256Hash([]byte("output"))

	// GetDigest is the expected proof digest of the empty-metadata
	// routing: keccak256(abi.encodePacked(inputHash, outputHash)).
	digest, err := client.GetDigest(inputHash, outputHash)
	if err != nil {
		t.Fatalf("GetDigest: %v", err)
	}
	wantDigest := crypto.Keccak256Hash(encodePackedDigest(inputHash, outputHash, nil))
	t.Logf("getDigest(%s, %s) = %s", inputHash.Hex(), outputHash.Hex(), digest.Hex())
	if digest != wantDigest {
		t.Errorf("GetDigest = %s, want %s", digest.Hex(), wantDigest.Hex())
	}

	// Cross-check: the derived digest is exactly the proof MockProofVerifier
	// accepts when the agent verifier routes with empty metadata, so the
	// proof verifier must accept it for that same empty-metadata call.
	proofClient := erc8274.NewProofVerifierClient(rpc, proofVerifierAddr)
	accepts, err := proofClient.Verify(inputHash, outputHash, nil, digest)
	if err != nil {
		t.Fatalf("ProofVerifier.Verify(derived digest): %v", err)
	}
	t.Logf("verify(empty metadata, derived digest) = %v", accepts)
	if !accepts {
		t.Error("verify(empty metadata, derived digest) = false, want true")
	}

	// VerifyTask derives the proof itself, broadcasts
	// IAgentVerifier.verify, and reports the valid flag from the
	// VerificationCompleted event.
	valid, err := client.VerifyTask(taskId, agentId, inputHash, outputHash)
	if err != nil {
		t.Fatalf("VerifyTask: %v", err)
	}
	t.Logf("verifyTask() = %v", valid)
	if !valid {
		t.Error("VerifyTask = false, want true")
	}
}

func TestERC8274TrustedVerifier(t *testing.T) {
	_, agentVerifierAddr, verifiableAddr := erc8274Addresses(t)

	rpc, err := ethclient.Dial(anvilRPC)
	if err != nil {
		t.Fatalf("ethclient.Dial(%s): %v", anvilRPC, err)
	}
	defer rpc.Close()

	trusted, err := erc8274.GetTrustedVerifier(rpc, verifiableAddr)
	if err != nil {
		t.Fatalf("GetTrustedVerifier: %v", err)
	}
	t.Logf("agentVerifier() = %s", trusted.Hex())
	if trusted != agentVerifierAddr {
		t.Errorf("agentVerifier() = %s, want %s", trusted.Hex(), agentVerifierAddr.Hex())
	}
}

// erc8274Addresses reads the ERC8274_ADDRESSES env var set from deploy.sh
// output: three addresses, newline-separated in broadcast order
// (proofVerifier, agentVerifier, agentVerifiable).
func erc8274Addresses(t *testing.T) (common.Address, common.Address, common.Address) {
	t.Helper()
	fields := strings.Fields(os.Getenv("ERC8274_ADDRESSES"))
	if len(fields) != 3 {
		t.Fatalf("ERC8274_ADDRESSES must hold 3 addresses (got %d) — deploy first via testkit/scripts/deploy.sh verify/ERC8274 DeployERC8274", len(fields))
	}
	return common.HexToAddress(fields[0]), common.HexToAddress(fields[1]), common.HexToAddress(fields[2])
}

// encodePackedDigest replicates Solidity's abi.encodePacked(bytes32,
// bytes32, bytes): the fixed members are concatenated raw and the dynamic
// bytes tail is appended as-is (no length prefix). The mock's expected
// digest is keccak256 of exactly this byte string.
func encodePackedDigest(a, b common.Hash, tail []byte) []byte {
	buf := make([]byte, 0, 64+len(tail))
	buf = append(buf, a[:]...)
	buf = append(buf, b[:]...)
	buf = append(buf, tail...)
	return buf
}
