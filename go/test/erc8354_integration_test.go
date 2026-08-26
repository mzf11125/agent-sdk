// Package test holds integration tests that run against a local anvil node
// deployed via the testkit harness.
//
// Usage:
//
//	testkit/scripts/start-anvil.sh
//	ERC8354_ADDRESSES=$(testkit/scripts/deploy.sh verify/ERC8354 DeployERC8354)
//	ERC8354_ADDRESSES="$ERC8354_ADDRESSES" go test -v ./go/test/ -run TestERC8354
//	testkit/scripts/stop-anvil.sh
//
// deploy.sh prints one contract-creation address per line in broadcast
// order: verifier, then registry, then guard.
package test

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/trustless-ai/agent-sdk/go/verify/erc8354"
)

const registryAdminABIJSON = `[
  {
    "type": "function",
    "name": "registerDomain",
    "stateMutability": "nonpayable",
    "inputs": [
      {"name": "domainId", "type": "bytes32"},
      {"name": "registrar", "type": "address"},
      {"name": "verifier", "type": "address"},
      {"name": "programKey", "type": "bytes32"},
      {"name": "maxRootAge", "type": "uint64"}
    ],
    "outputs": []
  },
  {
    "type": "function",
    "name": "updateRoot",
    "stateMutability": "nonpayable",
    "inputs": [
      {"name": "domainId", "type": "bytes32"},
      {"name": "newRoot", "type": "bytes32"}
    ],
    "outputs": []
  }
]`

const mockVerifierABIJSON = `[
  {
    "type": "function",
    "name": "setResult",
    "stateMutability": "nonpayable",
    "inputs": [{"name": "r", "type": "bool"}],
    "outputs": []
  }
]`

func erc8354Addresses(t *testing.T) (verifier, registry, guard common.Address) {
	t.Helper()
	raw := os.Getenv("ERC8354_ADDRESSES")
	if raw == "" {
		t.Fatal("ERC8354_ADDRESSES not set, deploy first via testkit/scripts/deploy.sh verify/ERC8354 DeployERC8354")
	}
	lines := strings.Fields(raw)
	if len(lines) != 3 {
		t.Fatalf("ERC8354_ADDRESSES must hold 3 addresses, got %d", len(lines))
	}
	return common.HexToAddress(lines[0]), common.HexToAddress(lines[1]), common.HexToAddress(lines[2])
}

// anvilPrivateKeyIndexed returns the anvil account at the given index from
// testkit/.anvil-accounts.json. The file always lists every anvil account, so
// it is the reliable source regardless of which account CI surfaces via
// ANVIL_KEY.
func anvilPrivateKeyIndexed(index int) (*ecdsa.PrivateKey, error) {
	path := filepath.Join("..", "..", "testkit", ".anvil-accounts.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	var parsed struct {
		Accounts []struct {
			PrivateKey string `json:"privateKey"`
		} `json:"accounts"`
	}
	if err := json.Unmarshal(data, &parsed); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	if index >= len(parsed.Accounts) || parsed.Accounts[index].PrivateKey == "" {
		return nil, fmt.Errorf("%s has no accounts[%d].privateKey", path, index)
	}
	privHex := strings.TrimPrefix(parsed.Accounts[index].PrivateKey, "0x")
	key, err := crypto.HexToECDSA(privHex)
	if err != nil {
		return nil, fmt.Errorf("decode private key: %w", err)
	}
	return key, nil
}

// domainRootPair derives a unique domain id and policy root from seed so that
// each test registers its own domain. A shared fixed domain id collides, and
// the second registerDomain reverts with the registry's DomainExists error.
func domainRootPair(seed string) (common.Hash, common.Hash) {
	domainID := crypto.Keccak256Hash([]byte(seed))
	root := crypto.Keccak256Hash([]byte("root-v1-" + seed))
	return domainID, root
}

func setupDomain(t *testing.T, rpc *ethclient.Client, registry common.Address, verifier common.Address, domainID common.Hash, root common.Hash) {
	t.Helper()
	programKey := crypto.Keccak256Hash([]byte("interpreter-vkey"))

	key, err := anvilPrivateKey()
	if err != nil {
		t.Fatalf("read anvil signer key: %v", err)
	}
	ctx := context.Background()
	chainID, err := rpc.ChainID(ctx)
	if err != nil {
		t.Fatalf("ChainID: %v", err)
	}
	auth, err := bind.NewKeyedTransactorWithChainID(key, chainID)
	if err != nil {
		t.Fatalf("NewKeyedTransactorWithChainID: %v", err)
	}
	parsed, err := abi.JSON(strings.NewReader(registryAdminABIJSON))
	if err != nil {
		t.Fatalf("parse registry admin ABI: %v", err)
	}
	bound := bind.NewBoundContract(registry, parsed, rpc, rpc, rpc)

	registrar := common.HexToAddress("0x000000000000000000000000000000000000a11c")

	// Wait for each setup transaction to mine before issuing the next one. The
	// next write depends on the prior one having landed (updateRoot requires
	// the domain to exist), so broadcasting both without waiting is a timing
	// race that surfaces only when anvil mines slowly.
	registerTx, err := bound.Transact(auth, "registerDomain", domainID, registrar, verifier, programKey, uint64(3600))
	if err != nil {
		t.Fatalf("registerDomain: %v", err)
	}
	if _, err := bind.WaitMined(ctx, rpc, registerTx); err != nil {
		t.Fatalf("wait for registerDomain to mine: %v", err)
	}

	updateTx, err := bound.Transact(auth, "updateRoot", domainID, root)
	if err != nil {
		t.Fatalf("updateRoot: %v", err)
	}
	if _, err := bind.WaitMined(ctx, rpc, updateTx); err != nil {
		t.Fatalf("wait for updateRoot to mine: %v", err)
	}
}

func setVerifierResult(t *testing.T, rpc *ethclient.Client, verifier common.Address, result bool) {
	t.Helper()
	key, err := anvilPrivateKey()
	if err != nil {
		t.Fatalf("read anvil signer key: %v", err)
	}
	ctx := context.Background()
	chainID, err := rpc.ChainID(ctx)
	if err != nil {
		t.Fatalf("ChainID: %v", err)
	}
	auth, err := bind.NewKeyedTransactorWithChainID(key, chainID)
	if err != nil {
		t.Fatalf("NewKeyedTransactorWithChainID: %v", err)
	}
	parsed, err := abi.JSON(strings.NewReader(mockVerifierABIJSON))
	if err != nil {
		t.Fatalf("parse mock verifier ABI: %v", err)
	}
	bound := bind.NewBoundContract(verifier, parsed, rpc, rpc, rpc)
	tx, err := bound.Transact(auth, "setResult", result)
	if err != nil {
		t.Fatalf("setResult: %v", err)
	}
	if _, err := bind.WaitMined(ctx, rpc, tx); err != nil {
		t.Fatalf("wait for setResult to mine: %v", err)
	}
}

// signVerdictDigest produces a canonical 65 byte signature over digest with
// the recover id adjusted to the 27/28 range expected by the ECDSA v handling
// in OpenZeppelin SignatureChecker.
func signVerdictDigest(t *testing.T, key *ecdsa.PrivateKey, digest common.Hash) []byte {
	t.Helper()
	sig, err := crypto.Sign(digest[:], key)
	if err != nil {
		t.Fatalf("sign verdict digest: %v", err)
	}
	sig[64] += 27
	return sig
}

func TestERC8354ConfidentialPolicyVerdict(t *testing.T) {
	verifier, registry, guard := erc8354Addresses(t)

	rpc, err := ethclient.Dial(anvilRPC)
	if err != nil {
		t.Fatalf("ethclient.Dial(%s): %v", anvilRPC, err)
	}
	defer rpc.Close()

	domainID, root := domainRootPair("test-domain-verify")
	setupDomain(t, rpc, registry, verifier, domainID, root)

	ctx := context.Background()

	registryClient := erc8354.NewPolicyDomainRegistryClient(rpc, registry)
	acceptable, err := registryClient.IsRootAcceptable(ctx, domainID, root)
	if err != nil {
		t.Fatalf("IsRootAcceptable: %v", err)
	}
	if !acceptable {
		t.Errorf("IsRootAcceptable(current root) = false, want true")
	}

	domain, err := registryClient.Domain(ctx, domainID)
	if err != nil {
		t.Fatalf("Domain: %v", err)
	}
	if !domain.Active {
		t.Errorf("Domain.Active = false, want true")
	}
	// No identity registry was declared, so the field decodes as the zero
	// address and the guard's agent-existence check is a no-op.
	if domain.IdentityRegistry != (common.Address{}) {
		t.Errorf("Domain.IdentityRegistry = %s, want zero address", domain.IdentityRegistry)
	}

	executorKey, err := anvilPrivateKey()
	if err != nil {
		t.Fatalf("read anvil signer key: %v", err)
	}

	commitment, err := erc8354.ComputeActionCommitment(
		big.NewInt(31337),
		domainID,
		big.NewInt(1),
		common.HexToAddress("0x0000000000000000000000000000000000000001"),
		big.NewInt(0),
		[]byte{},
		big.NewInt(0),
	)
	if err != nil {
		t.Fatalf("ComputeActionCommitment: %v", err)
	}

	executor := crypto.PubkeyToAddress(executorKey.PublicKey)
	buildVerdict := func(nullifierSeed string) erc8354.Verdict {
		return erc8354.Verdict{
			AgentId:          big.NewInt(1),
			DomainId:         domainID,
			PolicyRoot:       root,
			ActionCommitment: commitment,
			Executor:         executor,
			Expiry:           4_000_000_000,
			Nullifier:        crypto.Keccak256Hash([]byte(nullifierSeed)),
			Decision:         1,
			PolicyKind:       0,
		}
	}

	guardClient := erc8354.NewConfidentialPolicyVerdictClient(rpc, guard, executorKey)

	ok, err := guardClient.Verify(ctx, buildVerdict("nf-verify"), []byte("proof"))
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}
	if !ok {
		t.Errorf("Verify(valid verdict) = false, want true")
	}

	// Direct consume: the transaction sender (anvil account 0) is the verdict
	// executor, so no executorAuth is required.
	directVerdict := buildVerdict("nf-consume")
	receipt, err := guardClient.Consume(ctx, directVerdict, []byte("proof"))
	if err != nil {
		t.Fatalf("Consume: %v", err)
	}
	if receipt.Status != types.ReceiptStatusSuccessful {
		t.Fatalf("Consume receipt status = %d, want 1", receipt.Status)
	}
	t.Logf("consume mined in block %d", receipt.BlockNumber)

	consumed, err := guardClient.IsConsumed(ctx, domainID, directVerdict.Nullifier)
	if err != nil {
		t.Fatalf("IsConsumed: %v", err)
	}
	if !consumed {
		t.Errorf("IsConsumed after consume = false, want true")
	}
}

func TestERC8354ConsumeRelayed(t *testing.T) {
	verifier, registry, guard := erc8354Addresses(t)

	rpc, err := ethclient.Dial(anvilRPC)
	if err != nil {
		t.Fatalf("ethclient.Dial(%s): %v", anvilRPC, err)
	}
	defer rpc.Close()

	domainID, root := domainRootPair("test-domain-relayed")
	setupDomain(t, rpc, registry, verifier, domainID, root)

	ctx := context.Background()

	commitment, err := erc8354.ComputeActionCommitment(
		big.NewInt(31337),
		domainID,
		big.NewInt(1),
		common.HexToAddress("0x0000000000000000000000000000000000000001"),
		big.NewInt(0),
		[]byte{},
		big.NewInt(0),
	)
	if err != nil {
		t.Fatalf("ComputeActionCommitment: %v", err)
	}

	// The executor (anvil account 1) authorizes a relayer (anvil account 0)
	// to submit the verdict on its behalf. The relayer signs the transaction,
	// the executor signs the EIP-712 verdict digest.
	executorKey, err := anvilPrivateKeyIndexed(1)
	if err != nil {
		t.Fatalf("read executor signer key: %v", err)
	}
	relayerKey, err := anvilPrivateKey()
	if err != nil {
		t.Fatalf("read relayer signer key: %v", err)
	}
	executor := crypto.PubkeyToAddress(executorKey.PublicKey)

	verdict := erc8354.Verdict{
		AgentId:          big.NewInt(1),
		DomainId:         domainID,
		PolicyRoot:       root,
		ActionCommitment: commitment,
		Executor:         executor,
		Expiry:           4_000_000_000,
		Nullifier:        crypto.Keccak256Hash([]byte("nf-relayed")),
		Decision:         1,
		PolicyKind:       0,
	}

	digest, err := erc8354.ComputeVerdictDigest(verdict, big.NewInt(31337), guard)
	if err != nil {
		t.Fatalf("ComputeVerdictDigest: %v", err)
	}
	executorAuth := signVerdictDigest(t, executorKey, digest)

	// The relayer sends the transaction, carrying the executor signature.
	relayClient := erc8354.NewConfidentialPolicyVerdictClient(rpc, guard, relayerKey)
	receipt, err := relayClient.ConsumeRelayed(ctx, verdict, []byte("proof"), executorAuth)
	if err != nil {
		t.Fatalf("ConsumeRelayed: %v", err)
	}
	if receipt.Status != types.ReceiptStatusSuccessful {
		t.Fatalf("ConsumeRelayed receipt status = %d, want 1", receipt.Status)
	}

	consumed, err := relayClient.IsConsumed(ctx, domainID, verdict.Nullifier)
	if err != nil {
		t.Fatalf("IsConsumed: %v", err)
	}
	if !consumed {
		t.Errorf("IsConsumed after ConsumeRelayed = false, want true")
	}
}

func TestERC8354RejectsBadExecutorSignature(t *testing.T) {
	verifier, registry, guard := erc8354Addresses(t)

	rpc, err := ethclient.Dial(anvilRPC)
	if err != nil {
		t.Fatalf("ethclient.Dial(%s): %v", anvilRPC, err)
	}
	defer rpc.Close()

	domainID, root := domainRootPair("test-domain-bad-sig")
	setupDomain(t, rpc, registry, verifier, domainID, root)

	ctx := context.Background()

	commitment, err := erc8354.ComputeActionCommitment(
		big.NewInt(31337),
		domainID,
		big.NewInt(1),
		common.HexToAddress("0x0000000000000000000000000000000000000001"),
		big.NewInt(0),
		[]byte{},
		big.NewInt(0),
	)
	if err != nil {
		t.Fatalf("ComputeActionCommitment: %v", err)
	}

	executorKey, err := anvilPrivateKeyIndexed(1)
	if err != nil {
		t.Fatalf("read executor signer key: %v", err)
	}
	relayerKey, err := anvilPrivateKey()
	if err != nil {
		t.Fatalf("read relayer signer key: %v", err)
	}
	executor := crypto.PubkeyToAddress(executorKey.PublicKey)

	verdict := erc8354.Verdict{
		AgentId:          big.NewInt(1),
		DomainId:         domainID,
		PolicyRoot:       root,
		ActionCommitment: commitment,
		Executor:         executor,
		Expiry:           4_000_000_000,
		Nullifier:        crypto.Keccak256Hash([]byte("nf-bad-sig")),
		Decision:         1,
		PolicyKind:       0,
	}

	// Garbage executorAuth: the relayer is not the executor and the signature
	// neither validates nor is empty, so the guard must revert.
	relayClient := erc8354.NewConfidentialPolicyVerdictClient(rpc, guard, relayerKey)
	if _, err := relayClient.ConsumeRelayed(ctx, verdict, []byte("proof"), []byte("not-a-signature")); err == nil {
		t.Error("ConsumeRelayed with bad executorAuth: got nil error, want revert")
	}
}

func TestERC8354RejectsInvalidProof(t *testing.T) {
	verifier, registry, guard := erc8354Addresses(t)

	rpc, err := ethclient.Dial(anvilRPC)
	if err != nil {
		t.Fatalf("ethclient.Dial(%s): %v", anvilRPC, err)
	}
	defer rpc.Close()

	domainID, root := domainRootPair("test-domain-invalid-proof")
	setupDomain(t, rpc, registry, verifier, domainID, root)

	// Force the domain verifier to reject the proof.
	setVerifierResult(t, rpc, verifier, false)

	ctx := context.Background()

	commitment, err := erc8354.ComputeActionCommitment(
		big.NewInt(31337),
		domainID,
		big.NewInt(1),
		common.HexToAddress("0x0000000000000000000000000000000000000001"),
		big.NewInt(0),
		[]byte{},
		big.NewInt(0),
	)
	if err != nil {
		t.Fatalf("ComputeActionCommitment: %v", err)
	}

	executorKey, err := anvilPrivateKey()
	if err != nil {
		t.Fatalf("read anvil signer key: %v", err)
	}
	executor := crypto.PubkeyToAddress(executorKey.PublicKey)

	verdict := erc8354.Verdict{
		AgentId:          big.NewInt(1),
		DomainId:         domainID,
		PolicyRoot:       root,
		ActionCommitment: commitment,
		Executor:         executor,
		Expiry:           4_000_000_000,
		Nullifier:        crypto.Keccak256Hash([]byte("nf-invalid-proof")),
		Decision:         1,
		PolicyKind:       0,
	}

	guardClient := erc8354.NewConfidentialPolicyVerdictClient(rpc, guard, executorKey)
	if _, err := guardClient.Consume(ctx, verdict, []byte("proof")); err == nil {
		t.Error("Consume with invalid proof: got nil error, want revert")
	}

	// Restore the shared verifier so later tests are not affected.
	setVerifierResult(t, rpc, verifier, true)
}