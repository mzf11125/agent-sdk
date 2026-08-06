package erc8004

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
)

// Golden vector: registryId 860 → 0x0000...035c
func TestGoldenRegistryID860(t *testing.T) {
	want := common.HexToHash("0x000000000000000000000000000000000000000000000000000000000000035c")
	got := ComputeAgentId(860)
	if got != want {
		t.Errorf("ComputeAgentId(860) = %s, want %s", got.Hex(), want.Hex())
	}
}

func TestZeroRegistryID(t *testing.T) {
	got := ComputeAgentId(0)
	if got != (common.Hash{}) {
		t.Errorf("ComputeAgentId(0) = %s, want all-zero hash", got.Hex())
	}
}

// 54848 = 0xD640 — still fits in the trailing two bytes, so the leading
// 24 bytes must stay zero (correct left padding).
func TestRegistryID54848(t *testing.T) {
	want := common.HexToHash("0x000000000000000000000000000000000000000000000000000000000000d640")
	got := ComputeAgentId(54848)
	if got != want {
		t.Errorf("ComputeAgentId(54848) = %s, want %s", got.Hex(), want.Hex())
	}
}

// Max u64: 0xffff_ffff_ffff_ffff — still fits in bytes32, left-padded with
// 24 zero bytes; the u64 occupies bytes 24..32.
func TestMaxUint64LeftPadded(t *testing.T) {
	got := ComputeAgentId(^uint64(0))
	for i := 0; i < 24; i++ {
		if got[i] != 0 {
			t.Fatalf("ComputeAgentId(^uint64(0)) byte %d = 0x%02x, want 0x00 (left-padded)", i, got[i])
		}
	}
	if got[24] != 0xff {
		t.Errorf("ComputeAgentId(^uint64(0)) byte 24 = 0x%02x, want 0xff", got[24])
	}
}
