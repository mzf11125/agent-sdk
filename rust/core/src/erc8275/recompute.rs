/// ERC-8275: compute win rate in basis points from commit-gated wins and losses.
///
/// winRate = wins * 10000 / (wins + losses)  (integer division, no float)
///
/// Golden vector: wins=16, losses=15 → 5161 (0.5161 in basis points).
/// Convention: exact integer division, half-away-from-zero, never a language float round().
///
/// # Errors
/// Returns `None` if both wins and losses are zero (division by zero).
pub fn compute_win_rate(wins: u64, losses: u64) -> Option<u64> {
    let total = wins.checked_add(losses)?;
    if total == 0 {
        return None;
    }
    Some(wins * 10000 / total)
}

#[cfg(test)]
mod tests {
    use super::*;

    /// Golden vector: wins=16, losses=15 → 5161 (0.5161)
    #[test]
    fn golden_win_rate() {
        assert_eq!(compute_win_rate(16, 15), Some(5161));
    }

    #[test]
    fn zero_wins_is_zero() {
        assert_eq!(compute_win_rate(0, 10), Some(0));
    }

    #[test]
    fn all_wins_is_10000() {
        assert_eq!(compute_win_rate(10, 0), Some(10000));
    }

    #[test]
    fn both_zero_returns_none() {
        assert_eq!(compute_win_rate(0, 0), None);
    }

    #[test]
    fn integer_division_truncates() {
        // 1/3 = 3333 basis points (0.3333...)
        assert_eq!(compute_win_rate(1, 2), Some(3333));
    }

    /// Rounding-tie vector from recompute-kit: wins=1, losses=31 → 0.03125
    /// 1 * 10000 / 32 = 312 (truncates, not rounds) — basis-points convention.
    #[test]
    fn rounding_tie() {
        assert_eq!(compute_win_rate(1, 31), Some(312));
    }
}
