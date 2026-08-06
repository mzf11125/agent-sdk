// Package test holds integration tests that run against a local anvil node
// deployed via the testkit harness.
//
// Usage:
//
//	testkit/scripts/start-anvil.sh
//	ERC8275_ADDRESS=$(testkit/scripts/deploy.sh reputation/ERC8275 DeployERC8275)
//	ERC8275_ADDRESS=$ERC8275_ADDRESS go test -v ./go/test/ -run TestERC8275
//	testkit/scripts/stop-anvil.sh
package test

import (
	"context"
	"os"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/trustless-ai/agent-sdk/go/reputation/erc8275"
)

const anvilRPC = "http://127.0.0.1:8545"

func TestERC8275GetReputation(t *testing.T) {
	addr := common.HexToAddress(os.Getenv("ERC8275_ADDRESS"))
	if addr == (common.Address{}) {
		t.Fatal("ERC8275_ADDRESS not set — deploy first via testkit/scripts/deploy.sh reputation/ERC8275 DeployERC8275")
	}

	rpc, err := ethclient.Dial(anvilRPC)
	if err != nil {
		t.Fatalf("ethclient.Dial(%s): %v", anvilRPC, err)
	}
	defer rpc.Close()

	client := erc8275.NewAgentReputationClient(rpc, addr)
	agentID := common.Hash{}

	rep, err := client.GetReputation(context.Background(), agentID)
	if err != nil {
		t.Fatalf("GetReputation(zero): %v", err)
	}
	t.Logf("getReputation(zero) = completed=%d, disputed=%d, volume=%d, lastActive=%d, score=%d",
		rep.CompletedOrders, rep.DisputedOrders, rep.TotalVolume, rep.LastActiveAt, rep.Score)
	if rep.CompletedOrders != 0 || rep.DisputedOrders != 0 || rep.TotalVolume != 0 ||
		rep.LastActiveAt != 0 || rep.Score != 0 {
		t.Errorf("getReputation(zero) = %+v, want all-zero default", rep)
	}

	weight, err := client.GetDecayWeight(context.Background(), agentID)
	if err != nil {
		t.Fatalf("GetDecayWeight(zero): %v", err)
	}
	t.Logf("getDecayWeight(zero) = %d", weight)
	if weight != 0 {
		t.Errorf("getDecayWeight(zero) = %d, want 0", weight)
	}

	valid, err := client.VerifyOutcome(context.Background(), agentID, []byte{})
	if err != nil {
		t.Fatalf("VerifyOutcome(zero, empty): %v", err)
	}
	t.Logf("verifyOutcome(zero, empty) = %v", valid)
	if valid {
		t.Error("verifyOutcome(zero, empty) = true, want false for unset outcome")
	}
}
