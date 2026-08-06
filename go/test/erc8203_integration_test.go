// Package test holds integration tests that run against a local anvil node
// deployed via the testkit harness.
//
// Usage:
//
//	testkit/scripts/start-anvil.sh
//	ERC8203_ADDRESS=$(testkit/scripts/deploy.sh settlement/ERC8203 DeployERC8203)
//	ERC8203_ADDRESS=$ERC8203_ADDRESS go test -v ./go/test/ -run TestERC8203
//	testkit/scripts/stop-anvil.sh
package test

import (
	"context"
	"math/big"
	"os"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/trustless-ai/agent-sdk/go/settlement/erc8203"
)

const escrowValue = 1000000000000000000 // 1 ETH in wei

// TestERC8203OpenReleaseSettle exercises the full ConsultEscrow flow on a
// deployed contract: open a job as the testkit consumer, read it back,
// recompute the release commitment off-chain, sign it with a fresh
// attestor key, resolve (release) the escrow, and verify the chain's
// emitted commitmentHash matches the off-chain recompute.
func TestERC8203OpenReleaseSettle(t *testing.T) {
	addr := common.HexToAddress(os.Getenv("ERC8203_ADDRESS"))
	if addr == (common.Address{}) {
		t.Fatal("ERC8203_ADDRESS not set — deploy first via testkit/scripts/deploy.sh settlement/ERC8203 DeployERC8203")
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
	consumer := crypto.PubkeyToAddress(key.PublicKey)

	ctx := context.Background()
	client := erc8203.NewConsultEscrowClient(rpc, addr, key)

	// Fresh, deterministic job — anvil state is fresh per testkit session.
	jobID := crypto.Keccak256Hash([]byte("erc8203:go:integration:job"))

	// A never-opened job reads as the all-zero default.
	unset, err := client.GetJob(ctx, jobID)
	if err != nil {
		t.Fatalf("GetJob(unopened): %v", err)
	}
	t.Logf("jobs(unopened) = consumer=%s status=%d", unset.Consumer.Hex(), unset.Status)
	if unset.Consumer != (common.Address{}) || unset.Provider != (common.Address{}) ||
		unset.Attestor != (common.Address{}) || unset.Status != erc8203.JobStatusNone {
		t.Errorf("GetJob(unopened) = %+v, want all-zero default with Status None", unset)
	}

	// Fresh attestor + provider keys: the attestor's key signs the release
	// commitment; the provider receives the escrow.
	attestorKey, err := crypto.GenerateKey()
	if err != nil {
		t.Fatalf("generate attestor key: %v", err)
	}
	attestor := crypto.PubkeyToAddress(attestorKey.PublicKey)
	providerKey, err := crypto.GenerateKey()
	if err != nil {
		t.Fatalf("generate provider key: %v", err)
	}
	provider := crypto.PubkeyToAddress(providerKey.PublicKey)

	// Open the escrow as the consumer, locking 1 ETH. The client has no
	// Open method (Go SDK scope: GetJob + Resolve), so the test sets the
	// job up through the exported ABI directly.
	a, err := erc8203.ConsultEscrowABI()
	if err != nil {
		t.Fatalf("parse ABI: %v", err)
	}
	chainID, err := rpc.ChainID(ctx)
	if err != nil {
		t.Fatalf("chain id: %v", err)
	}
	header, err := rpc.HeaderByNumber(ctx, nil)
	if err != nil {
		t.Fatalf("latest header: %v", err)
	}
	deadline := new(big.Int).Add(new(big.Int).SetUint64(header.Time), big.NewInt(3600))

	auth, err := bind.NewKeyedTransactorWithChainID(key, chainID)
	if err != nil {
		t.Fatalf("create transactor: %v", err)
	}
	auth.Value = big.NewInt(escrowValue)
	bound := bind.NewBoundContract(addr, a, rpc, rpc, rpc)
	tx, err := bound.Transact(auth, "open", jobID, provider, attestor, deadline)
	if err != nil {
		t.Fatalf("open(%s): %v", jobID.Hex(), err)
	}
	receipt, err := bind.WaitMined(ctx, rpc, tx)
	if err != nil {
		t.Fatalf("WaitMined(open %s): %v", tx.Hash().Hex(), err)
	}
	t.Logf("open tx %s mined in block %d", tx.Hash().Hex(), receipt.BlockNumber)

	// The job is now Open with the testkit consumer as msg.sender.
	opened, err := client.GetJob(ctx, jobID)
	if err != nil {
		t.Fatalf("GetJob(opened): %v", err)
	}
	t.Logf("jobs(opened) = consumer=%s provider=%s attestor=%s amount=%d deadline=%d status=%d",
		opened.Consumer.Hex(), opened.Provider.Hex(), opened.Attestor.Hex(), opened.Amount, opened.Deadline, opened.Status)
	if opened.Status != erc8203.JobStatusOpen {
		t.Errorf("GetJob(opened).Status = %d, want %d (Open)", opened.Status, erc8203.JobStatusOpen)
	}
	if opened.Consumer != consumer {
		t.Errorf("GetJob(opened).Consumer = %s, want %s", opened.Consumer.Hex(), consumer.Hex())
	}
	if opened.Attestor != attestor {
		t.Errorf("GetJob(opened).Attestor = %s, want %s", opened.Attestor.Hex(), attestor.Hex())
	}
	if opened.Amount.Cmp(big.NewInt(escrowValue)) != 0 {
		t.Errorf("GetJob(opened).Amount = %d, want %d", opened.Amount, escrowValue)
	}

	// Recompute the release commitment off-chain from public data.
	resultText := "No intermediaries required, cryptographic verification only."
	resultHash := crypto.Keccak256Hash([]byte(resultText))
	verdict, err := erc8203.ComputeVerdictHash(jobID, resultText)
	if err != nil {
		t.Fatalf("ComputeVerdictHash: %v", err)
	}
	t.Logf("recomputed commitment = %s", verdict.Hex())

	// Attestor signs the commitment (EIP-191 personal_sign); v must be
	// 27/28 for ecrecover.
	ethHash := crypto.Keccak256Hash(append([]byte("\x19Ethereum Signed Message:\n32"), verdict[:]...))
	sig, err := crypto.Sign(ethHash[:], attestorKey)
	if err != nil {
		t.Fatalf("attestor sign: %v", err)
	}
	sig[64] += 27

	// Resolve (release) the escrow.
	tx, err = client.Resolve(ctx, jobID, resultHash, sig)
	if err != nil {
		t.Fatalf("Resolve(%s): %v", jobID.Hex(), err)
	}
	receipt, err = bind.WaitMined(ctx, rpc, tx)
	if err != nil {
		t.Fatalf("WaitMined(release %s): %v", tx.Hash().Hex(), err)
	}
	t.Logf("release tx %s mined in block %d", tx.Hash().Hex(), receipt.BlockNumber)

	// The escrow is Released, and the on-chain commitmentHash matches the
	// off-chain recompute — the recompute-to-verify assertion.
	settled, err := client.GetJob(ctx, jobID)
	if err != nil {
		t.Fatalf("GetJob(settled): %v", err)
	}
	t.Logf("jobs(settled) = status=%d", settled.Status)
	if settled.Status != erc8203.JobStatusReleased {
		t.Errorf("GetJob(settled).Status = %d, want %d (Released)", settled.Status, erc8203.JobStatusReleased)
	}

	released, ok := a.Events["Released"]
	if !ok {
		t.Fatal("ABI has no Released event")
	}
	for _, log := range receipt.Logs {
		if len(log.Topics) == 0 || log.Topics[0] != released.ID {
			continue
		}
		vals, err := released.Inputs.Unpack(log.Data)
		if err != nil {
			t.Fatalf("unpack Released data: %v", err)
		}
		// go-ethereum unpacks bytes32 event inputs as [32]byte.
		commitmentBytes, ok := vals[1].([32]byte)
		if !ok {
			t.Fatalf("Released commitmentHash is %T, want [32]byte", vals[1])
		}
		commitment := common.BytesToHash(commitmentBytes[:])
		t.Logf("on-chain commitmentHash = %s", commitment.Hex())
		if commitment != verdict {
			t.Errorf("on-chain commitmentHash %s != off-chain recompute %s", commitment.Hex(), verdict.Hex())
		}
		emittedResultHashBytes, ok := vals[0].([32]byte)
		if !ok {
			t.Fatalf("Released resultHash is %T, want [32]byte", vals[0])
		}
		emittedResultHash := common.BytesToHash(emittedResultHashBytes[:])
		if emittedResultHash != resultHash {
			t.Errorf("on-chain resultHash %s != %s", emittedResultHash.Hex(), resultHash.Hex())
		}
		return
	}
	t.Fatal("no Released event found in release receipt")
}
