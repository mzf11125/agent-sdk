// Package test holds integration tests that run against a local anvil node
// deployed via the testkit harness.
//
// Usage:
//
//	testkit/scripts/start-anvil.sh
//	ERC8281_ADDRESS=$(testkit/scripts/deploy.sh verify/ERC8281 DeployERC8281)
//	ERC8281_ADDRESS=$ERC8281_ADDRESS go test -v ./go/test/ -run TestERC8281
//	testkit/scripts/stop-anvil.sh
package test

import (
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/trustless-ai/agent-sdk/go/verify/erc8281"
)

func TestERC8281RecordAndCheckRecorded(t *testing.T) {
	addr := common.HexToAddress(os.Getenv("ERC8281_ADDRESS"))
	if addr == (common.Address{}) {
		t.Fatal("ERC8281_ADDRESS not set — deploy first via testkit/scripts/deploy.sh verify/ERC8281 DeployERC8281")
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

	client := erc8281.NewObservationCommitmentClient(rpc, addr, key)

	digest := crypto.Keccak256Hash([]byte("hello"))
	unknown := crypto.Keccak256Hash([]byte("never-recorded"))

	if _, err := client.Record(digest); err != nil {
		t.Fatalf("Record(%s): %v", digest.Hex(), err)
	}

	recorded, err := client.CheckRecorded(digest)
	if err != nil {
		t.Fatalf("CheckRecorded(%s): %v", digest.Hex(), err)
	}
	if !recorded {
		t.Errorf("CheckRecorded(recorded digest) = false, want true")
	}

	notRecorded, err := client.CheckRecorded(unknown)
	if err != nil {
		t.Fatalf("CheckRecorded(%s): %v", unknown.Hex(), err)
	}
	if notRecorded {
		t.Errorf("CheckRecorded(unrecorded digest) = true, want false")
	}
}

// anvilPrivateKey returns the anvil account[0] private key used by
// testkit/scripts/deploy.sh — from the ANVIL_KEY env var if set, otherwise
// from testkit/.anvil-accounts.json (anvil regenerates keys on every start,
// so the key must be read at test time, never hardcoded).
func anvilPrivateKey() (*ecdsa.PrivateKey, error) {
	privHex := os.Getenv("ANVIL_KEY")
	if privHex == "" {
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
		if len(parsed.Accounts) == 0 || parsed.Accounts[0].PrivateKey == "" {
			return nil, fmt.Errorf("%s has no accounts[0].privateKey", path)
		}
		privHex = parsed.Accounts[0].PrivateKey
	}
	// anvil writes the key with a 0x prefix; HexToECDSA wants raw hex.
	privHex = strings.TrimPrefix(privHex, "0x")
	key, err := crypto.HexToECDSA(privHex)
	if err != nil {
		return nil, fmt.Errorf("decode private key: %w", err)
	}
	return key, nil
}
