// Package erc8312 implements the ERC-8312 Bounded Agent Actions SDK.
//
// ERC-8312 meters agent mandates on-chain: an envelope commits to a
// capability root and an opaque consumption cursor that only advanceCursor
// moves. The metering claims are pure integer invariants:
//
//	(reserved + confirmed) <= cap     (StatefulBound variant)
//	aggregate <= cap                  (Orbmis/headroom variant)
//	remaining = cap - spent           (IBudgetSubstrate profile)
//
// All four are pure functions of public integers — no trust required. The
// contracts expose the full state via getEnvelope/bound/spent/remaining, so
// anyone can independently recompute the invariants without trusting the
// party that advanced the cursor.
package erc8312

// CheckStatefulBound checks the StatefulBound variant of ERC-8312:
// (reserved + confirmed) <= cap.
//
// The comparison is computed without overflow: it is equivalent to
// reserved.saturating_add(confirmed) <= cap as in the Rust SDK — a sum that
// would overflow uint64 can never satisfy a cap <= MaxUint64.
//
// Golden vector (recompute-kit "8312/cap-conservation", cross-verified
// against the TypeScript, Python and Rust SDKs): reserved=100, confirmed=0,
// cap=150 → true.
func CheckStatefulBound(reserved, confirmed, cap uint64) bool {
	return cap >= reserved && confirmed <= cap-reserved
}

// CheckCursorHeadroom checks the Orbmis/headroom variant of ERC-8312:
// aggregate <= cap.
//
// Golden vector (recompute-kit "8312/cap-conservation"): aggregate=0,
// cap=8000 → true.
func CheckCursorHeadroom(aggregate, cap uint64) bool {
	return aggregate <= cap
}

// ComputeRemainingHeadroom computes the remaining headroom from cap and
// cumulative spent: remaining = cap - spent (ERC-8312 §IBudgetSubstrate).
// Returns 0 if spent exceeds cap (exhausted or inactive envelope) — the
// saturating subtraction never underflows.
//
// Golden vector (recompute-kit "8312/budget-substrate — budget-headroom"):
// cap=150, spent=60 → 90.
func ComputeRemainingHeadroom(cap, spent uint64) uint64 {
	if spent > cap {
		return 0
	}
	return cap - spent
}

// VerifyRemaining verifies that a reported remaining value matches the
// recomputed headroom: spent <= cap AND (cap - spent) == reported.
//
// CRITICAL: the spent <= cap guard is load-bearing. ComputeRemainingHeadroom
// saturates, so cap - spent == 0 when spent > cap — without the guard,
// VerifyRemaining(cap, spent, 0) would wrongly accept spent > cap. remaining
// is recomputed, never trusted (ERC-8312 §IBudgetSubstrate).
//
// Golden vector (recompute-kit "8312/budget-substrate"): cap=150, spent=60,
// reported=90 → true; reported=100 → false; spent=200 (with reported=0) →
// false.
func VerifyRemaining(cap, spent, reported uint64) bool {
	return spent <= cap && ComputeRemainingHeadroom(cap, spent) == reported
}
