package test

import (
	"context"
	"os"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/trustless-ai/agent-sdk/go/anchor/erc8263"
)

func TestERC8263AnchorAndIsAnchored(t *testing.T) {
	addr := common.HexToAddress(os.Getenv("ERC8263_ADDRESS"))
	if addr == (common.Address{}) {
		t.Fatal("ERC8263_ADDRESS not set — deploy first via testkit/scripts/deploy.sh anchor/ERC8263 DeployERC8263")
	}

	rpc, err := ethclient.Dial(anvilRPC)
	if err != nil {
		t.Fatalf("ethclient.Dial(%s): %v", anvilRPC, err)
	}
	defer rpc.Close()

	key, err := anvilPrivateKey()
	if err != nil {
		t.Fatalf("read anvil signer key: %v", err)
	}

	client := erc8263.NewOnChainProofClient(rpc, addr, key)
	ctx := context.Background()

	// 1. ANONYMOUS scheme (0x00): agentId MUST be zero.
	proofAnonymous := crypto.Keccak256Hash([]byte("proof-anonymous"))
	if _, err := client.Anchor(ctx, 0x00, common.Hash{}, proofAnonymous); err != nil {
		t.Fatalf("Anchor(ANONYMOUS): %v", err)
	}
	if anchored, err := client.IsAnchored(proofAnonymous); err != nil {
		t.Fatalf("IsAnchored(ANONYMOUS): %v", err)
	} else if !anchored {
		t.Error("IsAnchored(ANONYMOUS) immediately after Anchor = false, want true")
	}

	// 2. REGISTRY scheme (0x01): non-zero agentId.
	proofRegistry := crypto.Keccak256Hash([]byte("proof-registry"))
	agentID := crypto.Keccak256Hash([]byte("agent-8004-42"))
	if _, err := client.Anchor(ctx, 0x01, agentID, proofRegistry); err != nil {
		t.Fatalf("Anchor(REGISTRY): %v", err)
	}
	if anchored, err := client.IsAnchored(proofRegistry); err != nil {
		t.Fatalf("IsAnchored(REGISTRY): %v", err)
	} else if !anchored {
		t.Error("IsAnchored(REGISTRY) immediately after Anchor = false, want true")
	}

	// 3. URI_HASH scheme (0x02) via anchorWithAux with opaque extension bytes.
	proofURI := crypto.Keccak256Hash([]byte("proof-uri"))
	uriHash := crypto.Keccak256Hash([]byte("did:agent:example"))
	if _, err := client.AnchorWithAux(ctx, 0x02, uriHash, proofURI, []byte("hello-aux")); err != nil {
		t.Fatalf("AnchorWithAux(URI_HASH): %v", err)
	}
	if anchored, err := client.IsAnchored(proofURI); err != nil {
		t.Fatalf("IsAnchored(URI_HASH): %v", err)
	} else if !anchored {
		t.Error("IsAnchored(URI_HASH) immediately after AnchorWithAux = false, want true")
	}

	// A never-anchored hash must not be found.

	unknown := crypto.Keccak256Hash([]byte("never-anchored"))
	anchored, err := client.IsAnchored(unknown)
	if err != nil {
		t.Fatalf("IsAnchored(unknown): %v", err)
	}
	if anchored {
		t.Errorf("IsAnchored(unanchored proof) = true, want false")
	}
}

func TestERC8263RejectsInvalidAnchors(t *testing.T) {
	addr := common.HexToAddress(os.Getenv("ERC8263_ADDRESS"))
	if addr == (common.Address{}) {
		t.Fatal("ERC8263_ADDRESS not set — deploy first via testkit/scripts/deploy.sh anchor/ERC8263 DeployERC8263")
	}

	rpc, err := ethclient.Dial(anvilRPC)
	if err != nil {
		t.Fatalf("ethclient.Dial(%s): %v", anvilRPC, err)
	}
	defer rpc.Close()

	key, err := anvilPrivateKey()
	if err != nil {
		t.Fatalf("read anvil signer key: %v", err)
	}

	client := erc8263.NewOnChainProofClient(rpc, addr, key)
	ctx := context.Background()

	// zero proofHash must revert.
	if _, err := client.Anchor(ctx, 0x01, crypto.Keccak256Hash([]byte("a")), common.Hash{}); err == nil {
		t.Error("Anchor with zero proofHash: got nil error, want revert")
	}

	// ANONYMOUS scheme requires agentId == 0.
	if _, err := client.Anchor(ctx, 0x00, crypto.Keccak256Hash([]byte("b")), crypto.Keccak256Hash([]byte("c"))); err == nil {
		t.Error("Anchor(ANONYMOUS) with non-zero agentId: got nil error, want revert")
	}

	// schemes 0x03+ are reserved.
	if _, err := client.Anchor(ctx, 0x03, common.Hash{}, crypto.Keccak256Hash([]byte("d"))); err == nil {
		t.Error("Anchor(reserved scheme): got nil error, want revert")
	}
}
