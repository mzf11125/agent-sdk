// Package test holds integration tests that run against a local anvil node
// deployed via the testkit harness.
//
// Usage:
//
//	testkit/scripts/start-anvil.sh
//	ERC8323_ADDRESS=$(testkit/scripts/deploy.sh identity/ERC8323 DeployERC8323)
//	ERC8323_ADDRESS="$ERC8323_ADDRESS" go test -v ./go/test/ -run TestERC8323
//	testkit/scripts/stop-anvil.sh
//
// deploy.sh prints one contract-creation address per line in broadcast
// order: the dummy source ERC-721 collection first, then the
// MockAgentSourceBinding. ERC8323_ADDRESS carries those two addresses
// separated by newlines (shell command substitution preserves internal
// newlines).
package test

import (
	"context"
	"math/big"
	"os"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/trustless-ai/agent-sdk/go/identity/erc8323"
)

// erc8323MintPrice matches MockAgentSourceBinding.MINT_PRICE (0.001 ether, in
// wei) — the testkit mock enforces msg.value == MINT_PRICE, mirroring paid
// production registries (see the 2026-07-16 value-threading bug this guards).
var erc8323MintPrice = new(big.Int).Exp(big.NewInt(10), big.NewInt(15), nil)

// erc8323Addresses reads the deployed dummy source collection and the
// MockAgentSourceBinding from ERC8323_ADDRESS (two lines, broadcast order).
func erc8323Addresses(t *testing.T) (dummy, binding common.Address) {
	t.Helper()
	lines := strings.Split(strings.TrimSpace(os.Getenv("ERC8323_ADDRESS")), "\n")
	if len(lines) < 2 {
		t.Fatal("ERC8323_ADDRESS must carry 2 addresses (dummy collection, binding) — deploy first via testkit/scripts/deploy.sh identity/ERC8323 DeployERC8323")
	}
	return common.HexToAddress(strings.TrimSpace(lines[0])), common.HexToAddress(strings.TrimSpace(lines[1]))
}

func TestERC8323RegisterAndRead(t *testing.T) {
	dummyAddr, bindingAddr := erc8323Addresses(t)

	rpc, err := ethclient.Dial(anvilRPC)
	if err != nil {
		t.Fatalf("ethclient.Dial(%s): %v", anvilRPC, err)
	}
	defer rpc.Close()

	key := anvilSignerKey(t)
	client := erc8323.NewSourceBindingClient(rpc, bindingAddr, key)

	// 1. boundCollection() must point at the deployed dummy collection.
	collection, err := client.BoundCollection(context.Background())
	if err != nil {
		t.Fatalf("BoundCollection: %v", err)
	}
	t.Logf("boundCollection() = %s", collection.Hex())
	if collection != dummyAddr {
		t.Errorf("boundCollection() = %s, want dummy collection %s", collection.Hex(), dummyAddr.Hex())
	}

	// 2. The registry advertises IAgentSourceBinding (ERC-165 0x27eba962).
	supported, err := client.SupportsSourceBinding(context.Background())
	if err != nil {
		t.Fatalf("SupportsSourceBinding: %v", err)
	}
	if !supported {
		t.Error("SupportsSourceBinding() = false, want true")
	}

	// 3. Register an agent derived from source token 42, paying the mint price.
	agentID, err := client.Register(context.Background(), big.NewInt(42), erc8323MintPrice)
	if err != nil {
		t.Fatalf("Register(42): %v", err)
	}
	t.Logf("registerWithSource(42) -> agentId = %s", agentID)
	if agentID.Sign() <= 0 {
		t.Fatalf("Register(42) agentId = %s, want > 0", agentID)
	}

	// 4. Read the immutable provenance back.
	src, err := client.GetSourceNFT(context.Background(), agentID)
	if err != nil {
		t.Fatalf("GetSourceNFT(%s): %v", agentID, err)
	}
	t.Logf("getSourceNFT(%s) = (%s, %s)", agentID, src.SourceContract.Hex(), src.SourceTokenID)
	if src.SourceContract != dummyAddr {
		t.Errorf("getSourceNFT(%s).sourceContract = %s, want %s", agentID, src.SourceContract.Hex(), dummyAddr.Hex())
	}
	if src.SourceTokenID.Cmp(big.NewInt(42)) != 0 {
		t.Errorf("getSourceNFT(%s).sourceTokenId = %s, want 42", agentID, src.SourceTokenID)
	}

	// 5. The agent claims a source binding and its ownership is valid.
	has, err := client.HasSourceNFT(context.Background(), agentID)
	if err != nil {
		t.Fatalf("HasSourceNFT(%s): %v", agentID, err)
	}
	if !has {
		t.Errorf("HasSourceNFT(registered agent) = false, want true")
	}

	valid, err := client.IsSourceNFTOwnershipValid(context.Background(), agentID)
	if err != nil {
		t.Fatalf("IsSourceNFTOwnershipValid(%s): %v", agentID, err)
	}
	if !valid {
		t.Errorf("IsSourceNFTOwnershipValid(registered agent) = false, want true")
	}

	// 6. The agent id is a minted ERC-721 owned by the registering signer.
	owner, err := client.OwnerOf(context.Background(), agentID)
	if err != nil {
		t.Fatalf("OwnerOf(%s): %v", agentID, err)
	}
	t.Logf("ownerOf(%s) = %s", agentID, owner.Hex())
	signer := crypto.PubkeyToAddress(key.PublicKey)
	if owner != signer {
		t.Errorf("OwnerOf(%s) = %s, want signer %s", agentID, owner.Hex(), signer.Hex())
	}
}

func TestERC8323RegisterRevertsWithoutMintPrice(t *testing.T) {
	_, bindingAddr := erc8323Addresses(t)

	rpc, err := ethclient.Dial(anvilRPC)
	if err != nil {
		t.Fatalf("ethclient.Dial(%s): %v", anvilRPC, err)
	}
	defer rpc.Close()

	key := anvilSignerKey(t)
	client := erc8323.NewSourceBindingClient(rpc, bindingAddr, key)

	// The mock enforces msg.value == MINT_PRICE; a value-less register must
	// revert loudly (guards the paid-registry case — a client that fails to
	// thread value through is silently broken against real registries).
	if _, err := client.Register(context.Background(), big.NewInt(7), nil); err == nil {
		t.Error("Register without mint price: got nil error, want revert")
	}
}

func TestERC8323UnregisteredAgentHasNoBinding(t *testing.T) {
	_, bindingAddr := erc8323Addresses(t)

	rpc, err := ethclient.Dial(anvilRPC)
	if err != nil {
		t.Fatalf("ethclient.Dial(%s): %v", anvilRPC, err)
	}
	defer rpc.Close()

	// Read-only client: no key.
	client := erc8323.NewSourceBindingClient(rpc, bindingAddr, nil)

	// A never-minted agent id has no source binding...
	has, err := client.HasSourceNFT(context.Background(), big.NewInt(9999))
	if err != nil {
		t.Fatalf("HasSourceNFT(9999): %v", err)
	}
	if has {
		t.Error("HasSourceNFT(unregistered agent) = true, want false")
	}

	// ...and reading its provenance reverts (per ERC-8323 getSourceNFT MUST
	// revert if agentId does not exist).
	if _, err := client.GetSourceNFT(context.Background(), big.NewInt(9999)); err == nil {
		t.Error("GetSourceNFT(unregistered agent): got nil error, want revert")
	}
}
