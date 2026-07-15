def compute_win_rate(wins: int, losses: int) -> float:
    """
    Compute the win rate from commit-gated wins and losses.

    ERC-8275: winRate = gated_wins / (gated_wins + gated_losses),
    rounded to 4 decimal places.

    Args:
        wins: Number of commit-gated wins (non-negative integer).
        losses: Number of commit-gated losses (non-negative integer).

    Returns:
        The win rate rounded to 4 decimal places.

    Raises:
        ValueError: If both wins and losses are zero (division by zero).
    """
    if wins == 0 and losses == 0:
        raise ValueError(
            "cannot compute win rate: both wins and losses are zero"
        )
    return round(wins / (wins + losses), 4)
