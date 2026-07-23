def compute_win_rate(wins: int, losses: int) -> int:
    """
    Compute the win rate in basis points from commit-gated wins and losses.

    ERC-8275: winRate = wins * 10000 / (wins + losses)  (integer division, no float).
    Convention: exact integer division, half-away-from-zero, never a language float round().

    Args:
        wins: Number of commit-gated wins (non-negative integer).
        losses: Number of commit-gated losses (non-negative integer).

    Returns:
        The win rate in basis points (10000 = 1.0, 5161 = 0.5161).

    Raises:
        ValueError: If both wins and losses are zero (division by zero).
    """
    if wins == 0 and losses == 0:
        raise ValueError(
            "cannot compute win rate: both wins and losses are zero"
        )
    return (wins * 10000) // (wins + losses)
