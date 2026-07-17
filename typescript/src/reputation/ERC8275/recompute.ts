/**
 * Compute the win rate from commit-gated wins and losses.
 *
 * ERC-8275: winRate = gated_wins / (gated_wins + gated_losses),
 * rounded to 4 decimal places.
 *
 * @param wins - Number of commit-gated wins (non-negative integer).
 * @param losses - Number of commit-gated losses (non-negative integer).
 * @returns The win rate rounded to 4 decimal places.
 * @throws If both wins and losses are zero (division by zero).
 */
export function computeWinRate(wins: number, losses: number): number {
  if (wins === 0 && losses === 0) {
    throw new Error(
      'cannot compute win rate: both wins and losses are zero',
    )
  }
  const total = wins + losses
  const rate = wins / total
  return Math.round(rate * 10000) / 10000
}
