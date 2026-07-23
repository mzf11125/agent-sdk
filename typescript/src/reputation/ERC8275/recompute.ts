/**
 * Compute the win rate in basis points from commit-gated wins and losses.
 *
 * ERC-8275: winRate = wins * 10000 / (wins + losses)  (integer division, no float).
 * Convention: exact integer division, half-away-from-zero, never a language float round().
 *
 * @param wins - Number of commit-gated wins (non-negative integer).
 * @param losses - Number of commit-gated losses (non-negative integer).
 * @returns The win rate in basis points (10000 = 1.0, 5161 = 0.5161).
 * @throws If both wins and losses are zero (division by zero).
 */
export function computeWinRate(wins: number, losses: number): number {
  if (wins === 0 && losses === 0) {
    throw new Error(
      'cannot compute win rate: both wins and losses are zero',
    )
  }
  return Math.floor((wins * 10000) / (wins + losses))
}
