package erc8312

import "testing"

// Golden vector from recompute-kit "8312/cap-conservation": reserved=100,
// confirmed=0, cap=150 → holds.
func TestCheckStatefulBoundGolden(t *testing.T) {
	if !CheckStatefulBound(100, 0, 150) {
		t.Error("CheckStatefulBound(100, 0, 150) = false, want true")
	}
}

// Golden vector: breach — reserved=100, confirmed=60, cap=150 → rejects.
func TestCheckStatefulBoundBreach(t *testing.T) {
	if CheckStatefulBound(100, 60, 150) {
		t.Error("CheckStatefulBound(100, 60, 150) = true, want false")
	}
}

// Exact cap boundary: reserved + confirmed == cap holds.
func TestCheckStatefulBoundExactCap(t *testing.T) {
	if !CheckStatefulBound(100, 50, 150) {
		t.Error("CheckStatefulBound(100, 50, 150) = false, want true (exact cap)")
	}
	if !CheckStatefulBound(0, 0, 0) {
		t.Error("CheckStatefulBound(0, 0, 0) = false, want true")
	}
}

// Overflow safety: naive reserved+confirmed would wrap around; the guarded
// form must reject an overflowing sum against any cap <= MaxUint64.
func TestCheckStatefulBoundOverflow(t *testing.T) {
	if CheckStatefulBound(^uint64(0), 1, ^uint64(0)) {
		t.Error("CheckStatefulBound(MaxUint64, 1, MaxUint64) = true — overflowing sum must not satisfy the cap")
	}
	if !CheckStatefulBound(^uint64(0), 0, ^uint64(0)) {
		t.Error("CheckStatefulBound(MaxUint64, 0, MaxUint64) = false, want true")
	}
}

// Golden vector from recompute-kit "8312/cap-conservation": aggregate=0,
// cap=8000 → holds.
func TestCheckCursorHeadroomGolden(t *testing.T) {
	if !CheckCursorHeadroom(0, 8000) {
		t.Error("CheckCursorHeadroom(0, 8000) = false, want true")
	}
}

// Golden vector: breach — aggregate=8001, cap=8000 → rejects.
func TestCheckCursorHeadroomBreach(t *testing.T) {
	if CheckCursorHeadroom(8001, 8000) {
		t.Error("CheckCursorHeadroom(8001, 8000) = true, want false")
	}
}

func TestCheckCursorHeadroomZeroCap(t *testing.T) {
	if !CheckCursorHeadroom(0, 0) {
		t.Error("CheckCursorHeadroom(0, 0) = false, want true")
	}
}

// Spec-aligned: remaining = cap - spent (IBudgetSubstrate).
func TestComputeRemainingHeadroom(t *testing.T) {
	if got := ComputeRemainingHeadroom(150, 60); got != 90 {
		t.Errorf("ComputeRemainingHeadroom(150, 60) = %d, want 90", got)
	}
	// Exhausted: spent > cap saturates to 0.
	if got := ComputeRemainingHeadroom(150, 200); got != 0 {
		t.Errorf("ComputeRemainingHeadroom(150, 200) = %d, want 0 (saturated)", got)
	}
	if got := ComputeRemainingHeadroom(150, 0); got != 150 {
		t.Errorf("ComputeRemainingHeadroom(150, 0) = %d, want 150", got)
	}
}

// Golden vector "8312/budget-substrate — budget-headroom": cap=150, spent=60,
// reported=90 → holds.
func TestVerifyRemainingHolds(t *testing.T) {
	if !VerifyRemaining(150, 60, 90) {
		t.Error("VerifyRemaining(150, 60, 90) = false, want true")
	}
}

// Golden vector "8312/budget-substrate — budget-substrate-misreport":
// reports 100 but cap - spent = 90 → rejects.
func TestVerifyRemainingMisreport(t *testing.T) {
	if VerifyRemaining(150, 60, 100) {
		t.Error("VerifyRemaining(150, 60, 100) = true, want false")
	}
}

// CRITICAL guard: spent > cap must be rejected even when reported = 0 —
// the saturating subtraction makes cap - spent == 0 when spent > cap, so
// without the spent <= cap guard this would wrongly pass.
func TestVerifyRemainingRejectsSpentOverCap(t *testing.T) {
	if VerifyRemaining(150, 200, 0) {
		t.Error("VerifyRemaining(150, 200, 0) = true, want false (spent > cap must be rejected)")
	}
	// A saturated misreport with a non-zero value must also be rejected.
	if VerifyRemaining(150, 200, 50) {
		t.Error("VerifyRemaining(150, 200, 50) = true, want false (spent > cap)")
	}
}

// Boundary sanity: spent == cap is exactly exhausted (remaining 0), and an
// untouched budget reports the full cap.
func TestVerifyRemainingBoundaries(t *testing.T) {
	if !VerifyRemaining(150, 150, 0) {
		t.Error("VerifyRemaining(150, 150, 0) = false, want true (spent == cap, remaining 0)")
	}
	if !VerifyRemaining(150, 0, 150) {
		t.Error("VerifyRemaining(150, 0, 150) = false, want true (untouched budget)")
	}
}
