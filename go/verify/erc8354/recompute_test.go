package erc8354

import (
	"encoding/json"
	"math/big"
	"os"
	"path/filepath"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

const goldenDomainId = "0x34a63641b78652cdd53505da4f32cac6058bd148e3ff543f39f75997a89c2815"

func TestGoldenActionCommitment(t *testing.T) {
	result, err := ComputeActionCommitment(
		big.NewInt(31337),
		common.HexToHash(goldenDomainId),
		big.NewInt(1),
		common.HexToAddress("0x0000000000000000000000000000000000000001"),
		big.NewInt(0),
		[]byte{},
		big.NewInt(0),
	)
	if err != nil {
		t.Fatalf("ComputeActionCommitment: %v", err)
	}
	expected := common.HexToHash("0xcc8e5dc414db5ed2340be02c3d7fdc725fe5f1463b382a7ed13f8036a4a0b7b1")
	if result != expected {
		t.Fatalf("action commitment = %s, want %s", result.Hex(), expected.Hex())
	}
}

func TestActionCommitmentEmptyCallDataIsKeccakEmpty(t *testing.T) {
	result, err := ComputeActionCommitment(
		big.NewInt(31337),
		common.HexToHash(goldenDomainId),
		big.NewInt(1),
		common.HexToAddress("0x0000000000000000000000000000000000000001"),
		big.NewInt(0),
		[]byte{},
		big.NewInt(0),
	)
	if err != nil {
		t.Fatalf("ComputeActionCommitment: %v", err)
	}
	zero := common.Hash{}
	if result == zero {
		t.Fatalf("action commitment with empty callData is bytes32 zero, want keccak hash")
	}
}

func TestActionCommitmentChangesWithNonce(t *testing.T) {
	args := func(nonce int64) (common.Hash, error) {
		return ComputeActionCommitment(
			big.NewInt(31337),
			common.HexToHash(goldenDomainId),
			big.NewInt(1),
			common.HexToAddress("0x0000000000000000000000000000000000000001"),
			big.NewInt(0),
			[]byte{},
			big.NewInt(nonce),
		)
	}
	a, err := args(0)
	if err != nil {
		t.Fatalf("ComputeActionCommitment: %v", err)
	}
	b, err := args(1)
	if err != nil {
		t.Fatalf("ComputeActionCommitment: %v", err)
	}
	if a == b {
		t.Fatalf("commitment does not change with nonce")
	}
}

func TestGoldenVerdictDigest(t *testing.T) {
	v := Verdict{
		AgentId:          big.NewInt(1),
		DomainId:         common.HexToHash(goldenDomainId),
		PolicyRoot:       common.HexToHash(goldenDomainId),
		ActionCommitment: common.HexToHash("0xcc8e5dc414db5ed2340be02c3d7fdc725fe5f1463b382a7ed13f8036a4a0b7b1"),
		Executor:         common.HexToAddress("0x0000000000000000000000000000000000000002"),
		Expiry:           2000000000,
		Nullifier:        common.HexToHash("0x6e47261c83f90eed41cda2b00caad094c33daa0a09fec22396b3e2bfe5e222b2"),
		Decision:         1,
		PolicyKind:       0,
	}
	result, err := ComputeVerdictDigest(v, big.NewInt(31337), common.HexToAddress("0x0000000000000000000000000000000000000003"))
	if err != nil {
		t.Fatalf("ComputeVerdictDigest: %v", err)
	}
	expected := common.HexToHash("0xf2345f63ba9e78a068eb4f74640e6543289010540b457d8016771175ad460f32")
	if result != expected {
		t.Fatalf("verdict digest = %s, want %s", result.Hex(), expected.Hex())
	}
}

func TestMechanismConstant(t *testing.T) {
	expected := common.HexToHash("0xa843829a78c66c29679817606d0c8a9fa26575b6c2ed0f9f97079d7c46577ac6")
	if MechanismZkSecretPolicy != expected {
		t.Fatalf("mechanism = %s, want %s", MechanismZkSecretPolicy.Hex(), expected.Hex())
	}
}

type vectorFile struct {
	Vectors []vector `json:"vectors"`
}

type vector struct {
	Step     string          `json:"step"`
	Inputs   json.RawMessage `json:"inputs"`
	Expected string          `json:"expected"`
}

func loadVectors(t *testing.T) []vector {
	t.Helper()
	path := filepath.Join("..", "..", "..", "testkit", "vectors", "erc8354-verdict.vectors.json")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Skipf("golden vectors not found, skipping: %v", err)
		return nil
	}
	var file vectorFile
	if err := json.Unmarshal(raw, &file); err != nil {
		t.Fatalf("parse golden vectors: %v", err)
	}
	return file.Vectors
}

func TestVectorsFile(t *testing.T) {
	for _, v := range loadVectors(t) {
		t.Run(v.Step, func(t *testing.T) {
			switch v.Step {
			case "8354/action-commitment":
				var in struct {
					ChainId     json.Number `json:"chainId"`
					DomainId    string      `json:"domainId"`
					AgentId     json.Number `json:"agentId"`
					Target      string      `json:"target"`
					Value       json.Number `json:"value"`
					CallData    string      `json:"callData"`
					ActionNonce json.Number `json:"actionNonce"`
				}
				if err := json.Unmarshal(v.Inputs, &in); err != nil {
					t.Fatalf("unmarshal inputs: %v", err)
				}
				callData, err := hexutil.Decode(in.CallData)
				if err != nil {
					t.Fatalf("decode callData: %v", err)
				}
				result, err := ComputeActionCommitment(
					bigFromNumber(in.ChainId),
					common.HexToHash(in.DomainId),
					bigFromNumber(in.AgentId),
					common.HexToAddress(in.Target),
					bigFromNumber(in.Value),
					callData,
					bigFromNumber(in.ActionNonce),
				)
				if err != nil {
					t.Fatalf("ComputeActionCommitment: %v", err)
				}
				assertHash(t, result, v.Expected)
			case "8354/verdict-digest":
				var in struct {
					AgentId           json.Number `json:"agentId"`
					DomainId          string      `json:"domainId"`
					PolicyRoot        string      `json:"policyRoot"`
					ActionCommitment  string      `json:"actionCommitment"`
					Executor          string      `json:"executor"`
					Expiry            uint64      `json:"expiry"`
					Nullifier         string      `json:"nullifier"`
					Decision          uint8       `json:"decision"`
					PolicyKind        uint8       `json:"policyKind"`
					ChainId           json.Number `json:"chainId"`
					VerifyingContract string      `json:"verifyingContract"`
				}
				if err := json.Unmarshal(v.Inputs, &in); err != nil {
					t.Fatalf("unmarshal inputs: %v", err)
				}
				verdict := Verdict{
					AgentId:          bigFromNumber(in.AgentId),
					DomainId:         common.HexToHash(in.DomainId),
					PolicyRoot:       common.HexToHash(in.PolicyRoot),
					ActionCommitment: common.HexToHash(in.ActionCommitment),
					Executor:         common.HexToAddress(in.Executor),
					Expiry:           in.Expiry,
					Nullifier:        common.HexToHash(in.Nullifier),
					Decision:         in.Decision,
					PolicyKind:       in.PolicyKind,
				}
				result, err := ComputeVerdictDigest(
					verdict,
					bigFromNumber(in.ChainId),
					common.HexToAddress(in.VerifyingContract),
				)
				if err != nil {
					t.Fatalf("ComputeVerdictDigest: %v", err)
				}
				assertHash(t, result, v.Expected)
			default:
				t.Fatalf("unknown step %q", v.Step)
			}
		})
	}
}

func bigFromNumber(n json.Number) *big.Int {
	i, err := n.Int64()
	if err != nil {
		val, ok := new(big.Int).SetString(n.String(), 10)
		if !ok {
			return big.NewInt(0)
		}
		return val
	}
	return big.NewInt(i)
}

func assertHash(t *testing.T, got common.Hash, wantHex string) {
	t.Helper()
	want := common.HexToHash(wantHex)
	if got != want {
		t.Fatalf("got %s, want %s", got.Hex(), want.Hex())
	}
}
