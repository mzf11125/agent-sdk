// Package erc8275 implements the ERC-8275 Agent Reputation SDK.
//
// ERC-8275 (Agent Service Discovery and Escrow Payments, Reputation Layer)
// derives reputation entirely from settlement events — derived-not-stored,
// every value is recomputable from public data.
package erc8275

import "errors"

// ErrZeroTotal is returned when a win rate cannot be computed because both
// wins and losses are zero (division by zero).
var ErrZeroTotal = errors.New("cannot compute win rate: both wins and losses are zero")

// ComputeWinRate computes the win rate in basis points from commit-gated
// wins and losses (ERC-8275).
//
// winRate = round_half_up(wins * 10000 / (wins + losses))
// Formula: (2*wins*10000 + total) / (2*total) — exact integer division,
// half-away-from-zero. Never a float round().
//
// Golden vector: wins=16, losses=15 → 5161 (0.5161).
// Rounding-tie: wins=1, losses=31 → 313 (0.0313), matches canonical ROUND_HALF_UP.
//
// Returns ErrZeroTotal if both wins and losses are zero.
func ComputeWinRate(wins, losses uint64) (uint64, error) {
	total := wins + losses
	if total == 0 {
		return 0, ErrZeroTotal
	}
	// (2*wins*10000 + total) / (2*total)
	// = wins*10000/total + 1/2 if the fractional part is exactly 0.5 → half-up
	num := wins * 20000
	return (num + total) / (2 * total), nil
}
