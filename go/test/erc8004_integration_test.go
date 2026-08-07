package test

import (
	"context"
	"math/big"
	"os"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/trustless-ai/agent-sdk/go/identity/erc8004"
)

// TestERC8004RegisterAndRead registers an agent on a live ERC-8004
// contract (via a raw signed transaction — the client is read-only), then
// reads the registration URI and owner back through the client.
//
// Prerequisites (see package comment in erc8275_integration_test.go):
//
//	testkit/scripts/start-anvil.sh
//	ERC8004_ADDRESS=$(testkit/scripts/deploy.sh identity/ERC8004 DeployERC8004)
//	ERC8004_ADDRESS=$ERC8004_ADDRESS go test -v ./go/test/ -run TestERC8004
//	testkit/scripts/stop-anvil.sh
func TestERC8004RegisterAndRead(t *testing.T) {
	addr := common.HexToAddress(os.Getenv("ERC8004_ADDRESS"))
	if addr == (common.Address{}) {
		t.Fatal("ERC8004_ADDRESS not set — deploy first via testkit/scripts/deploy.sh identity/ERC8004 DeployERC8004")
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
	chainID, err := rpc.ChainID(context.Background())
	if err != nil {
		t.Fatalf("ChainID: %v", err)
	}
	opts, err := bind.NewKeyedTransactorWithChainID(key, chainID)
	if err != nil {
		t.Fatalf("NewKeyedTransactorWithChainID: %v", err)
	}
	opts.GasLimit = 300000

	a, err := erc8004.IdentityRegistryABI()
	if err != nil {
		t.Fatalf("IdentityRegistryABI: %v", err)
	}
	bound := bind.NewBoundContract(addr, a, rpc, rpc, rpc)

	// Simulate the register call to learn the agentId the contract is about
	// to assign (fresh mocks start at 1, but this stays correct on reruns).
	uri := "ipfs://test-agent"
	var sim []interface{}
	if err := bound.Call(
		&bind.CallOpts{Context: context.Background(), From: opts.From},
		&sim, "register", uri, []erc8004.MetadataEntry{},
	); err != nil {
		t.Fatalf("simulate register: %v", err)
	}
	agentID, ok := sim[0].(*big.Int)
	if !ok {
		t.Fatalf("simulate register returned %T, want *big.Int", sim[0])
	}

	tx, err := bound.Transact(opts, "register", uri, []erc8004.MetadataEntry{})
	if err != nil {
		t.Fatalf("register(%q): %v", uri, err)
	}
	if _, err := bind.WaitMined(context.Background(), rpc, tx); err != nil {
		t.Fatalf("WaitMined(%s): %v", tx.Hash().Hex(), err)
	}
	t.Logf("register(%q) mined, agentId = %s", uri, agentID)

	client := erc8004.NewIdentityRegistryClient(rpc, addr)

	gotURI, err := client.GetAgentURI(context.Background(), agentID)
	if err != nil {
		t.Fatalf("GetAgentURI(%s): %v", agentID, err)
	}
	t.Logf("getAgentURI(%s) = %q", agentID, gotURI)
	if gotURI != uri {
		t.Errorf("getAgentURI(%s) = %q, want %q", agentID, gotURI, uri)
	}

	owner, err := client.OwnerOf(context.Background(), agentID)
	if err != nil {
		t.Fatalf("OwnerOf(%s): %v", agentID, err)
	}
	t.Logf("ownerOf(%s) = %s", agentID, owner.Hex())
	if owner != opts.From {
		t.Errorf("ownerOf(%s) = %s, want registering sender %s", agentID, owner.Hex(), opts.From.Hex())
	}
}
