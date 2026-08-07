package erc8275

import (
	"strings"
	"testing"
)

func TestHashReproducesLockedIdentity(t *testing.T) {
	got, err := GoverningConventionHash(WinRateBpsV0Spec)
	if err != nil {
		t.Fatalf("GoverningConventionHash: %v", err)
	}
	if got != WinRateBpsV0Hash {
		t.Errorf("GoverningConventionHash(WinRateBpsV0Spec) = %s, want %s (convention-hash drift)", got, WinRateBpsV0Hash)
	}
}

func TestPinWinRateBps(t *testing.T) {
	pinned, err := PinWinRateBps(1, 31)
	if err != nil {
		t.Fatalf("PinWinRateBps(1, 31): %v", err)
	}
	if pinned.Value != 313 {
		t.Errorf("PinWinRateBps(1, 31).Value = %d, want 313", pinned.Value)
	}
	if pinned.GoverningConventionHash != WinRateBpsV0Hash {
		t.Errorf("PinWinRateBps(1, 31).GoverningConventionHash = %s, want %s", pinned.GoverningConventionHash, WinRateBpsV0Hash)
	}
	if _, err := PinWinRateBps(0, 0); err != ErrZeroTotal {
		t.Errorf("PinWinRateBps(0, 0) error = %v, want ErrZeroTotal", err)
	}
}

func TestVerifyWinRateVectors(t *testing.T) {
	// bps golden + ties -> verified
	for _, tc := range []struct {
		w, l, v uint64
	}{
		{0, 15, 0}, {16, 0, 10000}, {1, 2, 3333}, {16, 15, 5161},
		{0, 10, 0}, {1, 31, 313}, {9, 23, 2813},
	} {
		verdict, err := VerifyWinRate(tc.v, WinRateBpsV0Hash, tc.w, tc.l)
		if err != nil {
			t.Errorf("VerifyWinRate(%d, hash, %d, %d): %v", tc.v, tc.w, tc.l, err)
			continue
		}
		if verdict != Verified {
			t.Errorf("VerifyWinRate(%d, hash, %d, %d) = %s, want verified", tc.v, tc.w, tc.l, verdict)
		}
	}
	// half-even value under bps -> rejected
	if verdict, err := VerifyWinRate(312, WinRateBpsV0Hash, 1, 31); err != nil || verdict != Rejected {
		t.Errorf("VerifyWinRate(312, hash, 1, 31) = %s, %v; want rejected", verdict, err)
	}
	// unknown convention -> unverifiable, never a silent pass
	unknown := "0x" + strings.Repeat("de", 32)
	verdict, err := VerifyWinRate(9500, unknown, 19, 1)
	if err == nil {
		t.Error("VerifyWinRate with unknown hash returned nil error, want ErrInvalidConventionHash")
	}
	if verdict != Unverifiable {
		t.Errorf("VerifyWinRate with unknown hash = %s, want unverifiable", verdict)
	}
}

func TestVerdictString(t *testing.T) {
	cases := map[Verdict]string{
		Verified:     "verified",
		Rejected:     "rejected",
		Unverifiable: "unverifiable",
	}
	for v, want := range cases {
		if got := v.String(); got != want {
			t.Errorf("Verdict(%d).String() = %q, want %q", int(v), got, want)
		}
	}
}
