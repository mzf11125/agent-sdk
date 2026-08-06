package erc8281

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
)

// Golden vector: keccak256("hello")
func TestGoldenDigest(t *testing.T) {
	got := ComputeObservationDigest([]byte("hello"))
	want := common.HexToHash("0x1c8aff950685c2ed4bc3174f3472287b56d9517b9c948127319a09a7a36deac8")
	if got != want {
		t.Errorf("ComputeObservationDigest(\"hello\") = %s, want %s", got.Hex(), want.Hex())
	}
}

// keccak256("") is a fixed, non-zero hash — an empty observation still
// commits to something.
func TestEmptyDigest(t *testing.T) {
	got := ComputeObservationDigest(nil)
	if got == (common.Hash{}) {
		t.Error("ComputeObservationDigest(nil) = zero hash, want non-zero")
	}
	want := common.HexToHash("0xc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470")
	if got != want {
		t.Errorf("ComputeObservationDigest(nil) = %s, want %s (keccak256 of empty)", got.Hex(), want.Hex())
	}
}

func TestDifferentInputsDifferentHashes(t *testing.T) {
	a := ComputeObservationDigest([]byte("a"))
	b := ComputeObservationDigest([]byte("b"))
	if a == b {
		t.Error("ComputeObservationDigest(\"a\") == ComputeObservationDigest(\"b\"), want different hashes")
	}
}
