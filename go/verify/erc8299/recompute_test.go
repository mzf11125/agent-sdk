package erc8299

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
)

// Golden vector: keccak256("hello") — recompute-kit "wyriwe/raw", also
// cross-verified against the TS, Python and Rust SDKs.
func TestGoldenRawInputHash(t *testing.T) {
	got := ComputeRawInputHash([]byte("hello"))
	want := common.HexToHash("0x1c8aff950685c2ed4bc3174f3472287b56d9517b9c948127319a09a7a36deac8")
	if got != want {
		t.Errorf("ComputeRawInputHash(\"hello\") = %s, want %s", got.Hex(), want.Hex())
	}
}

// Golden vector: keccak256(utf8(cid) || raw_input_hash) — recompute-kit
// "wyriwe/pipeline", cross-verified against the TS, Python and Rust SDKs.
func TestGoldenSanitizationPipelineHash(t *testing.T) {
	rawHash := common.HexToHash("0x1c8aff950685c2ed4bc3174f3472287b56d9517b9c948127319a09a7a36deac8")
	got := ComputeSanitizationPipelineHash("ipfs://QmccvoM6aRVgZ2dtFWvT6Wm3DmTvoAUHHotK7uQufnStVR", rawHash)
	want := common.HexToHash("0x5798efed4aa92f96a0622fc30268042b067294bdb5fd06f599bf8d84fd5d734b")
	if got != want {
		t.Errorf("ComputeSanitizationPipelineHash(cid, %s) = %s, want %s",
			rawHash.Hex(), got.Hex(), want.Hex())
	}
}

// keccak256("") is a fixed, non-zero hash — an empty raw input still hashes
// to something deterministic.
func TestEmptyRawInputHash(t *testing.T) {
	got := ComputeRawInputHash(nil)
	if got == (common.Hash{}) {
		t.Error("ComputeRawInputHash(nil) = zero hash, want non-zero")
	}
	want := common.HexToHash("0xc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470")
	if got != want {
		t.Errorf("ComputeRawInputHash(nil) = %s, want %s (keccak256 of empty)", got.Hex(), want.Hex())
	}
}

// An empty CID still yields a valid, non-zero pipeline hash.
func TestEmptyCIDPipelineHash(t *testing.T) {
	rawHash := common.HexToHash("0x1c8aff950685c2ed4bc3174f3472287b56d9517b9c948127319a09a7a36deac8")
	got := ComputeSanitizationPipelineHash("", rawHash)
	if got == (common.Hash{}) {
		t.Error("ComputeSanitizationPipelineHash(\"\", raw) = zero hash, want non-zero")
	}
}

func TestDifferentInputsDifferentHashes(t *testing.T) {
	a := ComputeRawInputHash([]byte("a"))
	b := ComputeRawInputHash([]byte("b"))
	if a == b {
		t.Error("ComputeRawInputHash(\"a\") == ComputeRawInputHash(\"b\"), want different hashes")
	}
}

func TestDifferentCIDsDifferentPipelineHashes(t *testing.T) {
	rawHash := common.HexToHash("0x1c8aff950685c2ed4bc3174f3472287b56d9517b9c948127319a09a7a36deac8")
	a := ComputeSanitizationPipelineHash("cid:a", rawHash)
	b := ComputeSanitizationPipelineHash("cid:b", rawHash)
	if a == b {
		t.Error("ComputeSanitizationPipelineHash(\"cid:a\") == ComputeSanitizationPipelineHash(\"cid:b\"), want different hashes")
	}
}
