package erc8275

import "testing"

// Golden vector: wins=16, losses=15 → 5161 (0.5161)
func TestGoldenWinRate(t *testing.T) {
	got, err := ComputeWinRate(16, 15)
	if err != nil {
		t.Fatalf("ComputeWinRate(16, 15) returned error: %v", err)
	}
	if got != 5161 {
		t.Errorf("ComputeWinRate(16, 15) = %d, want 5161", got)
	}
}

func TestZeroWins(t *testing.T) {
	got, err := ComputeWinRate(0, 10)
	if err != nil {
		t.Fatalf("ComputeWinRate(0, 10) returned error: %v", err)
	}
	if got != 0 {
		t.Errorf("ComputeWinRate(0, 10) = %d, want 0", got)
	}
}

func TestAllWins(t *testing.T) {
	got, err := ComputeWinRate(10, 0)
	if err != nil {
		t.Fatalf("ComputeWinRate(10, 0) returned error: %v", err)
	}
	if got != 10000 {
		t.Errorf("ComputeWinRate(10, 0) = %d, want 10000", got)
	}
}

func TestBothZeroReturnsError(t *testing.T) {
	if _, err := ComputeWinRate(0, 0); err == nil {
		t.Error("ComputeWinRate(0, 0) returned nil error, want ErrZeroTotal")
	} else if err != ErrZeroTotal {
		t.Errorf("ComputeWinRate(0, 0) error = %v, want ErrZeroTotal", err)
	}
}

// 1/3 = 0.3333... → 3333 basis points
func TestIntegerDivisionTruncates(t *testing.T) {
	got, err := ComputeWinRate(1, 2)
	if err != nil {
		t.Fatalf("ComputeWinRate(1, 2) returned error: %v", err)
	}
	if got != 3333 {
		t.Errorf("ComputeWinRate(1, 2) = %d, want 3333", got)
	}
}

// Rounding-tie vector: wins=1, losses=31 → 1/32 = 0.03125 → 313 (ROUND_HALF_UP)
func TestRoundingTieHalfUp(t *testing.T) {
	got, err := ComputeWinRate(1, 31)
	if err != nil {
		t.Fatalf("ComputeWinRate(1, 31) returned error: %v", err)
	}
	if got != 313 {
		t.Errorf("ComputeWinRate(1, 31) = %d, want 313", got)
	}
}
