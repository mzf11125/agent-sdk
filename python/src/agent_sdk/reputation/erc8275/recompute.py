def compute_win_rate(wins: int, losses: int) -> int:
    """
    Compute the win rate in basis points from commit-gated wins and losses.

    ERC-8275: winRate = round_half_up(wins * 10000 / (wins + losses))
    Formula: (2*wins*10000 + total) // (2*total) — exact integer division,
    half-away-from-zero. Never a float round().

    Args:
        wins: Number of commit-gated wins (non-negative integer).
        losses: Number of commit-gated losses (non-negative integer).

    Returns:
        The win rate in basis points (10000 = 1.0, 5161 = 0.5161).

    Raises:
        ValueError: If both wins and losses are zero (division by zero).
    """
    total = wins + losses
    if total == 0:
        raise ValueError(
            "cannot compute win rate: both wins and losses are zero"
        )
    return (wins * 20000 + total) // (2 * total)
