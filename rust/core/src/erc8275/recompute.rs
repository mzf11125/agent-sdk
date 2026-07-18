/// ERC-8275: compute win rate from commit-gated wins and losses.
///
/// winRate = gated_wins / (gated_wins + gated_losses), rounded to 4 decimal places.
///
/// Golden vector: wins=16, losses=15 → 0.5161.
///
/// # Errors
/// Returns `None` if both wins and losses are zero (division by zero).
pub fn compute_win_rate(wins: u64, losses: u64) -> Option<f64> {
    let total = wins.checked_add(losses)?;
    if total == 0 {
        return None;
    }
    let rate = wins as f64 / total as f64;
    Some((rate * 10000.0).round() / 10000.0)
}

#[cfg(test)]
mod tests {
    use super::*;

    /// Golden vector: wins=16, losses=15 → 0.5161
    #[test]
    fn golden_win_rate() {
        let rate = compute_win_rate(16, 15).unwrap();
        assert!((rate - 0.5161).abs() < 0.0001);
    }

    #[test]
    fn zero_wins_is_zero_win_rate() {
        let rate = compute_win_rate(0, 10).unwrap();
        assert_eq!(rate, 0.0);
    }

    #[test]
    fn all_wins_is_one() {
        let rate = compute_win_rate(10, 0).unwrap();
        assert_eq!(rate, 1.0);
    }

    #[test]
    fn both_zero_returns_none() {
        assert!(compute_win_rate(0, 0).is_none());
    }

    #[test]
    fn rounding_to_4_decimals() {
        // 1/3 = 0.3333... → 0.3333
        let rate = compute_win_rate(1, 2).unwrap();
        assert!((rate - 0.3333).abs() < 0.0001);
    }
}
