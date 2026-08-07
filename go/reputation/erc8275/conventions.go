package erc8275

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
)

// governing_convention_hash — pin-at-issuance, resolve-at-verification for
// ERC-8275 winRateBps.
//
// A computation rule (formula + representation + rounding mode) is
// content-addressed:
//
//	governing_convention_hash = "0x" + sha256(JCS(convention_spec))
//
// because number formatting and rounding are exactly where two honest
// implementations diverge without disagreeing on the inputs (1 win / 31
// losses = 312.5 bps -> 313 half-up vs 312 half-even). A producer pins the
// hash at issuance; a verifier resolves it and recomputes under *that*
// convention. Tri-state, fail-closed: an unknown hash is Unverifiable, never
// a silent pass.
//
// Binds the convention this SDK produces — `win_rate.bps.v0`. The spec and
// derived hash are byte-identical to trustless-ai/recompute-kit
// conformance/convention-hash-v0; the hash is DERIVED from the spec
// (reproduce, don't trust) and self-checked against the locked identity in
// tests.

// WinRateBpsV0Spec is the `win_rate.bps.v0` convention spec as sorted
// (key, value) pairs — byte-identical to recompute-kit convention-hash-v0.
// The hash is DERIVED from this, never hardcoded.
var WinRateBpsV0Spec = [][2]string{
	{"erc", "ERC-8275"},
	{"formula", "winRateBps = (gated_wins*20000 + total) // (2*total), total = wins+losses"},
	{"id", "win_rate.bps.v0"},
	{"quantity", "erc8275.win_rate"},
	{"representation", "integer basis points, 0..10000"},
	{"rounding_mode", "round-half-up (half-away-from-zero), exact integer division — never a float round()"},
	{"source", "agent-sdk#5 @87b08f3 reputation/erc8275 — Python/Rust/TS identical; winRateBps live on babyblueviper /ledger"},
}

// WinRateBpsV0Hash is the locked identity (recompute-kit
// convention-hash-v0) — reproduced from the spec, not trusted.
const WinRateBpsV0Hash = "0x0501b75db8e9ef4ef67c74efcfbe2a200b0a7e5aea5ca62f778c91c119e68daf"

// GoverningConventionHash content-addresses a convention spec:
// "0x" + sha256(JCS(spec)). RFC-8785 JCS for a flat string map: sorted keys,
// compact separators, raw UTF-8. Pairs are already in sorted-key order.
func GoverningConventionHash(pairs [][2]string) (string, error) {
	var b strings.Builder
	b.WriteByte('{')
	for i, p := range pairs {
		if i > 0 {
			b.WriteByte(',')
		}
		b.WriteByte('"')
		b.WriteString(p[0])
		b.WriteString(`":"`)
		b.WriteString(p[1])
		b.WriteByte('"')
	}
	b.WriteByte('}')
	sum := sha256.Sum256([]byte(b.String()))
	return "0x" + hex.EncodeToString(sum[:]), nil
}

// PinnedWinRate is the pin-at-issuance record: (winRateBps,
// governing_convention_hash).
type PinnedWinRate struct {
	Value                   uint64
	GoverningConventionHash string
}

// PinWinRateBps pins the win rate at issuance, stamping the convention hash
// that produced it. Returns ErrZeroTotal iff wins == losses == 0.
func PinWinRateBps(wins, losses uint64) (PinnedWinRate, error) {
	value, err := ComputeWinRate(wins, losses)
	if err != nil {
		return PinnedWinRate{}, err
	}
	return PinnedWinRate{Value: value, GoverningConventionHash: WinRateBpsV0Hash}, nil
}

// Verdict is the tri-state result of resolving a pinned value against a
// recompute under a governing convention. Fail-closed: an unknown convention
// hash is Unverifiable, never a silent pass.
type Verdict int

const (
	// Verified means the persisted value equals the recompute under the pinned convention.
	Verified Verdict = iota
	// Rejected means the convention resolves but the value disagrees.
	Rejected
	// Unverifiable means the hash is not one this SDK produces.
	Unverifiable
)

func (v Verdict) String() string {
	switch v {
	case Verified:
		return "verified"
	case Rejected:
		return "rejected"
	case Unverifiable:
		return "unverifiable"
	default:
		return "unknown"
	}
}

// ErrInvalidConventionHash is returned when the governing convention hash is
// not one this SDK produces.
var ErrInvalidConventionHash = errors.New("unverifiable: unknown governing convention hash")

// VerifyWinRate resolves a persisted win rate at verification: verified iff
// the value equals the recompute under the pinned convention; rejected iff
// the convention resolves but disagrees; unverifiable iff the hash is not
// one this SDK produces.
func VerifyWinRate(value uint64, conventionHash string, wins, losses uint64) (Verdict, error) {
	if !strings.EqualFold(conventionHash, WinRateBpsV0Hash) {
		return Unverifiable, ErrInvalidConventionHash
	}
	recomputed, err := ComputeWinRate(wins, losses)
	if err != nil {
		return Rejected, err
	}
	if recomputed == value {
		return Verified, nil
	}
	return Rejected, nil
}
